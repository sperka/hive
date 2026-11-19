package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/components"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/styles"
)

type sessionsRefreshedMsg struct {
	sessions   []session.Session
	err        error
	generation uint64
}

func (m Model) refreshSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := lifecycle.New().List()
		return sessionsRefreshedMsg{sessions: sessions, err: err, generation: m.sessionGeneration}
	}
}

func (m *Model) setSessions(sessions []session.Session) {
	selected := ""
	if m.cursor >= 0 && m.cursor < len(m.sessions) {
		selected = m.sessions[m.cursor].Name
	}
	m.sessions = sessions
	m.cursor = max(0, min(m.cursor, len(sessions)-1))
	for i, sess := range sessions {
		if sess.Name == selected {
			m.cursor = i
		}
	}
	if m.app.Searching() {
		m.searchResults = m.getIndices(m.searchInput.Value())
		m.searchCursor = max(0, min(m.searchCursor, len(m.searchResults)-1))
	}
	m.recoverySessions = slices.DeleteFunc(m.recoverySessions, func(record state.Metadata) bool {
		return slices.ContainsFunc(sessions, func(sess session.Session) bool { return sess.Name == record.Name })
	})
	if m.app.RecoveringSessions() && len(m.recoverySessions) == 0 {
		m.app.CloseOverlay()
	}
}

func (m Model) updateRecovery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.recoverySessions = nil
		m.err = nil
		m.app.CloseOverlay()
		return m, nil
	case "left", "shift+tab":
		m.recoveryButton = (m.recoveryButton + 2) % 3
	case "right", "tab":
		m.recoveryButton = (m.recoveryButton + 1) % 3
	case "o":
		m.recoveryButton = 0
		return m.resolveRecovery()
	case "d":
		m.recoveryButton = 1
		return m.resolveRecovery()
	case "l":
		m.recoveryButton = 2
		return m.resolveRecovery()
	case "enter":
		return m.resolveRecovery()
	}
	return m, nil
}

func (m Model) resolveRecovery() (tea.Model, tea.Cmd) {
	if len(m.recoverySessions) == 0 {
		m.app.CloseOverlay()
		return m, nil
	}
	name := m.recoverySessions[0].Name
	var err error
	switch m.recoveryButton {
	case 0:
		err = lifecycle.New().Restore(name)
	case 1:
		err = lifecycle.New().Forget(name)
	}
	m.err = err
	if err != nil {
		return m, nil
	}
	m.recoverySessions = m.recoverySessions[1:]
	m.recoveryButton = 0
	m.sessionGeneration++
	sessions, err := lifecycle.New().List()
	m.err = err
	if err == nil {
		m.setSessions(sessions)
	}
	if len(m.recoverySessions) == 0 {
		m.app.CloseOverlay()
	}
	return m, nil
}

func renderRecoveryView(m Model) string {
	if len(m.recoverySessions) == 0 {
		return renderListView(m)
	}
	width := 66
	if m.width > 0 {
		width = min(width, m.width)
	}
	if width < 40 || (m.height > 0 && m.height < 18) {
		return ansi.Truncate("HIVE - enlarge terminal to 40 x 18 for recovery", width, "")
	}
	inner := width - 6
	record := m.recoverySessions[0]
	line := func(label, value string) string {
		value = strings.Join(strings.Fields(ansi.Strip(value)), " ")
		return styles.Normal.Render(ansi.Truncate(label+value, inner, "..."))
	}
	content := styles.PanelTitle.Render("RECOVER SESSION") + "\n\n"
	content += styles.Normal.Render("Saved session missing from tmux:") + "\n\n"
	content += line("Name: ", record.Name) + "\n"
	content += line("Tool: ", record.Tool) + "\n"
	content += line("Path: ", record.Path) + "\n\n"
	labels := []string{"Open", "Delete saved entry", "Later"}
	if inner >= 45 {
		content += components.RenderButtons(labels, m.recoveryButton, true, inner)
	} else {
		for i, label := range labels {
			if i == m.recoveryButton {
				content += styles.Selected.Render("> "+label) + "\n"
			} else {
				content += styles.Normal.Render("  "+label) + "\n"
			}
		}
		content = strings.TrimSuffix(content, "\n")
	}
	if m.err != nil {
		content += "\n" + line("Error: ", m.err.Error())
	}
	content += "\n\n" + styles.Dim.Render(fmt.Sprintf("%d saved session(s) to review", len(m.recoverySessions)))
	padding := 1
	if m.height > 0 && m.height < 22 {
		padding = 0
	}
	popup := styles.Panel.Width(width-2).Padding(padding, 2).Render(content)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "tab", Desc: "select"}, {Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "later for all"},
	}, width)
	if m.width > 0 && m.height > 0 {
		popup = lipgloss.Place(m.width, max(0, m.height-lipgloss.Height(help)-1), lipgloss.Center, lipgloss.Center, popup)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}
	return popup + "\n" + help
}
