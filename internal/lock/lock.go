// internal/lock/lock.go

package lock

import (
	"fmt"
	"os"
	"syscall"
)

type Locker struct {
	path string
}

func New(path string) *Locker {
	return &Locker{path: path}
}

func (l *Locker) Acquire() (func(), error) {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, fmt.Errorf("another gitback process already running")
	}

	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}
