//go:build !windows

package provider

import (
	"os"
	"syscall"
)

// lockQueue 通过进程锁保护共享目录，进程退出后由操作系统释放。
func lockQueue(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

// replaceQueueFile 在同一目录中原子替换快照。
func replaceQueueFile(source, destination string) error { return os.Rename(source, destination) }

// syncQueueDirectory 在数据同步后持久化目录项。
func syncQueueDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	return dir.Sync()
}
