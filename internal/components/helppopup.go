package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/styles"
)

func RenderHelpPopup() string {
	title := styles.PanelTitle.Render("HELP")

	keyStyle := lipgloss.NewStyle().Foreground(styles.ColorCyan).Width(12)
	descStyle := lipgloss.NewStyle().Foreground(styles.ColorWhite)

	keys := []struct {
		key  string
		desc string
	}{
		{"n", "Create new session"},
		{"enter", "Attach to session"},
		{"d", "Delete session"},
		{"r", "Rename session"},
		{"t", "Switch theme"},
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"?", "Toggle this help"},
		{"q", "Quit"},
	}

	var rows []string
	for _, k := range keys {
		row := keyStyle.Render(k.key) + descStyle.Render(k.desc)
		rows = append(rows, row)
	}

	content := title + "\n\n" + strings.Join(rows, "\n")
	footer := "\n\n" + styles.Dim.Render("hive "+Version+" • press ? or esc to close")

	popupStyle := styles.Panel.
		Width(40).
		Padding(1, 2)

	return popupStyle.Render(content + footer)
}
