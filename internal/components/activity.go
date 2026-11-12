package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/metrics"
	"github.com/mdipanjan/hive/internal/styles"
)

var graphBlocks = []rune{' ', '\u2581', '\u2582', '\u2583', '\u2584', '\u2585', '\u2586', '\u2587', '\u2588'}

// CPUColumns keeps numbered percentages and a gap between cells on each row.
func CPUColumns(width, coreCount int) int {
	digits := len(strconv.Itoa(max(1, coreCount)))
	if width >= 8*(digits+9) {
		return 8
	}
	return 4
}

func RenderActivity(width int, sample metrics.CPU, history []float64, detailRows int) string {
	state := "CPU - collecting data"
	if sample.Unavailable {
		state = "CPU unavailable"
	} else if sample.Valid {
		state = fmt.Sprintf("CPU  %.1f%%", sample.Usage)
	}
	counts := fmt.Sprintf("%d logical CPUs", sample.LogicalCPUs)
	if sample.PhysicalCores > 0 {
		counts = fmt.Sprintf("%d cores / %s", sample.PhysicalCores, counts)
	}
	if sample.LogicalCPUs == 0 {
		counts = "core count unavailable"
	}
	summaryStyle := styles.Normal.Faint(true)
	if detailRows < 5 {
		return summaryStyle.Render(cellText(state+"  |  "+counts, width))
	}
	inner := width - 4
	gap := max(1, inner-lipgloss.Width(state)-lipgloss.Width(counts))
	title := state + strings.Repeat(" ", gap) + counts
	var lines []string
	if !sample.Valid {
		message := "Collecting real CPU counters..."
		if sample.Unavailable {
			message = "CPU counters unavailable; waiting for the next sample."
		}
		lines = append(lines, styles.Normal.Render(cellText(message, inner)))
		return panel(title, lines, width, detailRows)
	}
	if len(history) > 0 {
		var sum, peak float64
		for _, usage := range history {
			sum += usage
			peak = math.Max(peak, usage)
		}
		scale := 1.0
		for scale < peak {
			scale *= 10
		}
		lines = append(lines, usageGraph(history, inner, 2, scale)...)
		lines = append(lines, summaryStyle.Render(cellText(
			fmt.Sprintf("avg %.1f%%   peak %.1f%%   scale 0-%.0f%%   %d samples / 1s",
				sum/float64(len(history)), peak, scale, len(history)), inner)))
	}
	available := max(0, detailRows-4-len(lines))
	if available == 0 || len(sample.PerCore) == 0 {
		return panel(title, lines, width, detailRows)
	}
	coreCount := max(sample.LogicalCPUs, len(sample.PerCore))
	columns := min(CPUColumns(inner, coreCount), len(sample.PerCore))
	rows := min(available, (len(sample.PerCore)+columns-1)/columns)
	visible := min(len(sample.PerCore), rows*columns)
	shortRows, tallerColumns := visible/columns, visible%columns
	digits := len(strconv.Itoa(max(1, coreCount)))
	// Share the panel's left edge and spread spare width between columns.
	// Reserve room for 100.0% so readings cannot change their positions.
	entryWidth := min(digits+8, max(1, inner/columns-1))
	for row := 0; row < rows; row++ {
		var cells []string
		for column := 0; column < columns; column++ {
			columnRows := shortRows
			if column < tallerColumns {
				columnRows++
			}
			i := column*shortRows + min(column, tallerColumns) + row
			text := ""
			if row < columnRows {
				if i == visible-1 && len(sample.PerCore) > visible {
					text = styles.Normal.Render(fmt.Sprintf("+%d more", len(sample.PerCore)-i))
				} else {
					text = styles.Normal.Render(fmt.Sprintf("#%-*d ", digits, i+1))
					if math.IsNaN(sample.PerCore[i]) {
						text += styles.Normal.Render("   --")
					} else {
						value := math.Max(0, math.Min(100, sample.PerCore[i]))
						color := styles.ColorGreen
						if value >= 80 {
							color = lipgloss.Color("9")
						} else if value >= 50 {
							color = styles.ColorYellow
						}
						text += styles.Normal.Foreground(color).Bold(true).Render(fmt.Sprintf("%4.1f%%", value))
					}
				}
			}
			start := column * (inner - entryWidth) / max(1, columns-1)
			end := inner
			if column < columns-1 {
				end = (column + 1) * (inner - entryWidth) / (columns - 1)
			}
			cells = append(cells, lipgloss.NewStyle().Width(end-start).
				Render(ansi.Truncate(text, entryWidth, "")))
		}
		lines = append(lines, strings.Join(cells, ""))
	}
	return panel(title, lines, width, detailRows)
}

func usageGraph(history []float64, width, height int, scale float64) []string {
	values := history[max(0, len(history)-width):]
	var lines []string
	for row := 0; row < height; row++ {
		var line strings.Builder
		line.WriteString(strings.Repeat(" ", max(0, width-len(values))))
		for _, value := range values {
			level := int(math.Round(math.Max(0, math.Min(scale, value)) / scale * float64(height*8)))
			level = max(0, min(8, level-(height-1-row)*8))
			line.WriteRune(graphBlocks[level])
		}
		lines = append(lines, styles.Logo.Render(line.String()))
	}
	return lines
}
