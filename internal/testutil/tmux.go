// Package testutil provides isolated tmux servers for integration tests.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Tmux(t *testing.T) {
	t.Helper()
	binary, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux is not installed")
	}
	dir, err := os.MkdirTemp("", "hive-tmux-")
	if err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "socket")
	t.Cleanup(func() {
		_ = exec.Command(binary, "-S", socket, "kill-server").Run()
		_ = os.RemoveAll(dir)
	})
	quote := func(value string) string {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
	script := "#!/bin/sh\nexec " + quote(binary) + " -S " + quote(socket) + " -f /dev/null \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "tmux"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMUX", "")
	t.Setenv("HIVE_STATE_DIR", t.TempDir())
}

func TmuxCommand(t *testing.T, args ...string) string {
	t.Helper()
	output, err := exec.Command("tmux", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("tmux %v: %v: %s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}
