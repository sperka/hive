package tmux

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/testutil"
)

func TestRunnerCommand(t *testing.T) {
	got := runnerCommand("work", "pi", "/tmp/project", "session-id")
	wantParts := []string{
		" run-session",
		"--name 'work'",
		"--tool 'pi'",
		"--path '/tmp/project'",
		"--state-id 'session-id'",
	}

	for _, part := range wantParts {
		if !strings.Contains(got, part) {
			t.Fatalf("runnerCommand() = %q, expected to contain %q", got, part)
		}
	}
}

func TestRenamePreservesLegacyStateAndRunningUpdates(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "/bin/sh")
	originalID := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=work:", "#{session_id}")
	meta := state.Metadata{Name: "work", Tool: "my-agent", Path: "/project", CreatedAt: time.Now().UTC()}
	if err := state.WriteMetadata(meta); err != nil {
		t.Fatal(err)
	}
	if err := Rename("work", "new name"); err != nil {
		t.Fatal(err)
	}
	if err := Rename("new name", "final"); err != nil {
		t.Fatal(err)
	}
	if got := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=final:", "#{session_id}"); got != originalID {
		t.Fatal("renaming replaced the tmux session")
	}
	// A runner launched before renaming still writes to its original state key.
	if err := state.WriteRuntime("work", state.Runtime{State: state.StateCompleted}); err != nil {
		t.Fatal(err)
	}
	sessions, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Name != "final" || sessions[0].Tool != meta.Tool ||
		sessions[0].Path != meta.Path || !sessions[0].CreatedAt.Equal(meta.CreatedAt) ||
		sessions[0].Status != session.StatusCompleted {
		t.Fatalf("renamed sessions = %#v", sessions)
	}
	saved, err := state.ReadMetadata("work")
	if err != nil || saved.Name != "final" {
		t.Fatalf("metadata = %#v, err = %v", saved, err)
	}
	if err := Kill("final"); err != nil {
		t.Fatal(err)
	}
	if _, err := state.ReadMetadata("work"); !os.IsNotExist(err) {
		t.Fatalf("renamed session state was not deleted: %v", err)
	}
}

func TestRenameDuplicateAndMissingTargetsLeaveMetadataIntact(t *testing.T) {
	testutil.Tmux(t)
	for _, name := range []string{"work", "taken"} {
		testutil.TmuxCommand(t, "new-session", "-d", "-s", name, "/bin/sh")
	}
	meta := state.Metadata{Name: "work", Tool: "zsh"}
	if err := state.WriteMetadata(meta); err != nil {
		t.Fatal(err)
	}
	for _, names := range [][2]string{{"work", "taken"}, {"wor", "other"}, {"missing", "other"}} {
		if err := Rename(names[0], names[1]); err == nil {
			t.Fatalf("Rename(%q, %q) should fail", names[0], names[1])
		}
	}
	got, err := state.ReadMetadata("work")
	if err != nil || got != meta {
		t.Fatalf("failed rename changed metadata: %#v, %v", got, err)
	}
	testutil.TmuxCommand(t, "has-session", "-t", "=work")
}

func TestRenameUnmanagedSession(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "outside", "/bin/sh")
	if err := Rename("outside", "-renamed"); err != nil {
		t.Fatal(err)
	}
	testutil.TmuxCommand(t, "has-session", "-t", "=-renamed")
}

func TestRenamePreservesStableStateID(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "/bin/sh")
	testutil.TmuxCommand(t, "set-option", "-t", "=work:", "@hive_state_id", "stable-id")
	meta := state.Metadata{Name: "work", Tool: "codex"}
	if err := state.WriteMetadataForID("stable-id", meta); err != nil {
		t.Fatal(err)
	}
	if err := Rename("work", "renamed"); err != nil {
		t.Fatal(err)
	}
	got, err := state.ReadMetadata("stable-id")
	if err != nil || got.Name != "renamed" || got.Tool != "codex" {
		t.Fatalf("metadata = %#v, err = %v", got, err)
	}
	if err := Kill("renamed"); err != nil {
		t.Fatal(err)
	}
	if _, err := state.ReadMetadata("stable-id"); !os.IsNotExist(err) {
		t.Fatalf("stable state was not deleted: %v", err)
	}
}

func TestRunnerCommandQuotesArgs(t *testing.T) {
	got := runnerCommand("work's", "nvim", "/tmp/my project", "session-id")
	wantParts := []string{
		"--name 'work'\\''s'",
		"--tool 'nvim'",
		"--path '/tmp/my project'",
	}

	for _, part := range wantParts {
		if !strings.Contains(got, part) {
			t.Fatalf("runnerCommand() = %q, expected to contain %q", got, part)
		}
	}
}

func TestEditorLayoutCommands(t *testing.T) {
	got := editorLayoutCommands("work", "/tmp/project")
	want := [][]string{
		{"split-window", "-h", "-p", "35", "-t", "work:0", "-c", "/tmp/project"},
		{"select-pane", "-t", "work:0.0"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("editorLayoutCommands() = %#v, want %#v", got, want)
	}
}

func TestSessionKeyBindingsOpenHiveSwitch(t *testing.T) {
	got := sessionKeyBindings()

	if len(got) != 1 {
		t.Fatalf("sessionKeyBindings() returned %d bindings, want 1", len(got))
	}

	binding := got[0]
	wantPrefix := []string{"bind-key", "h", "display-popup", "-B", "-w", "90%", "-h", "80%", "-E"}
	if len(binding) != len(wantPrefix)+1 {
		t.Fatalf("sessionKeyBindings()[0] = %#v, want %d args", binding, len(wantPrefix)+1)
	}
	for i, want := range wantPrefix {
		if binding[i] != want {
			t.Fatalf("sessionKeyBindings()[0] = %#v, arg %d want %q", binding, i, want)
		}
	}
	if !strings.HasSuffix(binding[len(binding)-1], " switch") {
		t.Fatalf("sessionKeyBindings()[0] command = %q, want hive switch command", binding[len(binding)-1])
	}
}

func TestSessionOptions(t *testing.T) {
	got := sessionOptions("work")
	want := [][]string{
		{"set-option", "-t", "work", "status", "off"},
		{"set-option", "-t", "work", "mouse", "on"},
		{"set-option", "-t", "work", "pane-border-style", "fg=#45475a"},
		{"set-option", "-t", "work", "pane-active-border-style", "fg=#a6e3a1"},
		{"set-option", "-t", "work", "window-style", "bg=#1e1e2e"},
		{"set-option", "-t", "work", "window-active-style", "bg=#1e1e2e"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sessionOptions() = %#v, want %#v", got, want)
	}
}
