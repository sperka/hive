package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/styles"
)

type HelpItem struct {
	Key  string
	Desc string
}

func RenderHelpBar(items []HelpItem, widths ...int) string {
	var parts []string
	for _, item := range items {
		part := styles.HelpKey.Render(item.Key) + styles.HelpDesc.Render(": "+item.Desc)
		parts = append(parts, part)
	}
	if len(widths) > 0 && widths[0] > 0 {
		helpStyle := styles.Help.Width(widths[0])
		contentWidth := widths[0] - helpStyle.GetHorizontalFrameSize()
		separator := styles.HelpDesc.Render("  ")
		var rows []string
		line := ""
		for _, part := range parts {
			if line == "" {
				line = part
			} else if lipgloss.Width(line+separator+part) <= contentWidth {
				line += separator + part
			} else {
				rows = append(rows, line)
				line = part
			}
		}
		if line != "" {
			rows = append(rows, line)
		}
		return helpStyle.Render(strings.Join(rows, "\n"))
	}
	return styles.Help.Render(strings.Join(parts, styles.HelpDesc.Render("   ")))
}

func RenderButtons(labels []string, selected int, focused bool, width int) string {
	baseStyle := lipgloss.NewStyle().
		Foreground(styles.ColorWhite).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorGray)

	var buttons []string
	for i, label := range labels {
		btnStyle := baseStyle
		if focused && i == selected {
			btnStyle = btnStyle.
				Background(styles.ColorCyan).
				BorderForeground(styles.ColorCyan).
				Bold(true)
		} else if i == selected {
			btnStyle = btnStyle.
				Background(styles.ColorGray).
				BorderForeground(styles.ColorWhite).
				Bold(true)
		}
		buttons = append(buttons, btnStyle.Render("  "+label+"  "))
	}

	joined := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, joined)
}

func RenderTextInput(value string, focused bool, width int) string {
	displayValue := value

	if len(displayValue) < width {
		displayValue = displayValue + strings.Repeat(" ", width-len(displayValue))
	} else if len(displayValue) > width {
		displayValue = displayValue[:width]
	}

	inputStyle := lipgloss.NewStyle().Foreground(styles.ColorWhite)

	if focused {
		inputStyle = inputStyle.Background(styles.ColorDim).Bold(true)
		if len(value) < width {
			displayValue = value + "█" + strings.Repeat(" ", width-len(value)-1)
		}
	}

	text := inputStyle.Render(displayValue)
	underline := styles.Dim.Render(strings.Repeat("─", width))

	return text + "\n" + underline
}
