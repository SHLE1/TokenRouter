package provider

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockQueue 通过文件锁保护待写目录，文件关闭后释放。
func lockQueue(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	var overlapped windows.Overlapped
	if err = windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

// replaceQueueFile 使用 Windows 的写透重命名提交目录项。
func replaceQueueFile(source, destination string) error {
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

// syncQueueDirectory 的目录同步由 replaceQueueFile 的写透重命名完成。
func syncQueueDirectory(string) error { return nil }
