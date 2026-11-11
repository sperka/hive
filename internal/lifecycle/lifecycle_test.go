package lifecycle

import (
	"strings"
	"testing"
)

func TestIsBuiltinToolRecognizesSupportedLaunchProfiles(t *testing.T) {
	for _, tool := range []string{"pi", "claude", "nvim", "bash"} {
		if !IsBuiltinTool(tool) {
			t.Fatalf("expected %q to be builtin", tool)
		}
	}

	if IsBuiltinTool("cursor") {
		t.Fatal("expected unknown tool to be rejected")
	}
}

func TestCreateRejectsUnknownToolBeforeTouchingTmux(t *testing.T) {
	svc := Service{
		generateID: func(int) string { return "fixed" },
		workingDir: func() (string, error) { return "/tmp", nil },
	}

	_, err := svc.Create(CreateRequest{Tool: "cursor"})
	if err == nil {
		t.Fatal("expected unknown tool error")
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

func TestRenameRejectsInvalidNamesBeforeTouchingTmux(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, name := range []string{"", "   ", ".", "../escape", "a:b", "a.b", "a/b", "a\\b", "a|b", "a\nb"} {
		err := New().Rename("work", name)
		if err == nil || !strings.Contains(err.Error(), "session name cannot") {
			t.Fatalf("Rename(%q) = %v, want validation error", name, err)
		}
	}
}

func TestCreateRejectsNamesTmuxWouldNormalize(t *testing.T) {
	_, err := New().Create(CreateRequest{Name: "a.b", Tool: "bash"})
	if err == nil || !strings.Contains(err.Error(), "session name cannot") {
		t.Fatalf("Create() = %v, want validation error", err)
	}
}
