package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/logger"
)

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.app.CloseOverlay()

	case "left", "right", "h", "l":
		m.deleteButton = 1 - m.deleteButton

	case "y":
		m.deleteButton = 0
		return m.executeDelete()

	case "enter":
		if m.deleteButton == 0 {
			return m.executeDelete()
		}
		m.app.CloseOverlay()
	}
	return m, nil
}

func (m Model) executeDelete() (tea.Model, tea.Cmd) {
	if m.deleteName != "" {
		name := m.deleteName
		logger.Log.Debug("deleting session", "name", name)
		if err := lifecycle.New().Delete(name); err != nil {
			m.err = err
			m.app.CloseOverlay()
			return m, nil
		}
		m.sessionGeneration++
		sessions, err := lifecycle.New().List()
		m.err = err
		if err == nil {
			m.setSessions(sessions)
		}
	}
	m.app.CloseOverlay()
	return m, nil
}
