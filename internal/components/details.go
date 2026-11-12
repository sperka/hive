package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/styles"
)

const (
	maxPathLength = 20
	labelWidth    = 10
	valueWidth    = 18
)

func RenderDetails(s *session.Session, width, rows int) string {
	if s == nil {
		return panel("DETAILS", []string{cellText("No session selected", width-4)}, width, rows)
	}
	values := [][2]string{
		{"Name", s.Name}, {"Tool", s.Tool}, {"Path", CollapsePath(s.Path)},
		{"Status", GetStatusText(s.Status)},
		{"Created", humanize.Time(s.CreatedAt)}, {"Activity", humanize.Time(s.LastActivity)},
	}
	labelSize := min(10, max(1, (width-4)/3))
	var lines []string
	for _, row := range values {
		label := styles.Label.Width(labelSize).Render(cellText(row[0], labelSize-1))
		value := styles.Value.Render(cellText(row[1], width-4-labelSize))
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, label, value))
	}
	return panel("DETAILS", lines, width, rows)
}
