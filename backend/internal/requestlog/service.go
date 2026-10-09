package requestlog

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
)

// Repository 保存请求摘要，并按调用者权限查询现存记录。
type Repository interface {
	Save(context.Context, []telemetry.RequestRecord) error
	Find(context.Context, string, int64, bool) ([]Detail, error)
	Cleanup(context.Context, time.Time) error
}

// Queue 在写入数据库前持久保存快照，确认时按版本移除。
type Queue interface {
	Put(telemetry.RequestRecord) error
	Peek(int) ([]telemetry.RequestRecord, error)
	Ack([]telemetry.RequestRecord) error
	Pending() (int, error)
}

// Usage 是请求详情中的费用和用量摘要。
type Usage struct {
	ID           int64   `json:"id"`
	Model        string  `json:"model"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	ActualCost   float64 `json:"actual_cost"`
}

// Failure 是经过归属检查的错误引用。
type Failure struct {
	ID     int64  `json:"id"`
	Status int    `json:"status_code"`
	Phase  string `json:"phase"`
}

// Detail 汇总请求摘要及仍在保留期内的业务记录。
type Detail struct {
	telemetry.RequestRecord
	Usage    []Usage   `json:"usage"`
	Errors   []Failure `json:"errors"`
	AuditIDs []int64   `json:"audit_ids,omitempty"`
	Legacy   bool      `json:"legacy"`
	Pending  bool      `json:"pending,omitempty"`
}

// Health 表示待写记录和累计写入故障。
type Health struct {
	Pending  int    `json:"pending"`
	Failures uint64 `json:"failures"`
}

// Service 在请求结束后批量写库，待写记录由 Queue 跨进程重启保存。
type Service struct {
	repo      Repository
	queue     Queue
	retention time.Duration
	logf      func(string, ...any)
	failures  atomic.Uint64
	flushMu   sync.Mutex
	startOnce sync.Once
	stopOnce  sync.Once
	stop      chan struct{}
	done      chan struct{}
}

// NewService 配置请求摘要的批写和留存。
// @project-doc docs/operations/request_lookup.md#request_storage
func NewService(repo Repository, queue Queue, retentionDays int, logf func(string, ...any)) *Service {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &Service{repo: repo, queue: queue, retention: time.Duration(retentionDays) * 24 * time.Hour, logf: logf, stop: make(chan struct{}), done: make(chan struct{})}
}

// Observe 先持久化快照，数据库批写由后台任务执行。
func (s *Service) Observe(record telemetry.RequestRecord) {
	if s == nil || s.queue == nil {
		return
	}
	if record.RequestID == "" {
		s.report("persist request record", fmt.Errorf("missing request ID"))
		return
	}
	if err := s.queue.Put(telemetry.NormalizeRequestRecord(record)); err != nil {
		s.report("persist request "+record.RequestID, err)
	}
}

func (s *Service) Start(context.Context) error {
	s.startOnce.Do(func() { go s.run() })
	return nil
}

func (s *Service) run() {
	defer close(s.done)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	cleanup := time.NewTicker(time.Hour)
	defer cleanup.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.Flush(ctx); err != nil {
				s.report("flush request records", err)
			}
			cancel()
		case <-cleanup.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.repo.Cleanup(ctx, time.Now().Add(-s.retention)); err != nil {
				s.report("clean request records", err)
			}
			cancel()
		}
	}
}

func (s *Service) Stop(ctx context.Context) error {
	defer func() {
		if queue, ok := s.queue.(interface{ Close() error }); ok {
			_ = queue.Close()
		}
	}()
	s.stopOnce.Do(func() { close(s.stop) })
	s.startOnce.Do(func() { close(s.done) })
	select {
	case <-s.done:
	case <-ctx.Done():
		return ctx.Err()
	}
	for {
		pending, err := s.queue.Pending()
		if err != nil || pending == 0 {
			return err
		}
		if err = s.Flush(ctx); err != nil {
			return err
		}
	}
}

// Flush 在一次数据库事务成功后确认队列中的对应版本。
func (s *Service) Flush(ctx context.Context) error {
	s.flushMu.Lock()
	defer s.flushMu.Unlock()
	for range 8 {
		if err := ctx.Err(); err != nil {
			return err
		}
		batch, err := s.queue.Peek(128)
		if err != nil || len(batch) == 0 {
			return err
		}
		if err = s.repo.Save(ctx, batch); err != nil {
			return err
		}
		if err = s.queue.Ack(batch); err != nil {
			return err
		}
		if len(batch) < 128 {
			return nil
		}
	}
	return nil
}

func (s *Service) Find(ctx context.Context, id string, userID int64, admin bool) ([]Detail, error) {
	items, err := s.repo.Find(ctx, id, userID, admin)
	if pending, ok := s.queue.(interface {
		Get(string) (telemetry.RequestRecord, bool)
	}); ok {
		if record, found := pending.Get(id); found && (admin || record.UserID > 0 && record.UserID == userID) {
			for i := range items {
				if items[i].RequestID == id {
					items[i].RequestRecord = telemetry.MergeRequestRecord(items[i].RequestRecord, record)
					items[i].Pending = true
					return items, nil
				}
			}
			return append(items, Detail{RequestRecord: record, Pending: true}), nil
		}
	}
	return items, err
}

func (s *Service) Health() (Health, error) {
	pending, err := s.queue.Pending()
	return Health{Pending: pending, Failures: s.failures.Load()}, err
}

func (s *Service) report(action string, err error) {
	s.failures.Add(1)
	if s.logf != nil {
		s.logf("%s: %v", action, fmt.Errorf("request records: %w", err))
	}
}
