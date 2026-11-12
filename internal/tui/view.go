package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/components"
	"github.com/mdipanjan/hive/internal/session"
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
	if m.app.RenamingSession() {
		return renderRenameView(m)
	}
	if m.app.CreatingSession() {
		return renderNewView(m)
	}
	if m.app.Searching() {
		return renderSearchView(m)
	}
	return renderListView(m)
}

func renderRenameView(m Model) string {
	content := styles.PanelTitle.Render("RENAME SESSION") + "\n\n"
	content += styles.Label.Render("NAME") + m.renameInput.View()
	if m.err != nil {
		content += "\n\n" + styles.Normal.Render("Error: "+m.err.Error())
	}
	popup := styles.Panel.Width(56).Padding(1, 2).Render(content)
	help := components.RenderHelpBar([]components.HelpItem{
		{Key: "enter", Desc: "rename"},
		{Key: "esc", Desc: "cancel"},
	})
	if m.width > 0 && m.height > 0 {
		popup = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, popup)
		help = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	}
	return popup + "\n" + help
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
	width := 100
	if m.width > 0 {
		width = min(width, m.width)
	}
	if width < 40 || (m.height > 0 && m.height < 12) {
		text := ansi.Truncate("HIVE - enlarge terminal to 40 x 12", width, "")
		return text
	}
	inner := width - 6
	header := components.RenderHeader(m.sessions, inner, m.mark.Render(string(styles.ColorCyan)))
	coreCount := max(m.cpu.LogicalCPUs, len(m.cpu.PerCore))
	columns := components.CPUColumns(inner-4, coreCount)
	cpuRows := max(8, 7+(coreCount+columns-1)/columns)
	if m.height > 0 {
		if m.height < 30 || width < 64 {
			cpuRows = 1
		} else {
			cpuRows = min(cpuRows, max(5, m.height-6-lipgloss.Height(header)-12))
		}
	}
	activity := components.RenderActivity(inner, m.cpu, m.cpuUsageHistory, cpuRows)
	panelRows := 12
	stacked := width < 64
	if m.height > 0 {
		available := m.height - 6 - lipgloss.Height(header) - lipgloss.Height(activity)
		if m.err != nil {
			available--
		}
		if stacked {
			available = (available - 1) / 2
		}
		panelRows = min(panelRows, available)
	}
	if panelRows < 5 {
		return ansi.Truncate("HIVE - enlarge terminal to show sessions", width, "")
	}
	cursor := max(0, min(m.cursor, len(m.sessions)-1))
	var selected *session.Session
	if len(m.sessions) > 0 {
		selected = &m.sessions[cursor]
	}
	var panels string
	if stacked {
		panels = lipgloss.JoinVertical(lipgloss.Left,
			components.RenderSessions(m.sessions, cursor, inner, panelRows), "",
			components.RenderDetails(selected, inner, panelRows))
	} else {
		leftWidth := (inner - 2) / 2
		panels = lipgloss.JoinHorizontal(lipgloss.Top,
			components.RenderSessions(m.sessions, cursor, leftWidth, panelRows), "  ",
			components.RenderDetails(selected, inner-leftWidth-2, panelRows))
	}
	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, panels, activity)
	if m.err != nil {
		// Preserve the frame even when an external command returns a long error.
		message := strings.Join(strings.Fields(ansi.Strip("Error: "+m.err.Error())), " ")
		fullContent += "\n" + styles.Normal.Render(ansi.Truncate(message, inner, "..."))
	}
	// The header includes its own top padding around the mark.
	box := styles.OuterBox.PaddingTop(0).Width(width - 2).Render(fullContent)
	help := components.RenderHelp()
	if lipgloss.Width(help) > width {
		help = styles.Dim.Render("j/k select  enter attach  n new  r rename  ? help  q quit")
	}
	help = ansi.Truncate(help, width, "")
	if m.width > 0 && m.height > 0 {
		// Keep the removed panel gap below the frame so centering does not
		// move the session panels down when the dashboard becomes shorter.
		box = lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center, box+"\n")
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
