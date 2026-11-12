// Package update installs checksum-verified releases from the Hive fork.
package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"
)

const Repository = "sperka/hive"

type updater struct {
	client     *http.Client
	latestURL  string
	executable func() (string, error)
	goos       string
	goarch     string
}

func Run(ctx context.Context, currentVersion string, checkOnly bool, out io.Writer) error {
	u := updater{
		client:     http.DefaultClient,
		latestURL:  "https://github.com/" + Repository + "/releases/latest",
		executable: os.Executable,
		goos:       runtime.GOOS,
		goarch:     runtime.GOARCH,
	}
	return u.run(ctx, currentVersion, checkOnly, out)
}

func (u updater) run(ctx context.Context, currentVersion string, checkOnly bool, out io.Writer) error {
	if (u.goos != "linux" && u.goos != "darwin") || (u.goarch != "amd64" && u.goarch != "arm64") {
		return fmt.Errorf("updates are not available for %s/%s", u.goos, u.goarch)
	}
	fmt.Fprintf(out, "Checking %s releases...\n", Repository)
	latest, err := u.latestRelease(ctx)
	if err != nil {
		return err
	}

	current := "v" + strings.TrimPrefix(currentVersion, "v")
	if semver.IsValid(current) && semver.Compare(current, latest) >= 0 {
		fmt.Fprintf(out, "Hive %s is up to date (latest release: %s).\n", currentVersion, latest)
		return nil
	}
	fmt.Fprintf(out, "Update available: %s -> %s\n", currentVersion, latest)
	if checkOnly {
		return nil
	}

	archiveName := fmt.Sprintf("hive_%s_%s.tar.gz", u.goos, u.goarch)
	// Pin both downloads to the resolved tag so a concurrent release cannot mix assets.
	downloadURL := strings.TrimSuffix(u.latestURL, "/latest") + "/download/" + url.PathEscape(latest) + "/"
	archiveURL, checksumURL := downloadURL+archiveName, downloadURL+"checksums.txt"

	target, err := u.executable()
	if err != nil {
		return fmt.Errorf("locate Hive executable: %w", err)
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return fmt.Errorf("resolve Hive executable: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("Hive executable is not a regular file: %s", target)
	}
	staged, err := os.CreateTemp(filepath.Dir(target), ".hive-update-*")
	if err != nil {
		return fmt.Errorf("cannot update %s: %w; use an account with write access or reinstall in a user-writable directory", target, err)
	}
	defer os.Remove(staged.Name())
	defer staged.Close()

	body, err := u.get(ctx, checksumURL)
	if err != nil {
		return err
	}
	expected, err := checksum(body, archiveName)
	body.Close()
	if err != nil {
		return err
	}

	archive, err := os.CreateTemp("", "hive-download-*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(archive.Name())
	defer archive.Close()
	fmt.Fprintf(out, "Downloading %s...\n", archiveName)
	body, err = u.get(ctx, archiveURL)
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(archive, hash), body)
	body.Close()
	if err != nil {
		return fmt.Errorf("download archive: %w", err)
	}
	if hex.EncodeToString(hash.Sum(nil)) != expected {
		return errors.New("archive checksum mismatch; Hive was not changed")
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := extractBinary(archive, staged); err != nil {
		return err
	}
	if err := staged.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if err := staged.Sync(); err != nil {
		return err
	}
	if err := staged.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(staged.Name(), target); err != nil {
		return fmt.Errorf("replace %s: %w", target, err)
	}
	fmt.Fprintf(out, "Updated Hive to %s at %s. Reopen Hive to use it.\n", latest, target)
	return nil
}

func (u updater) latestRelease(ctx context.Context) (string, error) {
	// GitHub's public redirect does not consume the anonymous REST API quota.
	resp, err := u.request(ctx, http.MethodHead, u.latestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	prefix := strings.TrimSuffix(u.latestURL, "/latest") + "/tag/"
	resolved := resp.Request.URL.String()
	if !strings.HasPrefix(resolved, prefix) {
		return "", fmt.Errorf("invalid stable release URL %q", resolved)
	}
	tag, err := url.PathUnescape(strings.TrimPrefix(resolved, prefix))
	if err != nil || !semver.IsValid(tag) || semver.Prerelease(tag) != "" {
		return "", fmt.Errorf("invalid stable release version %q", tag)
	}
	return tag, nil
}

func (u updater) get(ctx context.Context, url string) (io.ReadCloser, error) {
	resp, err := u.request(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (u updater) request(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "hive-updater")
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch %s: HTTP %s", url, resp.Status)
	}
	return resp, nil
}

func checksum(r io.Reader, name string) (string, error) {
	var found string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		value, err := hex.DecodeString(fields[0])
		if err != nil || len(value) != sha256.Size || found != "" {
			return "", fmt.Errorf("invalid or duplicate checksum for %s", name)
		}
		found = hex.EncodeToString(value)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	if found == "" {
		return "", fmt.Errorf("no checksum for %s", name)
	}
	return found, nil
}

func extractBinary(archive io.Reader, dst io.Writer) error {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	found := false
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read release archive: %w", err)
		}
		// Extract only the expected root entry, never paths or links from the archive.
		if header.Name != "hive" {
			continue
		}
		if found || header.Typeflag != tar.TypeReg || header.Size == 0 {
			return errors.New("release archive contains an invalid Hive binary")
		}
		if _, err := io.Copy(dst, tr); err != nil {
			return fmt.Errorf("extract Hive binary: %w", err)
		}
		found = true
	}
	if !found {
		return errors.New("release archive does not contain the Hive binary")
	}
	// Read the gzip trailer too, so truncated or corrupt streams fail before replacement.
	if _, err := io.Copy(io.Discard, gz); err != nil {
		return fmt.Errorf("read release archive trailer: %w", err)
	}
	return nil
}
