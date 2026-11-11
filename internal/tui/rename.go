package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/lifecycle"
)

func (m Model) startRename() (tea.Model, tea.Cmd) {
	if len(m.sessions) == 0 {
		return m, nil
	}
	m.cursor = max(0, min(m.cursor, len(m.sessions)-1))
	m.renameName = m.sessions[m.cursor].Name
	m.renameInput = textinput.New()
	m.renameInput.Width = 40
	m.renameInput.SetValue(m.renameName)
	m.renameInput.CursorEnd()
	m.err = nil
	m.app.RenameSession()
	return m, m.renameInput.Focus()
}

func (m Model) updateRename(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.err = nil
		m.app.CloseOverlay()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.renameInput.Value())
		if err := lifecycle.New().Rename(m.renameName, name); err != nil {
			m.err = err
			return m, nil
		}
		// Keep the local list accurate even if refreshing from tmux fails.
		for i := range m.sessions {
			if m.sessions[i].Name == m.renameName {
				m.sessions[i].Name = name
			}
		}
		sessions, err := lifecycle.New().List()
		if err == nil {
			m.sessions = sessions
		}
		m.cursor = 0
		for i, session := range m.sessions {
			if session.Name == name {
				m.cursor = i
				break
			}
		}
		m.err = err
		m.app.CloseOverlay()
		return m, nil
	}
	m.err = nil
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}
