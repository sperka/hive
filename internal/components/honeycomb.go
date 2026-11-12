package components

import (
	"strings"

	"github.com/mdipanjan/hive/internal/styles"
)

func RenderHoneycomb(rowCounts ...int) string {
	maxCount := 0
	for _, count := range rowCounts {
		maxCount = max(maxCount, count)
	}
	var lines []string

	for _, count := range rowCounts {
		line := strings.Repeat(" ", maxCount-count+1) + strings.Repeat("\u2b21 ", count)
		lines = append(lines, strings.TrimRight(line, " "))
	}

	return styles.Logo.Render(strings.Join(lines, "\n"))
}
