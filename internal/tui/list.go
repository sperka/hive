package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/config"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/logger"
	"github.com/mdipanjan/hive/internal/styles"
)

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.cursor = max(0, m.cursor-1)

	case "down", "j":
		m.cursor = min(len(m.sessions)-1, m.cursor+1)

	case "enter":
		if len(m.sessions) > 0 {
			name := m.sessions[m.cursor].Name
			logger.Log.Debug("attaching to session", "name", name)
			c := lifecycle.New().AttachCmd(name)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return sessionAttachedMsg{err}
			})
		}

	case "n":
		m.err = nil
		m.app.StartNewSession()
		m.form = newForm()
		m.form.FilePicker = newFilePicker()

	case "d":
		if len(m.sessions) > 0 {
			m.app.ConfirmDelete()
			m.deleteButton = 1
		}

	case "t":
		theme := styles.NextTheme()
		logger.Log.Debug("theme switched", "theme", theme.Name)
		cfg := config.Load()
		cfg.Theme = theme.Key
		m.err = config.Save(cfg)

	case "?":
		m.app.ShowHelp()

	case "/":
		m.searchInput = newSearchInput()
		m.app.Search()
		m.searchResults = m.getIndices("")
		m.searchCursor = 0
		return m, textinput.Blink
	}
	return m, nil
}
