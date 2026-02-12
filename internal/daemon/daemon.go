package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dopejs/opencc/internal/config"
)

// PidPath returns the path to the PID file.
func PidPath() string {
	return filepath.Join(config.ConfigDirPath(), config.WebPidFile)
}

// LogPath returns the path to the web log file.
func LogPath() string {
	return filepath.Join(config.ConfigDirPath(), config.WebLogFile)
}

// WritePid writes the given PID to the PID file atomically with 0600 permissions.
func WritePid(pid int) error {
	dir := config.ConfigDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := PidPath() + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)+"\n"), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, PidPath())
}

// ReadPid reads the PID from the PID file.
func ReadPid() (int, error) {
	data, err := os.ReadFile(PidPath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file: %w", err)
	}
	return pid, nil
}

// RemovePid removes the PID file.
func RemovePid() {
	os.Remove(PidPath())
}
