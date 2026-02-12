package cmd

import (
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

	// Download binary
	assetName := fmt.Sprintf("opencc-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("https://github.com/dopejs/opencc/releases/download/v%s/%s", target, assetName)

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

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	// Replace current binary
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	// Try direct copy first, fall back to sudo
	if err := copyFile(tmpFile.Name(), binPath); err != nil {
		fmt.Println("Need elevated privileges, trying sudo...")
		if sudoErr := exec.Command("sudo", "cp", tmpFile.Name(), binPath).Run(); sudoErr != nil {
			return fmt.Errorf("install failed: %w", sudoErr)
		}
	}

	fmt.Printf("Successfully upgraded to %s\n", target)
	return nil
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
		fmt.Fprintf(os.Stderr, "\r  %s %5.1f%% %s", bar, pct, formatBytes(pr.current))
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
