package lifecycle_test

import (
	"os"
	"testing"

	"github.com/mdipanjan/hive/internal/cli"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/testutil"
)

// tmux launches the test executable as Hive's session runner. Use the real CLI
// and runner so restoration exercises the same subprocess path as a built Hive.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "run-session" {
		cli.Run(os.Args)
		return
	}
	os.Exit(m.Run())
}

func TestCreateAndRestoreLaunchSavedToolInSavedDirectory(t *testing.T) {
	testutil.Tmux(t)
	dir := t.TempDir() + "/project with ' quote"
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	svc := lifecycle.New()
	if _, err := svc.Create(lifecycle.CreateRequest{Name: "work", Tool: "/bin/sh", Path: dir}); err != nil {
		t.Fatal(err)
	}
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 1 || saved[0].Name != "work" || saved[0].Tool != "/bin/sh" || saved[0].Path != dir {
		t.Fatalf("create did not save the setup: %#v, %v", saved, err)
	}
	testutil.TmuxCommand(t, "kill-server")
	if err := svc.Restore("work"); err != nil {
		t.Fatal(err)
	}
	got := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=work:", "#{pane_current_path}")
	if got != dir {
		t.Fatalf("restored directory = %q, want %q", got, dir)
	}
	live, missing, err := svc.Reconcile()
	if err != nil || len(live) != 1 || len(missing) != 0 || live[0].Tool != "/bin/sh" || live[0].Path != dir {
		t.Fatalf("restored sessions = %#v, missing = %#v, err = %v", live, missing, err)
	}
}
