// Package update implements self-update for the kyc-cli binary.
//
// Design (per user/security review):
//   - Explicit `kyc-cli update` subcommand. Never auto-update silently.
//   - On startup, optionally print a one-line nudge if a newer release
//     is available and the user hasn't been told within the last 24h.
//   - Refuse to overwrite the binary if it's owned by a package manager
//     (Homebrew Cellar, Scoop, apt, etc.) — those users get told to use
//     their package manager instead.
//   - Verify the downloaded archive's sha256 against the release's
//     checksums.txt before swapping.
//
// No external Go deps — everything here is std lib + sha256 + tar.
package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repo           = "kyc-rip/cli"
	releaseAPI     = "https://api.github.com/repos/" + repo + "/releases/latest"
	checkCacheFile = "kyc-cli/lastcheck"
	checkInterval  = 24 * time.Hour
)

// LatestRelease is the slice of the GitHub releases API we use.
type LatestRelease struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

// CheckLatest queries GitHub for the most recent (non-draft, non-pre)
// release. Returns ("", nil) if there's nothing newer than `current`.
// Returns the tag (e.g. "v0.1.6") if an upgrade is available.
func CheckLatest(ctx context.Context, current string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "kyc-cli-self-update")
	resp, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github releases: %s", resp.Status)
	}
	var rel LatestRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.Draft || rel.Prerelease || rel.TagName == "" {
		return "", nil
	}
	if !semverNewer(rel.TagName, current) {
		return "", nil
	}
	return rel.TagName, nil
}

// CheckLatestThrottled is CheckLatest with a 24h-on-disk cache so we
// only hit the GH API at most once per day per machine. Returns "" if
// either (a) no newer version, (b) we hit the cache and skipped, or
// (c) the network call failed (don't pester users on flaky networks).
func CheckLatestThrottled(ctx context.Context, current string) string {
	if shouldSkipCheck() {
		return ""
	}
	tag, err := CheckLatest(ctx, current)
	if err != nil {
		return ""
	}
	stampCheck()
	return tag
}

// PromptIfNewer prints a one-line nudge to stderr (so it doesn't pollute
// stdout pipelines) if an update is available. Returns true if a nudge
// was printed.
func PromptIfNewer(ctx context.Context, current string) bool {
	tag := CheckLatestThrottled(ctx, current)
	if tag == "" {
		return false
	}
	fmt.Fprintf(os.Stderr, "kyc-cli %s available — run `kyc-cli update` to upgrade (you have %s)\n", tag, current)
	return true
}

// PackageManagerOrigin returns a non-empty string when the running
// binary appears to live under a package manager's tree, signalling
// "tell the user to run brew/scoop/apt instead of self-update".
func PackageManagerOrigin() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, _ = filepath.EvalSymlinks(exe)
	switch {
	case strings.Contains(exe, "/Cellar/"), strings.Contains(exe, "/.linuxbrew/"):
		return "Homebrew — run `brew upgrade kyc-cli`"
	case strings.Contains(exe, "/scoop/apps/"):
		return "Scoop — run `scoop update kyc-cli`"
	case strings.HasPrefix(exe, "/usr/bin/"), strings.HasPrefix(exe, "/usr/local/bin/dpkg"):
		// /usr/local/bin is fine for self-replace (curl|sh installs land
		// there) but /usr/bin only ever comes from apt.
		if strings.HasPrefix(exe, "/usr/bin/") {
			return "system package manager — run `apt upgrade` or your distro equivalent"
		}
	}
	return ""
}

// DoUpdate downloads the matching tarball/zip for `tag`, verifies sha256
// against checksums.txt, and atomically replaces the running binary.
// Returns nil on success.
func DoUpdate(ctx context.Context, tag string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = resolved
	}

	verNumeric := strings.TrimPrefix(tag, "v")
	asset, ext := assetName(verNumeric)
	if asset == "" {
		return fmt.Errorf("no release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)
	checksumURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/checksums.txt", repo, tag)

	// 1. fetch checksums
	expected, err := fetchChecksum(ctx, checksumURL, asset)
	if err != nil {
		return fmt.Errorf("checksums: %w", err)
	}

	// 2. download archive
	tmp, err := os.CreateTemp("", "kyc-cli-*.bin")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err := downloadVerified(ctx, url, tmp, expected); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// 3. extract binary out of the archive
	binPath, err := extractBinary(tmp.Name(), ext)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	defer os.Remove(binPath)

	// 4. atomic swap (rename within the same dir, then chmod +x)
	dst := exe
	bak := dst + ".bak"
	if err := os.Rename(dst, bak); err != nil {
		return fmt.Errorf("backup current: %w", err)
	}
	if err := os.Rename(binPath, dst); err != nil {
		// Try to roll back
		_ = os.Rename(bak, dst)
		return fmt.Errorf("install new: %w", err)
	}
	_ = os.Chmod(dst, 0o755)
	_ = os.Remove(bak)

	// 5. quick sanity check — run --version on the new binary
	out, err := exec.CommandContext(ctx, dst, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("new binary smoke-test: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	fmt.Printf("upgraded to %s\n", strings.TrimSpace(string(out)))
	return nil
}

// --- helpers ---

func assetName(versionNumeric string) (asset, ext string) {
	osStr, ok := osMap[runtime.GOOS]
	if !ok {
		return "", ""
	}
	archStr, ok := archMap[runtime.GOARCH]
	if !ok {
		return "", ""
	}
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("kyc-cli_%s_%s_%s.zip", versionNumeric, osStr, archStr), ".zip"
	}
	return fmt.Sprintf("kyc-cli_%s_%s_%s.tar.gz", versionNumeric, osStr, archStr), ".tar.gz"
}

var osMap = map[string]string{"linux": "linux", "darwin": "darwin", "windows": "windows"}
var archMap = map[string]string{"amd64": "amd64", "arm64": "arm64"}

func fetchChecksum(ctx context.Context, url, asset string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == asset {
			return fields[0], nil
		}
	}
	return "", errors.New("asset not in checksums.txt")
}

func downloadVerified(ctx context.Context, url string, w *os.File, expectedHex string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(w, h), resp.Body); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expectedHex {
		return fmt.Errorf("sha256 mismatch: want %s got %s", expectedHex, actual)
	}
	if _, err := w.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

func extractBinary(archivePath, ext string) (string, error) {
	if ext == ".tar.gz" {
		return extractTarGz(archivePath, "kyc-cli")
	}
	// .zip
	return "", errors.New(".zip extraction not implemented (Windows users use scoop)")
}

func extractTarGz(path, wanted string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != wanted {
			continue
		}
		out, err := os.CreateTemp("", "kyc-cli-new-*")
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			os.Remove(out.Name())
			return "", err
		}
		out.Close()
		_ = os.Chmod(out.Name(), 0o755)
		return out.Name(), nil
	}
	return "", fmt.Errorf("%q not found in archive", wanted)
}

// --- semver-lite ---

// semverNewer reports whether a is a strictly greater version than b.
// Both are in the form "v1.2.3" or "1.2.3"; non-numeric segments fall
// back to lexical compare.
func semverNewer(a, b string) bool {
	pa := strings.Split(strings.TrimPrefix(strings.TrimSpace(a), "v"), ".")
	pb := strings.Split(strings.TrimPrefix(strings.TrimSpace(b), "v"), ".")
	for i := 0; i < 3 && (i < len(pa) || i < len(pb)); i++ {
		ai, bi := 0, 0
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &ai)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &bi)
		}
		if ai != bi {
			return ai > bi
		}
	}
	return false
}

// --- check throttle ---

func cachePath() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, checkCacheFile)
}

func shouldSkipCheck() bool {
	st, err := os.Stat(cachePath())
	if err != nil {
		return false // never checked → don't skip
	}
	return time.Since(st.ModTime()) < checkInterval
}

func stampCheck() {
	p := cachePath()
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	_ = os.WriteFile(p, []byte(time.Now().UTC().Format(time.RFC3339)), 0o600)
}
