package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/styles"
)

// cellText bounds terminal text without splitting UTF-8 or wide characters.
func cellText(text string, width int) string {
	text = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, ansi.Strip(text))
	tail := "..."
	if width < 4 {
		tail = ""
	}
	return ansi.Truncate(text, max(0, width), tail)
}

// panel's dimensions include its border and padding.
func panel(title string, lines []string, width, height int) string {
	inner := max(1, width-4)
	content := []string{styles.PanelTitle.Render(cellText(title, inner)), ""}
	for _, line := range lines {
		if len(content) >= height-2 {
			break
		}
		content = append(content, ansi.Truncate(line, inner, ""))
	}
	return styles.Panel.Padding(0, 1).
		Width(width - 2).Height(max(1, height-2)).
		Render(strings.Join(content, "\n"))
}
