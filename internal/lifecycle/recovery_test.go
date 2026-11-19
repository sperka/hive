package lifecycle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/testutil"
)

func saveRecords(t *testing.T, records ...state.Metadata) {
	t.Helper()
	if err := state.UpdateSavedSessions(func([]state.Metadata) ([]state.Metadata, error) {
		return records, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReconcileAdoptsLiveSessionsAndRetainsThemAfterReboot(t *testing.T) {
	testutil.Tmux(t)
	dir := t.TempDir()
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "-c", dir, "/bin/sh")
	live, missing, err := New().Reconcile()
	if err != nil || len(live) != 1 || len(missing) != 0 {
		t.Fatalf("initial reconcile = %#v, %#v, %v", live, missing, err)
	}
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 1 || saved[0].Name != "work" || saved[0].Path != dir || saved[0].Tool != "sh" {
		t.Fatalf("saved = %#v, err = %v", saved, err)
	}
	testutil.TmuxCommand(t, "kill-server")
	live, missing, err = New().Reconcile()
	if err != nil || len(live) != 0 || len(missing) != 1 || missing[0] != saved[0] {
		t.Fatalf("reboot reconcile = %#v, %#v, %v", live, missing, err)
	}
	if _, err := New().List(); err != nil {
		t.Fatal(err)
	}
	stillSaved, _ := state.ReadSavedSessions()
	if len(stillSaved) != 1 || stillSaved[0] != saved[0] {
		t.Fatal("listing an empty server erased recovery records")
	}
}

func TestReconcileMergesChangesWithoutDroppingMissingSessions(t *testing.T) {
	testutil.Tmux(t)
	saveRecords(t, state.Metadata{Name: "missing", Tool: "codex", Path: "/old"},
		state.Metadata{Name: "work", Tool: "old-tool", Path: "/old"})
	dir := t.TempDir()
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "-c", dir, "/bin/sh")
	live, missing, err := New().Reconcile()
	if err != nil || len(live) != 1 || len(missing) != 1 || missing[0].Name != "missing" {
		t.Fatalf("reconcile = %#v, %#v, %v", live, missing, err)
	}
	saved, _ := state.ReadSavedSessions()
	if len(saved) != 2 || saved[1].Tool != "sh" || saved[1].Path != dir {
		t.Fatalf("live setup not updated: %#v", saved)
	}
}

func TestRenameAndDeleteUpdateSavedSessions(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "/bin/sh")
	if _, err := New().List(); err != nil {
		t.Fatal(err)
	}
	if err := New().Rename("work", "renamed"); err != nil {
		t.Fatal(err)
	}
	saved, _ := state.ReadSavedSessions()
	if len(saved) != 1 || saved[0].Name != "renamed" {
		t.Fatalf("rename left stale saved sessions: %#v", saved)
	}
	if err := New().Delete("renamed"); err != nil {
		t.Fatal(err)
	}
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 0 {
		t.Fatalf("explicit delete left saved sessions: %#v, %v", saved, err)
	}
}

func TestRestoreFailureKeepsSavedEntryAndExistingSessionIsNotReplaced(t *testing.T) {
	testutil.Tmux(t)
	saveRecords(t, state.Metadata{Name: "work", Tool: "missing-tool-for-hive-test", Path: t.TempDir()})
	if err := New().Restore("work"); err == nil {
		t.Fatal("missing executable should prevent restoration")
	}
	saved, _ := state.ReadSavedSessions()
	if len(saved) != 1 {
		t.Fatal("failed restore lost saved entry")
	}
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "/bin/sh")
	id := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=work:", "#{session_id}")
	if err := New().Restore("work"); err != nil {
		t.Fatal(err)
	}
	if got := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=work:", "#{session_id}"); got != id {
		t.Fatal("restore replaced an existing session")
	}
}

func TestForgetDoesNotKillSession(t *testing.T) {
	testutil.Tmux(t)
	saveRecords(t, state.Metadata{Name: "work", Tool: "sh", Path: "/tmp"})
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "work", "/bin/sh")
	if err := New().Forget("work"); err != nil {
		t.Fatal(err)
	}
	testutil.TmuxCommand(t, "has-session", "-t", "=work")
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 0 {
		t.Fatalf("forget = %#v, %v", saved, err)
	}
}

func TestReconcileTmuxErrorPreservesCatalog(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	saveRecords(t, state.Metadata{Name: "work", Tool: "sh", Path: "/tmp"})
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "tmux"), []byte("#!/bin/sh\necho 'permission denied' >&2\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	_, _, err := New().Reconcile()
	if err == nil {
		t.Fatal("tmux error was treated as an empty server")
	}
	saved, _ := state.ReadSavedSessions()
	if len(saved) != 1 || saved[0].Name != "work" {
		t.Fatalf("tmux error lost saved sessions: %#v", saved)
	}
}
