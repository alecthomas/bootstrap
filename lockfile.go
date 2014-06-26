package util

import (
	"os"
	"syscall"
)

// LockFile represents an exclusive, advisory OS file lock
type LockFile struct {
	file *os.File
}

// AcquireLock locks a file with an exclusive, advisory lock
func AcquireLock(file *os.File) (*LockFile, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return nil, err
	}
	return &LockFile{file}, nil
}

// Release releases the OS file lock and closes the respective file
func (lf *LockFile) Release() error {
	if err := syscall.Flock(int(lf.file.Fd()), syscall.LOCK_UN); err != nil {
		lf.file.Close()
		lf.file = nil
		return err
	}
	if err := lf.file.Close(); err != nil {
		lf.file = nil
		return err
	}
	return nil
}
