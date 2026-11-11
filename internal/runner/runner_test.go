package runner

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/testutil"
)

func TestRunOnceWritesRuntimeUnderStableID(t *testing.T) {
	t.Setenv("HIVE_STATE_DIR", t.TempDir())
	t.Setenv("SHELL", "/bin/true")
	opts := Options{Name: "original", StateID: "stable-id", Tool: "/bin/true"}
	if err := runOnce(opts); err != nil {
		t.Fatal(err)
	}
	runtime, err := state.ReadRuntime("stable-id")
	if err != nil || runtime.State != state.StateCompleted || runtime.ExitCode == nil || *runtime.ExitCode != 0 {
		t.Fatalf("runtime = %#v, err = %v", runtime, err)
	}
	entries, err := os.ReadDir(os.Getenv("HIVE_STATE_DIR"))
	if err != nil || len(entries) != 1 || entries[0].Name() != "stable-id" {
		t.Fatalf("runner created state under the display name: %v, %v", entries, err)
	}
}

func TestDetachResolvesRenamedSessionFromPane(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "original", "/bin/sh")
	pane := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=original:", "#{pane_id}")
	t.Setenv("TMUX_PANE", pane)
	client := exec.Command("tmux", "-C", "attach-session", "-t", "=original")
	input, err := client.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := client.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	client.Stderr = os.Stderr
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = input.Close()
		_ = client.Process.Kill()
		_ = client.Wait()
	})
	scanner := bufio.NewScanner(output)
	attached := false
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "%session-changed ") {
			attached = true
			break
		}
	}
	if !attached {
		t.Fatal("control client did not attach")
	}
	testutil.TmuxCommand(t, "rename-session", "-t", "=original", "renamed")
	if err := detachClient("original"); err != nil {
		t.Fatalf("detach did not resolve the renamed session: %v", err)
	}
	if got := testutil.TmuxCommand(t, "display-message", "-p", "-t", "=renamed:", "#{session_attached}"); got != "0" {
		t.Fatal("client remained attached after editor detach")
	}
}
