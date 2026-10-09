package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
)

const (
	queuePendingBytesLimit = 64 << 20
	queueCheckpointBytes   = 8 << 20
)

// Queue 使用追加日志和批量 fsync 保存待写快照，数据库确认后回收内存。
// 检查点压缩已确认的日志，进程锁保证同一目录只有一个写入者。
type Queue struct {
	mu           sync.Mutex
	submitMu     sync.RWMutex
	dir          string
	lock         *os.File
	file         *os.File
	pending      map[string]telemetry.RequestRecord
	sizes        map[string]int
	pendingBytes int
	journalBytes int64
	compactAt    int64
	commands     chan *queueCommand
	done         chan struct{}
	closed       bool
	closeErr     error
}

// queueCommand 的完成通知表示日志已同步到磁盘。
type queueCommand struct {
	record *telemetry.RequestRecord
	ack    []queueAck
	done   chan error
	err    error
}

type queueAck struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type queueFrame struct {
	Record *telemetry.RequestRecord `json:"record,omitempty"`
	Ack    []queueAck               `json:"ack,omitempty"`
}

func NewQueue(dir string) (*Queue, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	lock, err := lockQueue(filepath.Join(dir, ".lock"))
	if err != nil {
		return nil, fmt.Errorf("request queue directory is already in use or unavailable: %w", err)
	}
	q := &Queue{dir: dir, lock: lock, pending: make(map[string]telemetry.RequestRecord), sizes: make(map[string]int), compactAt: queueCheckpointBytes, commands: make(chan *queueCommand, 1024), done: make(chan struct{})}
	q.file, err = os.OpenFile(filepath.Join(dir, "requests.journal"), os.O_CREATE|os.O_RDWR, 0o600)
	if err == nil {
		err = q.recover()
	}
	if err != nil {
		if q.file != nil {
			_ = q.file.Close()
		}
		_ = lock.Close()
		return nil, err
	}
	go q.run()
	return q, nil
}

// recover 读取完整日志帧，并移除进程中断时未写完的末帧。
func (q *Queue) recover() error {
	reader := bufio.NewReader(q.file)
	var offset int64
	for {
		line, err := reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		var frame queueFrame
		if err := json.Unmarshal(line, &frame); err != nil {
			return fmt.Errorf("corrupt request journal at %d: %w", offset, err)
		}
		q.apply(frame)
		offset += int64(len(line))
	}
	if err := q.file.Truncate(offset); err != nil {
		return err
	}
	if _, err := q.file.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	q.journalBytes = offset
	return q.file.Sync()
}

// Put 等待同批次的追加日志完成同步，再把控制权交回请求处理器。
func (q *Queue) Put(record telemetry.RequestRecord) error {
	record = telemetry.NormalizeRequestRecord(record)
	return q.submit(&queueCommand{record: &record, done: make(chan error, 1)})
}

func (q *Queue) submit(command *queueCommand) error {
	q.submitMu.RLock()
	if q.closed {
		q.submitMu.RUnlock()
		return fs.ErrClosed
	}
	q.commands <- command
	q.submitMu.RUnlock()
	return <-command.done
}

// run 将短窗口内的更新合并为一次磁盘同步。
func (q *Queue) run() {
	defer close(q.done)
	for first := range q.commands {
		batch := []*queueCommand{first}
		timer := time.NewTimer(time.Millisecond)
	collect:
		for len(batch) < 256 {
			select {
			case command, ok := <-q.commands:
				if !ok {
					break collect
				}
				batch = append(batch, command)
			case <-timer.C:
				break collect
			}
		}
		timer.Stop()
		err := q.writeBatch(batch)
		for _, command := range batch {
			command.done <- errors.Join(err, command.err)
		}
	}
	// 已确认记录可以在关闭时从日志移除，未确认记录继续等待重启恢复。
	q.closeErr = q.checkpoint()
	if q.file != nil {
		q.closeErr = errors.Join(q.closeErr, q.file.Close())
	}
	q.closeErr = errors.Join(q.closeErr, q.lock.Close())
}

func (q *Queue) writeBatch(batch []*queueCommand) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	var output bytes.Buffer
	frames := make([]queueFrame, 0, len(batch))
	projected := q.pendingBytes
	for _, command := range batch {
		frame := queueFrame{Record: command.record, Ack: command.ack}
		if frame.Record != nil {
			// 同一批次中的后续帧在 apply 时按版本合并，持久日志保留原顺序。
			if previous, ok := q.pending[frame.Record.RequestID]; ok {
				merged := telemetry.MergeRequestRecord(previous, *frame.Record)
				frame.Record = &merged
			}
		}
		data, err := json.Marshal(frame)
		if err != nil {
			return err
		}
		if frame.Record != nil {
			next := projected + len(data) - q.sizes[frame.Record.RequestID]
			if next > queuePendingBytesLimit {
				command.err = fmt.Errorf("request queue exceeds %d bytes", queuePendingBytesLimit)
				continue
			}
			projected = next
		}
		_, _ = output.Write(data)
		_ = output.WriteByte('\n')
		frames = append(frames, frame)
	}
	if output.Len() == 0 {
		return nil
	}
	if err := q.openJournal(); err != nil {
		return err
	}
	before := q.journalBytes
	if _, err := q.file.Write(output.Bytes()); err != nil {
		return q.rollbackAppend(before, err)
	}
	if err := q.file.Sync(); err != nil {
		return q.rollbackAppend(before, err)
	}
	q.journalBytes += int64(output.Len())
	for _, frame := range frames {
		q.apply(frame)
	}
	if q.journalBytes >= q.compactAt {
		return q.checkpoint()
	}
	return nil
}

func (q *Queue) rollbackAppend(offset int64, cause error) error {
	truncateErr := q.file.Truncate(offset)
	_, seekErr := q.file.Seek(offset, io.SeekStart)
	return errors.Join(cause, truncateErr, seekErr)
}

// apply 使确认操作只移除数据库实际收到的版本。
func (q *Queue) apply(frame queueFrame) {
	if frame.Record != nil {
		record := *frame.Record
		if previous, ok := q.pending[record.RequestID]; ok {
			record = telemetry.MergeRequestRecord(previous, record)
		}
		data, _ := json.Marshal(queueFrame{Record: &record})
		size := len(data) + 1
		q.pendingBytes += size - q.sizes[record.RequestID]
		q.pending[record.RequestID] = record
		q.sizes[record.RequestID] = size
	}
	for _, ack := range frame.Ack {
		if current, ok := q.pending[ack.ID]; ok && current.UpdatedAt.Equal(ack.UpdatedAt) {
			delete(q.pending, ack.ID)
			q.pendingBytes -= q.sizes[ack.ID]
			delete(q.sizes, ack.ID)
		}
	}
}

// checkpoint 原子替换为仍未确认的快照；追加写入由同一个协程串行执行。
func (q *Queue) checkpoint() error {
	if err := q.openJournal(); err != nil {
		return err
	}
	file, err := os.CreateTemp(q.dir, ".checkpoint-")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	encoder := json.NewEncoder(file)
	for _, record := range q.pending {
		if err = encoder.Encode(queueFrame{Record: &record}); err != nil {
			_ = file.Close()
			return err
		}
	}
	position, err := file.Seek(0, io.SeekCurrent)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		return errors.Join(err, closeErr)
	}
	err = q.file.Close()
	q.file = nil
	if err != nil {
		return err
	}
	path := filepath.Join(q.dir, "requests.journal")
	replaceErr := replaceQueueFile(file.Name(), path)
	q.file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if replaceErr != nil {
		return replaceErr
	}
	if err = syncQueueDirectory(q.dir); err != nil {
		return err
	}
	q.journalBytes = position
	q.compactAt = max(queueCheckpointBytes, 2*position)
	return nil
}

func (q *Queue) Peek(limit int) ([]telemetry.RequestRecord, error) {
	if limit <= 0 {
		return nil, nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	batch := make([]telemetry.RequestRecord, 0, min(limit, len(q.pending)))
	for _, record := range q.pending {
		batch = append(batch, telemetry.CloneRequestRecord(record))
		if len(batch) == limit {
			break
		}
	}
	return batch, nil
}

func (q *Queue) Ack(batch []telemetry.RequestRecord) error {
	ack := make([]queueAck, 0, len(batch))
	for _, record := range batch {
		ack = append(ack, queueAck{ID: record.RequestID, UpdatedAt: record.UpdatedAt})
	}
	return q.submit(&queueCommand{ack: ack, done: make(chan error, 1)})
}

func (q *Queue) Pending() (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending), nil
}

func (q *Queue) Get(id string) (telemetry.RequestRecord, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	record, ok := q.pending[id]
	return telemetry.CloneRequestRecord(record), ok
}

// Close 等待已经提交的日志落盘，并释放进程锁。
func (q *Queue) Close() error {
	q.submitMu.Lock()
	if !q.closed {
		q.closed = true
		close(q.commands)
	}
	q.submitMu.Unlock()
	<-q.done
	return q.closeErr
}

// openJournal 在检查点文件切换失败后重新打开最后保留下来的日志。
func (q *Queue) openJournal() error {
	if q.file != nil {
		return nil
	}
	file, err := os.OpenFile(filepath.Join(q.dir, "requests.journal"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	q.file = file
	q.journalBytes = info.Size()
	return nil
}
