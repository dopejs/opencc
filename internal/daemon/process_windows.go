package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// IsRunning checks if the daemon process is still alive on Windows.
func IsRunning() (int, bool) {
	pid, err := ReadPid()
	if err != nil {
		return 0, false
	}
	// On Windows, FindProcess always succeeds. Use tasklist to verify.
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		return 0, false
	}
	// tasklist output contains the PID if the process exists.
	if len(out) > 0 && fmt.Sprintf("%d", pid) != "" {
		// Check if output contains the PID number
		if contains(string(out), fmt.Sprintf(" %d ", pid)) {
			return pid, true
		}
	}
	return 0, false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// StopDaemon terminates the daemon process on Windows.
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
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to stop daemon (PID %d): %w", pid, err)
	}
	RemovePid()
	return nil
}

const _CREATE_NEW_PROCESS_GROUP = 0x00000200

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Windows.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: _CREATE_NEW_PROCESS_GROUP}
}
