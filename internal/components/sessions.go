package components

import (
	"fmt"

	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/styles"
)

const (
	maxNameLength = 20
)

func RenderSessions(sessions []session.Session, cursor, width, rows int) string {
	if len(sessions) == 0 {
		return panel("SESSIONS", []string{
			cellText("No sessions. Press n to create one.", width-4),
		}, width, rows)
	}
	cursor = max(0, min(cursor, len(sessions)-1))
	visible := max(1, rows-5)
	start, end := calculateWindow(len(sessions), cursor, visible)
	var lines []string
	for index := start; index < end; index++ {
		sessionItem := sessions[index]
		name := cellText(sessionItem.Name, width-6)
		if cursor == index {
			lines = append(lines, styles.Logo.Render("> "+name))
		} else {
			lines = append(lines, GetStatusIcon(sessionItem.Status)+" "+styles.Normal.Render(name))
		}
	}
	lines = append(lines, styles.Dim.Render(fmt.Sprintf("%d-%d / %d", start+1, end, len(sessions))))
	return panel("SESSIONS", lines, width, rows)
}

func calculateWindow(total, cursor, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 0, total
	}

	half := maxVisible / 2
	start := cursor - half
	end := cursor + half + (maxVisible % 2)

	if start < 0 {
		start = 0
		end = maxVisible
	}
	if end > total {
		end = total
		start = total - maxVisible
	}

	return start, end
}
