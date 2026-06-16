//go:build unix

package importio

import (
	"fmt"
	"os"
	"syscall"
)

func flockShared(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("importio: flock shared: %w", err)
	}
	return nil
}

func funlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
