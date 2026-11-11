package lifecycle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateRejectsMissingExecutableBeforeTouchingTmux(t *testing.T) {
	svc := Service{
		generateID: func(int) string { return "fixed" },
		workingDir: func() (string, error) { return "/tmp", nil },
	}

	t.Setenv("PATH", t.TempDir())
	_, err := svc.Create(CreateRequest{Tool: "missing-tool"})
	if err == nil || !strings.Contains(err.Error(), "not an executable") {
		t.Fatalf("expected executable error, got %v", err)
	}
}

func TestCreateAcceptsCustomExecutable(t *testing.T) {
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	t.Setenv("HIVE_STATE_DIR", t.TempDir())
	for _, name := range []string{"my-agent", "tmux"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	svc := New()
	for _, tool := range []string{"my-agent", filepath.Join(bin, "my-agent")} {
		name, err := svc.Create(CreateRequest{Name: "custom", Tool: tool, Path: bin})
		if err != nil {
			t.Fatal(err)
		}
		if name != "custom" {
			t.Fatalf("name = %q, want custom", name)
		}
	}
}

func TestResolvePathUsesWorkingDirForEmptyOrDot(t *testing.T) {
	svc := Service{workingDir: func() (string, error) { return "/work/project", nil }}

	for _, value := range []string{"", "."} {
		got, err := svc.resolvePath(value)
		if err != nil {
			t.Fatal(err)
		}
		if got != "/work/project" {
			t.Fatalf("resolvePath(%q) = %q", value, got)
		}
	}
}
