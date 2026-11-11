package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/testutil"
)

func renameKey(m Model, key tea.KeyType) Model {
	model, _ := m.Update(tea.KeyMsg{Type: key})
	return model.(Model)
}

func TestRenameDialogEditingAndCancel(t *testing.T) {
	m := Model{app: NewAppState(), sessions: []session.Session{{Name: "work"}}}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = model.(Model)
	if cmd == nil || !m.app.RenamingSession() || m.renameInput.Value() != "work" {
		t.Fatal("rename should open with current name focused")
	}
	m = renameKey(m, tea.KeyBackspace)
	if m.renameInput.Value() != "wor" {
		t.Fatal("rename input is not editable")
	}
	m = renameKey(m, tea.KeyEsc)
	if m.app.RenamingSession() || m.sessions[0].Name != "work" {
		t.Fatal("cancel should preserve the old name")
	}
}

func TestRenameWithNoSessions(t *testing.T) {
	m := Model{app: NewAppState()}
	m = renameKey(m, tea.KeyDown)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = model.(Model)
	if m.app.RenamingSession() || m.cursor != 0 {
		t.Fatal("empty list should not open rename or move cursor out of bounds")
	}
}

func TestRenameDialogShowsErrorsAndKeepsSelectionAfterSuccess(t *testing.T) {
	testutil.Tmux(t)
	for _, name := range []string{"work", "taken"} {
		testutil.TmuxCommand(t, "new-session", "-d", "-s", name, "/bin/sh")
	}
	m := Model{app: NewAppState(), sessions: []session.Session{{Name: "work"}, {Name: "taken"}}}
	model, _ := m.startRename()
	m = model.(Model)
	for _, name := range []string{"", "taken"} {
		m.renameInput.SetValue(name)
		m = renameKey(m, tea.KeyEnter)
		if m.err == nil || !m.app.RenamingSession() || !strings.Contains(m.View(), "Error:") {
			t.Fatal("invalid or duplicate name should keep dialog open with an error")
		}
	}
	m.renameInput.SetValue("z-renamed")
	m = renameKey(m, tea.KeyEnter)
	if m.err != nil || m.app.RenamingSession() || m.sessions[m.cursor].Name != "z-renamed" {
		t.Fatalf("rename failed to keep selection: cursor=%d, sessions=%#v, err=%v", m.cursor, m.sessions, m.err)
	}
}
