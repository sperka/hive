package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/components"
	"github.com/mdipanjan/hive/internal/styles"
)

func RenderView(m Model) string {
	if m.app.PickingPath() {
		return renderFilePickerView(m)
	}
	if m.app.ShowingHelp() {
		return renderHelpView(m)
	}
	if m.app.ConfirmingDelete() {
		return renderDeleteConfirmView(m)
	}
	if m.app.CreatingSession() {
		return renderNewView(m)
	}
	if m.app.Searching() {
		return renderSearchView(m)
	}
	return renderListView(m)
}

func renderDeleteConfirmView(m Model) string {
	sessionName := ""
	if len(m.sessions) > 0 {
		sessionName = m.sessions[m.cursor].Name
	}

	popup := components.RenderDeleteConfirm(sessionName, m.deleteButton)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "←→", Desc: "select"},
		{Key: "enter", Desc: "confirm"},
		{Key: "y", Desc: "yes"},
		{Key: "n/esc", Desc: "cancel"},
	})

	if m.width > 0 && m.height > 0 {
		popup = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, popup)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return popup + "\n" + help
}

func renderHelpView(m Model) string {
	popup := components.RenderHelpPopup()
	help := components.RenderHelp()

	if m.width > 0 && m.height > 0 {
		popup = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, popup)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return popup + "\n" + help
}

func renderFilePickerView(m Model) string {
	title := styles.PanelTitle.Render("SELECT DIRECTORY") + "\n\n"
	picker := m.form.FilePicker.View()

	panelStyle := styles.Panel.Width(50).Padding(1, 2)
	panel := panelStyle.Render(title + picker)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "esc", Desc: "cancel"},
	})

	if m.width > 0 && m.height > 0 {
		panel = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, panel)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return panel + "\n" + help
}

func renderSearchView(m Model) string {
	title := "SEARCH"
	action := "attach"
	if m.mode == ModeSwitch {
		title = "SWITCH SESSION"
		action = "switch"
	}

	popup := components.RenderSearchPopupTitled(title, m.searchInput.View(), m.searchInput.Value(), m.sessions, m.searchResults, m.searchCursor)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "↑↓", Desc: "select"},
		{Key: "enter", Desc: action},
		{Key: "esc", Desc: "close"},
	})

	if m.width > 0 && m.height > 0 {
		popup = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, popup)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return popup + "\n" + help
}
func renderListView(m Model) string {
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		components.RenderLogo(),
		"",
		components.RenderHoneycomb(3, 3),
		"",
		components.RenderStats(m.sessions),
	)

	var rightColumn string
	cursor := m.cursor
	if cursor >= len(m.sessions) {
		cursor = max(0, len(m.sessions)-1)
	}

	if len(m.sessions) > 0 {
		rightColumn = lipgloss.JoinVertical(lipgloss.Left,
			components.RenderSessions(m.sessions, cursor),
			components.RenderDetails(m.sessions[cursor]),
		)
	} else {
		rightColumn = components.RenderSessions(m.sessions, cursor)
	}

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)

	contentWidth := maxLineWidth(mainContent)
	boxWidth := contentWidth + 6

	if m.width > 0 && boxWidth > m.width-2 {
		boxWidth = m.width - 2
	}

	activity := components.RenderActivity(contentWidth, m.cpuUsageHistory)
	fullContent := lipgloss.JoinVertical(lipgloss.Right, mainContent, "", activity)

	boxStyle := styles.OuterBox.Width(boxWidth)
	box := boxStyle.Render(fullContent)

	help := components.RenderHelp()
	if m.width > 0 && m.height > 0 {
		box = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, box)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return box + "\n" + help
}

func renderNewView(m Model) string {
	formData := components.FormData{
		Tool:       m.form.Tool,
		CustomTool: m.form.CustomTool.View(),
		Path:       m.form.Path,
		Name:       m.form.Name,
		Focus:      m.form.Focus,
		Button:     m.form.Button,
	}
	if m.err != nil {
		formData.Error = m.err.Error()
	}

	dialog := components.RenderNewDialog(m.form.Tools, formData)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "tab", Desc: "next"},
		{Key: "type", Desc: "custom tool"},
		{Key: "←→", Desc: "select"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "cancel"},
	})

	if m.width > 0 && m.height > 0 {
		dialog = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, dialog)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}

	return dialog + "\n" + help
}

func maxLineWidth(s string) int {
	maxWidth := 0
	for _, line := range strings.Split(s, "\n") {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}
