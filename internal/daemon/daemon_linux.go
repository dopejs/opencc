package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func systemdUnitPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", "opencc-web.service")
}

const unitTemplate = `[Unit]
Description=OpenCC Web Config Server
After=network.target

[Service]
Type=simple
ExecStart={{.Executable}} web
Environment=OPENCC_WEB_DAEMON=1
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

// EnableService installs and enables the systemd user unit on Linux.
func EnableService() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	unitPath := systemdUnitPath()
	dir := filepath.Dir(unitPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(unitPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("unit").Parse(unitTemplate))
	if err := tmpl.Execute(f, struct {
		Executable string
	}{
		Executable: exe,
	}); err != nil {
		return err
	}

	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload failed: %s: %w", string(out), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", "opencc-web.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable failed: %s: %w", string(out), err)
	}

	return nil
}

// DisableService disables and removes the systemd user unit on Linux.
func DisableService() error {
	unitPath := systemdUnitPath()

	exec.Command("systemctl", "--user", "stop", "opencc-web.service").Run()
	exec.Command("systemctl", "--user", "disable", "opencc-web.service").Run()

	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()

	return nil
}
