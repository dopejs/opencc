//go:build !windows

package daemon

import (
	"fmt"
	"os"
	"syscall"
)

// IsRunning checks if the daemon process is still alive.
func IsRunning() (int, bool) {
	pid, err := ReadPid()
	if err != nil {
		return 0, false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	// Signal 0 checks if process exists without actually signaling it.
	err = proc.Signal(syscall.Signal(0))
	return pid, err == nil
}

// StopDaemon sends SIGTERM to the daemon process.
func StopDaemon() error {
	pid, running := IsRunning()
	if !running {
		RemovePid()
		return fmt.Errorf("daemon is not running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon (PID %d): %w", pid, err)
	}
	RemovePid()
	return nil
}

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Unix.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
