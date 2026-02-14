package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/dopejs/opencc/internal/daemon"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [version]",
	Short: "Upgrade opencc to latest or specified version",
	Long: `Upgrade opencc to the latest version, or a specific version.

Examples:
  opencc upgrade          # latest version
  opencc upgrade 1        # latest 1.x.x
  opencc upgrade 1.2      # latest 1.2.x
  opencc upgrade 1.2.3    # exact version 1.2.3`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpgrade,
}

const repoAPI = "https://api.github.com/repos/dopejs/opencc/releases"

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}

	fmt.Println("Fetching releases...")
	target, err := resolveVersion(prefix)
	if err != nil {
		return err
	}

	current := Version
	if target == current {
		fmt.Printf("Already at version %s\n", current)
		return nil
	}

	fmt.Printf("Upgrading: %s → %s\n", current, target)

	// Determine download format based on version
	// v1.4.0+ uses tar.gz, earlier versions use raw binary
	useTarball := shouldUseTarball(target)

	var downloadURL string
	if useTarball {
		assetName := fmt.Sprintf("opencc-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		downloadURL = fmt.Sprintf("https://github.com/dopejs/opencc/releases/download/v%s/%s", target, assetName)
	} else {
		assetName := fmt.Sprintf("opencc-%s-%s", runtime.GOOS, runtime.GOARCH)
		downloadURL = fmt.Sprintf("https://github.com/dopejs/opencc/releases/download/v%s/%s", target, assetName)
	}

	fmt.Printf("Downloading %s...\n", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d (binary for %s/%s may not exist)", resp.StatusCode, runtime.GOOS, runtime.GOARCH)
	}

	tmpFile, err := os.CreateTemp("", "opencc-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, &progressReader{
		reader: resp.Body,
		total:  resp.ContentLength,
	}); err != nil {
		tmpFile.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Println()
	tmpFile.Close()

	// Extract binary path
	var binaryPath string
	if useTarball {
		// Extract from tar.gz
		extractedPath, err := extractTarGz(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}
		defer os.Remove(extractedPath)
		binaryPath = extractedPath
	} else {
		binaryPath = tmpFile.Name()
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	// Replace current binary
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	// Try direct copy first, fall back to sudo
	usedSudo := false
	if err := copyFile(binaryPath, binPath); err != nil {
		fmt.Println("Need elevated privileges, trying sudo...")
		if sudoErr := exec.Command("sudo", "cp", binaryPath, binPath).Run(); sudoErr != nil {
			return fmt.Errorf("install failed: %w", sudoErr)
		}
		usedSudo = true
	}

	// On macOS, re-sign the binary to clear com.apple.provenance
	// so Gatekeeper won't kill the downloaded binary
	if runtime.GOOS == "darwin" {
		codesignArgs := []string{"codesign", "--force", "--sign", "-", binPath}
		if usedSudo {
			codesignArgs = append([]string{"sudo"}, codesignArgs...)
		}
		if err := exec.Command(codesignArgs[0], codesignArgs[1:]...).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: ad-hoc codesign failed: %v\n", err)
		}
	}

	fmt.Printf("Successfully upgraded to %s\n", target)

	// Restart web daemon if it was running
	if pid, running := daemon.IsRunning(); running {
		fmt.Printf("Web daemon is running (PID %d), restarting...\n", pid)
		if err := daemon.StopDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to stop web daemon: %v\n", err)
		} else if err := startDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to restart web daemon: %v\n", err)
		} else {
			fmt.Println("Web daemon restarted.")
		}
	}

	return nil
}

// shouldUseTarball returns true if the version should use tar.gz format.
// v1.4.0+ uses tar.gz, earlier versions use raw binary.
func shouldUseTarball(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	return major > 1 || (major == 1 && minor >= 4)
}

// extractTarGz extracts the opencc binary from a tar.gz file.
func extractTarGz(tarPath string) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the opencc binary
		if header.Typeflag == tar.TypeReg && (header.Name == "opencc" || strings.HasSuffix(header.Name, "/opencc")) {
			tmpFile, err := os.CreateTemp("", "opencc-extracted-*")
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return "", err
			}
			tmpFile.Close()
			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("opencc binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// progressReader wraps an io.Reader and prints a download progress bar.
type progressReader struct {
	reader  io.Reader
	total   int64 // -1 if unknown
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	if pr.total > 0 {
		pct := float64(pr.current) / float64(pr.total) * 100
		barWidth := 30
		filled := int(float64(barWidth) * float64(pr.current) / float64(pr.total))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		status := fmt.Sprintf("\r  %s %5.1f%% %s/%s", bar, pct, formatBytes(pr.current), formatBytes(pr.total))
		// Pad with spaces to clear any leftover characters from previous shorter output
		fmt.Fprintf(os.Stderr, "%-60s", status)
	} else {
		fmt.Fprintf(os.Stderr, "\r  %s downloaded", formatBytes(pr.current))
	}

	return n, err
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// resolveVersion finds the best matching version for the given prefix.
// "" → latest, "1" → latest 1.x.x, "1.2" → latest 1.2.x, "1.2.3" → exact
func resolveVersion(prefix string) (string, error) {
	resp, err := http.Get(repoAPI)
	if err != nil {
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var releases []ghRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to parse releases: %w", err)
	}

	var versions []string
	for _, r := range releases {
		v := strings.TrimPrefix(r.TagName, "v")
		if v != "" {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	if prefix == "" {
		sortVersions(versions)
		return versions[len(versions)-1], nil
	}

	// Filter by prefix match
	prefix = strings.TrimPrefix(prefix, "v")
	var matched []string
	for _, v := range versions {
		if matchVersionPrefix(v, prefix) {
			matched = append(matched, v)
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("no release matching %q", prefix)
	}

	sortVersions(matched)
	return matched[len(matched)-1], nil
}

// matchVersionPrefix checks if version matches the given prefix.
// "1" matches "1.x.x", "1.2" matches "1.2.x", "1.2.3" matches exactly.
func matchVersionPrefix(version, prefix string) bool {
	vParts := strings.Split(version, ".")
	pParts := strings.Split(prefix, ".")

	for i, p := range pParts {
		if i >= len(vParts) {
			return false
		}
		if vParts[i] != p {
			return false
		}
	}
	return true
}

// sortVersions sorts semver strings in ascending order.
func sortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) < 0
	})
}

// compareVersions returns -1, 0, or 1.
func compareVersions(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")

	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}

	for i := 0; i < maxLen; i++ {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	return 0
}
