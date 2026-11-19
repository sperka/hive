package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/testutil"
)

func recoveryModel(t *testing.T) Model {
	t.Helper()
	testutil.Tmux(t)
	err := state.UpdateSavedSessions(func([]state.Metadata) ([]state.Metadata, error) {
		return []state.Metadata{
			{Name: "first", Tool: "sh", Path: "/tmp"},
			{Name: "second", Tool: "sh", Path: "/tmp"},
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return New()
}

func recoveryKey(m Model, key string) Model {
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return model.(Model)
}

func TestStartupRecoveryShowsSetupAndLaterKeepsSavedEntries(t *testing.T) {
	m := recoveryModel(t)
	if !m.app.RecoveringSessions() || len(m.recoverySessions) != 2 {
		t.Fatalf("startup did not offer recovery: %#v", m)
	}
	for _, text := range []string{"RECOVER SESSION", "first", "sh", "/tmp", "Open", "Delete saved entry", "Later"} {
		if !strings.Contains(m.View(), text) {
			t.Fatalf("recovery view lacks %q", text)
		}
	}
	m = recoveryKey(m, "l")
	if len(m.recoverySessions) != 1 || m.recoverySessions[0].Name != "second" {
		t.Fatal("Later did not advance to the next saved session")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.app.RecoveringSessions() {
		t.Fatal("Escape did not dismiss recovery")
	}
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 2 {
		t.Fatalf("postponing recovery lost sessions: %#v, %v", saved, err)
	}
}

func TestRecoveryDeleteOnlyRemovesSelectedSavedEntry(t *testing.T) {
	m := recoveryModel(t)
	m = recoveryKey(m, "d")
	saved, err := state.ReadSavedSessions()
	if err != nil || len(saved) != 1 || saved[0].Name != "second" {
		t.Fatalf("delete saved entry = %#v, %v", saved, err)
	}
	if !m.app.RecoveringSessions() || m.recoverySessions[0].Name != "second" {
		t.Fatal("delete did not advance to remaining recovery")
	}
}

func TestRecoveryOpenFailureStaysVisible(t *testing.T) {
	m := recoveryModel(t)
	err := state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		saved[0].Path = "/missing-hive-test-directory"
		return saved, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	m = recoveryKey(m, "o")
	if m.err == nil || !m.app.RecoveringSessions() || !strings.Contains(m.View(), "Error:") {
		t.Fatal("failed restoration should keep the prompt and show the error")
	}
	saved, _ := state.ReadSavedSessions()
	if len(saved) != 2 {
		t.Fatal("failed restoration lost saved entry")
	}
}

func TestRefreshPreservesSelectionAndOverlay(t *testing.T) {
	m := Model{
		app: NewAppState(), sessions: []session.Session{{Name: "b"}, {Name: "c"}}, cursor: 1,
	}
	m.app.ShowHelp()
	updated, _ := m.Update(sessionsRefreshedMsg{sessions: []session.Session{{Name: "a"}, {Name: "b"}, {Name: "c"}}})
	m = updated.(Model)
	if !m.app.ShowingHelp() || m.cursor != 2 || m.sessions[m.cursor].Name != "c" {
		t.Fatal("refresh disturbed overlay or selected session")
	}
	updated, _ = m.Update(sessionsRefreshedMsg{err: fmt.Errorf("tmux unavailable")})
	m = updated.(Model)
	if len(m.sessions) != 3 || m.err == nil {
		t.Fatal("failed refresh should retain the live list and show an error")
	}
}

func TestStaleRefreshCannotUndoMutation(t *testing.T) {
	m := Model{app: NewAppState(), sessions: []session.Session{{Name: "renamed"}}, sessionGeneration: 1}
	updated, _ := m.Update(sessionsRefreshedMsg{sessions: []session.Session{{Name: "old"}}, generation: 0})
	m = updated.(Model)
	if len(m.sessions) != 1 || m.sessions[0].Name != "renamed" {
		t.Fatal("stale snapshot undid a newer local session change")
	}
}

func TestRefreshDismissesRecoveryForSessionsNowLive(t *testing.T) {
	m := recoveryModel(t)
	updated, _ := m.Update(sessionsRefreshedMsg{sessions: []session.Session{{Name: "first"}}})
	m = updated.(Model)
	if len(m.recoverySessions) != 1 || m.recoverySessions[0].Name != "second" {
		t.Fatal("recovery still offered an active session")
	}
	updated, _ = m.Update(sessionsRefreshedMsg{sessions: []session.Session{{Name: "first"}, {Name: "second"}}})
	if updated.(Model).app.RecoveringSessions() {
		t.Fatal("recovery should close when all sessions have reappeared")
	}
}

func TestSwitcherDoesNotInterruptWithRecovery(t *testing.T) {
	recoveryModel(t)
	m := NewSwitch()
	if m.app.RecoveringSessions() || !m.app.Searching() {
		t.Fatal("compact switcher should stay focused on switching live sessions")
	}
}

func TestRefreshCannotChangeDeleteConfirmationTarget(t *testing.T) {
	testutil.Tmux(t)
	for _, name := range []string{"first", "second"} {
		testutil.TmuxCommand(t, "new-session", "-d", "-s", name, "/bin/sh")
	}
	m := recoveryKey(New(), "d")
	testutil.TmuxCommand(t, "kill-session", "-t", "=first")
	updated, _ := m.Update(m.refreshSessions()())
	m = recoveryKey(updated.(Model), "y")
	if m.err == nil {
		t.Fatal("deleting a vanished session should report an error")
	}
	testutil.TmuxCommand(t, "has-session", "-t", "=second")
}

func TestRecoveryViewFitsViewport(t *testing.T) {
	for _, size := range [][2]int{{40, 18}, {60, 20}, {80, 24}} {
		m := Model{
			app: NewAppState(), width: size[0], height: size[1],
			recoverySessions: []state.Metadata{{
				Name: strings.Repeat("long name", 15), Tool: strings.Repeat("tool", 30),
				Path: strings.Repeat("/long/path", 30),
			}},
			err: fmt.Errorf("%s", strings.Repeat("long error ", 30)),
		}
		m.app.RecoverSessions()
		view := m.View()
		if w, h := lipgloss.Size(view); w > size[0] || h > size[1] {
			t.Fatalf("recovery overflowed %dx%d: %dx%d", size[0], size[1], w, h)
		}
		if !strings.Contains(view, "RECOVER SESSION") || !strings.Contains(view, "Delete saved entry") {
			t.Fatal("recovery choices must remain usable in the viewport")
		}
	}
}
