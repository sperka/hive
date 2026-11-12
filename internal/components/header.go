package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/styles"
)

func RenderHeader(sessions []session.Session, width int, mark string) string {
	if mark == "" {
		mark = RenderHoneycomb(3, 4, 3)
	}
	mark = lipgloss.NewStyle().Padding(1, 0).Render(mark)
	remaining := max(0, width-lipgloss.Width(mark)-2)
	brand := RenderLogo()
	counts := make(map[session.Status]int)
	for _, s := range sessions {
		counts[s.Status]++
	}
	version := Version
	if version != "dev" {
		version = "v" + strings.TrimPrefix(version, "v")
	}
	version = styles.Logo.Render(cellText(version, max(0, remaining-lipgloss.Width(brand)-1)))
	top := lipgloss.PlaceHorizontal(remaining, lipgloss.Right, brand+" "+version)

	total := styles.Normal.Render(fmt.Sprintf("%d sessions", len(sessions)))
	var full, compact []string
	for _, state := range []session.Status{
		session.StatusActive, session.StatusRunning, session.StatusReady,
		session.StatusCompleted, session.StatusFailed, session.StatusIdle, session.StatusUnknown,
	} {
		icon, label := GetStatusIcon(state), state.String()
		if state == session.StatusCompleted {
			label = "done"
		}
		if state == session.StatusUnknown {
			icon = styles.Normal.Render("?")
		}
		number := fmt.Sprintf("%d", counts[state])
		full = append(full, icon+" "+styles.Normal.Render(number+" "+label))
		compact = append(compact, icon+styles.Normal.Render(number))
	}
	bottom := strings.Join(full, " ")
	if lipgloss.Width(bottom) > remaining {
		bottom = strings.Join(compact, " ")
	}
	right := strings.Join([]string{
		"",
		top,
		lipgloss.PlaceHorizontal(remaining, lipgloss.Right, ansi.Truncate(total, remaining, "")),
		lipgloss.PlaceHorizontal(remaining, lipgloss.Right, ansi.Truncate(bottom, remaining, "")),
	}, "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, mark, "  ", right)
}
