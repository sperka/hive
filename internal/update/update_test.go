package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

type archiveEntry struct {
	name string
	kind byte
	body string
}

func makeArchive(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		size := int64(len(entry.body))
		if entry.kind == tar.TypeSymlink {
			size = 0
		}
		if err := tw.WriteHeader(&tar.Header{
			Name: entry.name, Typeflag: entry.kind, Mode: 0755, Size: size, Linkname: "elsewhere",
		}); err != nil {
			t.Fatal(err)
		}
		if size > 0 {
			if _, err := io.WriteString(tw, entry.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUpdate(t *testing.T) {
	goodArchive := makeArchive(t, archiveEntry{"hive", tar.TypeReg, "new binary"})
	tests := []struct {
		name       string
		current    string
		tag        string
		checkOnly  bool
		symlink    bool
		archive    []byte
		checksums  string
		status     int
		missing    bool
		prerelease bool
		wantUpdate bool
		wantError  string
		wantOutput string
	}{
		{name: "upgrade", current: "0.1.2", wantUpdate: true},
		{name: "symlink", current: "v0.1.2", symlink: true, wantUpdate: true},
		{name: "development build", current: "dev", wantUpdate: true},
		{name: "release candidate", current: "0.2.0-rc.1", wantUpdate: true},
		{name: "check", current: "0.1.2", checkOnly: true, wantOutput: "Update available"},
		{name: "same version", current: "0.2.0", wantOutput: "up to date"},
		{name: "build metadata", current: "0.2.0+local", wantOutput: "up to date"},
		{name: "newer version", current: "0.10.0", wantOutput: "up to date"},
		{name: "invalid version", current: "0.1.2", tag: "not-a-version", wantError: "invalid stable release"},
		{name: "prerelease", current: "0.1.2", prerelease: true, wantError: "invalid stable release"},
		{name: "no release", current: "0.1.2", status: 404, wantError: "404"},
		{name: "rate limited", current: "0.1.2", status: 403, wantError: "403"},
		{name: "missing assets", current: "0.1.2", missing: true, wantError: "missing"},
		{name: "wrong checksum", current: "0.1.2",
			checksums: strings.Repeat("0", 64) + "  hive_linux_arm64.tar.gz\n", wantError: "checksum mismatch"},
		{name: "missing checksum", current: "0.1.2", checksums: "unrelated\n", wantError: "no checksum"},
		{name: "invalid archive", current: "0.1.2", archive: []byte("not gzip"), wantError: "open release archive"},
		{name: "truncated archive", current: "0.1.2", archive: goodArchive[:len(goodArchive)-6], wantError: "archive"},
		{name: "missing binary", current: "0.1.2",
			archive: makeArchive(t, archiveEntry{"README.md", tar.TypeReg, "docs"}), wantError: "does not contain"},
		{name: "path traversal", current: "0.1.2",
			archive: makeArchive(t, archiveEntry{"../hive", tar.TypeReg, "bad"}), wantError: "does not contain"},
		{name: "binary symlink", current: "0.1.2",
			archive: makeArchive(t, archiveEntry{"hive", tar.TypeSymlink, ""}), wantError: "invalid Hive binary"},
		{name: "empty binary", current: "0.1.2",
			archive: makeArchive(t, archiveEntry{"hive", tar.TypeReg, ""}), wantError: "invalid Hive binary"},
		{name: "duplicate binary", current: "0.1.2",
			archive:   makeArchive(t, archiveEntry{"hive", tar.TypeReg, "one"}, archiveEntry{"hive", tar.TypeReg, "two"}),
			wantError: "invalid Hive binary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "hive")
			if err := os.WriteFile(target, []byte("old binary"), 0751); err != nil {
				t.Fatal(err)
			}
			old, err := os.Open(target)
			if err != nil {
				t.Fatal(err)
			}
			defer old.Close()
			executable := target
			if tt.symlink {
				executable = filepath.Join(dir, "hive-link")
				if err := os.Symlink("hive", executable); err != nil {
					t.Fatal(err)
				}
			}
			archive := tt.archive
			if archive == nil {
				archive = goodArchive
			}
			checksums := tt.checksums
			if checksums == "" {
				checksums = fmt.Sprintf("%x  hive_linux_arm64.tar.gz\n", sha256.Sum256(archive))
			}
			tag := tt.tag
			if tag == "" {
				tag = "v0.2.0"
			}
			var downloads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/latest":
					if tt.status != 0 {
						w.WriteHeader(tt.status)
						return
					}
					assets := []map[string]string{}
					if !tt.missing {
						for _, name := range []string{"hive_linux_arm64.tar.gz", "checksums.txt"} {
							assets = append(assets, map[string]string{
								"name": name, "browser_download_url": "http://" + r.Host + "/" + name,
							})
						}
					}
					json.NewEncoder(w).Encode(map[string]any{
						"tag_name": tag, "prerelease": tt.prerelease, "assets": assets,
					})
				case "/checksums.txt":
					downloads.Add(1)
					io.WriteString(w, checksums)
				case "/hive_linux_arm64.tar.gz":
					downloads.Add(1)
					w.Write(archive)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			u := updater{
				client: server.Client(), latestURL: server.URL + "/latest",
				executable: func() (string, error) { return executable, nil },
				goos:       "linux", goarch: "arm64",
			}
			var out bytes.Buffer
			err = u.run(context.Background(), tt.current, tt.checkOnly, &out)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), tt.wantOutput) {
				t.Fatalf("output = %q, want %q", out.String(), tt.wantOutput)
			}
			if tt.checkOnly || tt.wantOutput == "up to date" {
				if downloads.Load() != 0 {
					t.Fatalf("unexpected asset downloads: %d", downloads.Load())
				}
			}
			want := "old binary"
			if tt.wantUpdate {
				want = "new binary"
			}
			data, err := os.ReadFile(target)
			if err != nil || string(data) != want {
				t.Fatalf("installed binary = %q, error %v; want %q", data, err, want)
			}
			// An existing process keeps its old inode when an update succeeds.
			data, err = io.ReadAll(old)
			if err != nil || string(data) != "old binary" {
				t.Fatalf("old executable was modified: %q, %v", data, err)
			}
			info, err := os.Stat(target)
			if err != nil || info.Mode().Perm() != 0751 {
				t.Fatalf("executable permissions changed: %v, %v", info, err)
			}
			if tt.symlink {
				if _, err := os.Readlink(executable); err != nil {
					t.Fatalf("symlink was replaced: %v", err)
				}
			}
			leftovers, _ := filepath.Glob(filepath.Join(dir, ".hive-update-*"))
			if len(leftovers) != 0 {
				t.Fatalf("leftover staged files: %v", leftovers)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	for _, text := range []string{
		hash + "  hive_linux_arm64.tar.gz\n",
		strings.ToUpper(hash) + " *hive_linux_arm64.tar.gz\r\n",
	} {
		got, err := checksum(strings.NewReader(text), "hive_linux_arm64.tar.gz")
		if err != nil || got != hash {
			t.Fatalf("checksum = %q, %v", got, err)
		}
	}
	for _, text := range []string{
		"bad  hive_linux_arm64.tar.gz\n",
		strings.Repeat(hash+"  hive_linux_arm64.tar.gz\n", 2),
	} {
		if _, err := checksum(strings.NewReader(text), "hive_linux_arm64.tar.gz"); err == nil {
			t.Fatalf("accepted invalid checksums %q", text)
		}
	}
}

func TestCancelledUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u := updater{client: http.DefaultClient, latestURL: "http://localhost/latest", goos: "linux", goarch: "amd64"}
	if err := u.run(ctx, "0.1.2", false, io.Discard); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("error = %v, want cancellation", err)
	}
}
