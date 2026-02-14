package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/internal/daemon"
	"github.com/dopejs/opencc/internal/proxy"
	"github.com/dopejs/opencc/internal/web"
	"github.com/spf13/cobra"
)

var webDaemonFlag bool
var webPortOverride int

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web configuration interface",
	Long:  "Start an embedded HTTP server on 127.0.0.1:19840 for managing providers and profiles.",
	RunE:  runWeb,
}

var webStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the web daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, running := daemon.IsRunning(); !running {
			fmt.Println("Web server is not running.")
			return nil
		}
		if err := daemon.StopDaemon(); err != nil {
			return err
		}
		fmt.Println("Web server stopped.")
		return nil
	},
}

var webStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show web daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		pid, running := daemon.IsRunning()
		if running {
			fmt.Printf("Web server is running (PID %d) on http://127.0.0.1:%d\n", pid, config.GetWebPort())
		} else {
			fmt.Println("Web server is not running.")
		}
	},
}

var webEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Install as a system service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.EnableService(); err != nil {
			return err
		}
		fmt.Println("Web server installed as system service.")
		return nil
	},
}

var webRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the web daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, running := daemon.IsRunning(); !running {
			fmt.Println("Web server is not running.")
			return nil
		}
		return restartWebDaemon()
	},
}

var webDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Uninstall system service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.DisableService(); err != nil {
			return err
		}
		fmt.Println("Web server system service removed.")
		return nil
	},
}

func init() {
	webCmd.Flags().BoolVarP(&webDaemonFlag, "daemon", "d", false, "run in background daemon mode")
	webCmd.Flags().IntVar(&webPortOverride, "port", 0, "")
	webCmd.Flags().MarkHidden("port")
	webCmd.AddCommand(webStopCmd)
	webCmd.AddCommand(webStatusCmd)
	webCmd.AddCommand(webRestartCmd)
	webCmd.AddCommand(webEnableCmd)
	webCmd.AddCommand(webDisableCmd)
}

func runWeb(cmd *cobra.Command, args []string) error {
	// If this is the daemon child process, run the server directly (no port override).
	if os.Getenv("OPENCC_WEB_DAEMON") == "1" {
		return runWebServer(0)
	}

	// Daemon mode: ignore --port, use configured port.
	if webDaemonFlag {
		return startDaemon()
	}

	// Foreground mode: --port override only applies here.
	return runWebForeground(webPortOverride)
}

func runWebServer(portOverride int) error {
	logFile, logger := setupWebLogger()
	if logFile != nil {
		defer logFile.Close()
	}

	// Initialize global structured logger for proxy
	if err := proxy.InitGlobalLogger(config.ConfigDirPath()); err != nil {
		logger.Printf("Warning: failed to initialize structured logger: %v", err)
	}

	srv := web.NewServer(Version, logger, portOverride)

	// Only manage PID file when running on the configured port (normal mode).
	managePid := portOverride == 0
	if managePid {
		daemon.WritePid(os.Getpid())
	}

	// Graceful shutdown on signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Println("Shutting down web server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		if managePid {
			daemon.RemovePid()
		}
	}()

	return srv.Start()
}

func runWebForeground(portOverride int) error {
	// When using an alternate port, skip PID check — this is a temporary test instance.
	if portOverride == 0 {
		if pid, running := daemon.IsRunning(); running {
			fmt.Printf("Web server already running (PID %d). Opening browser...\n", pid)
			openBrowser(fmt.Sprintf("http://127.0.0.1:%d", config.GetWebPort()))
			return nil
		}
	}

	port := config.GetWebPort()
	if portOverride > 0 {
		port = portOverride
	}
	fmt.Printf("Starting web server on http://127.0.0.1:%d\n", port)

	// Open browser after a short delay to let server start.
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))
	}()

	return runWebServer(portOverride)
}

func startDaemon() error {
	// Check if already running.
	if pid, running := daemon.IsRunning(); running {
		fmt.Printf("Web server already running (PID %d).\n", pid)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	logPath := daemon.LogPath()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}
	defer logFile.Close()

	child := exec.Command(exe, "web")
	child.Env = append(os.Environ(), "OPENCC_WEB_DAEMON=1")
	child.Stdout = logFile
	child.Stderr = logFile
	child.SysProcAttr = daemon.DaemonSysProcAttr()

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	daemon.WritePid(child.Process.Pid)

	// Wait for the server to be ready.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := web.WaitForReady(ctx, 0); err != nil {
		return fmt.Errorf("daemon started but server did not become ready: %w", err)
	}

	fmt.Printf("Web server started in background (PID %d) on http://127.0.0.1:%d\n", child.Process.Pid, config.GetWebPort())
	return nil
}

// restartWebDaemon stops and restarts the web daemon with a spinner display.
func restartWebDaemon() error {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r  %s Restarting web server...", frames[i%len(frames)])
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	var restartErr error
	if err := daemon.StopDaemon(); err != nil {
		restartErr = fmt.Errorf("failed to stop web daemon: %w", err)
	} else {
		// Brief pause to let the port be released
		time.Sleep(200 * time.Millisecond)
		if err := startDaemon(); err != nil {
			restartErr = fmt.Errorf("failed to start web daemon: %w", err)
		}
	}

	close(done)
	if restartErr != nil {
		fmt.Fprintf(os.Stderr, "\r  Web server restart failed: %v\n", restartErr)
		return restartErr
	}
	fmt.Fprintf(os.Stderr, "\r  Web server restarted.                \n")
	return nil
}

func setupWebLogger() (*os.File, *log.Logger) {
	logDir := config.ConfigDirPath()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(daemon.LogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, log.New(os.Stderr, "[web] ", log.LstdFlags)
	}
	return logFile, log.New(logFile, "[web] ", log.LstdFlags)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	}
}
