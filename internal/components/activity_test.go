package components

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/metrics"
)

func TestActivityShowsCoreCountsAndHistorySummary(t *testing.T) {
	sample := metrics.CPU{
		Usage: 25, Valid: true, PhysicalCores: 2, LogicalCPUs: 4,
		PerCore: []float64{0, 25, 50, 25},
	}
	view := RenderActivity(94, sample, []float64{10, 40, 25}, 12)
	text := ansi.Strip(view)
	for _, want := range []string{"25.0%", "2 cores", "4 logical CPUs", "avg 25.0%", "peak 40.0%", "#1"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if lipgloss.Width(view) != 94 || lipgloss.Height(view) > 12 {
		t.Fatalf("CPU panel exceeded its allocated space: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
}

func TestCoreGridUsesFourOrEightColumnsOfNumberedPercentages(t *testing.T) {
	for _, tt := range []struct {
		width, cores, columns int
	}{{94, 32, 8}, {74, 32, 4}, {94, 128, 4}} {
		t.Run(fmt.Sprintf("%d-columns-%d-cores", tt.width, tt.cores), func(t *testing.T) {
			sample := metrics.CPU{Valid: true, LogicalCPUs: tt.cores, PerCore: make([]float64, tt.cores)}
			for i := range sample.PerCore {
				sample.PerCore[i] = float64(i % 101)
			}
			sample.PerCore[0] = math.NaN()
			rows := 7 + tt.cores/tt.columns
			view := ansi.Strip(RenderActivity(tt.width, sample, []float64{20}, rows))
			coreRows := 0
			for _, line := range strings.Split(view, "\n") {
				if strings.Contains(line, "#") {
					coreRows++
					if got := strings.Count(line, "#"); got != tt.columns {
						t.Errorf("core row has %d entries, want %d: %s", got, tt.columns, line)
					}
					if strings.ContainsAny(line, "\u2588\u2591") || strings.Contains(line, "CPU") {
						t.Errorf("core row still contains bars or repeated CPU labels: %s", line)
					}
				}
			}
			if coreRows != tt.cores/tt.columns || !strings.Contains(view, "--") ||
				!strings.Contains(view, fmt.Sprintf("#%d", tt.cores)) {
				t.Errorf("missing numbered cores or unavailable marker:\n%s", view)
			}
			if tt.cores > 100 && !strings.Contains(view, "100.0%") {
				t.Errorf("a fully utilized core lost its percentage:\n%s", view)
			}
			if w, h := lipgloss.Size(view); w != tt.width || h != rows {
				t.Errorf("grid size = %dx%d, want %dx%d", w, h, tt.width, rows)
			}
		})
	}
}

func TestCoreGridReadsDownColumns(t *testing.T) {
	for _, tt := range []struct {
		name         string
		width, cores int
		height       int
		want         [][]int
		hidden       int
	}{
		{
			name: "eight full columns", width: 94, cores: 32, height: 11,
			want: [][]int{
				{1, 5, 9, 13, 17, 21, 25, 29},
				{2, 6, 10, 14, 18, 22, 26, 30},
				{3, 7, 11, 15, 19, 23, 27, 31},
				{4, 8, 12, 16, 20, 24, 28, 32},
			},
		},
		{
			name: "uneven columns", width: 74, cores: 10, height: 10,
			want: [][]int{{1, 4, 7, 9}, {2, 5, 8, 10}, {3, 6}},
		},
		{
			name: "limited height", width: 94, cores: 32, height: 9,
			want: [][]int{
				{1, 3, 5, 7, 9, 11, 13, 15},
				{2, 4, 6, 8, 10, 12, 14},
			},
			hidden: 17,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sample := metrics.CPU{Valid: true, LogicalCPUs: tt.cores, PerCore: make([]float64, tt.cores)}
			for i := range sample.PerCore {
				sample.PerCore[i] = float64(i + 1)
			}
			view := ansi.Strip(RenderActivity(tt.width, sample, []float64{20}, tt.height))
			entries := regexp.MustCompile(`#([0-9]+) +([0-9.]+)%`)
			row := 0
			for _, line := range strings.Split(view, "\n") {
				matches := entries.FindAllStringSubmatch(line, -1)
				if len(matches) == 0 {
					continue
				}
				if row >= len(tt.want) || len(matches) != len(tt.want[row]) {
					t.Fatalf("unexpected core row %d: %s", row, line)
				}
				for column, core := range tt.want[row] {
					if matches[column][1] != fmt.Sprint(core) ||
						matches[column][2] != fmt.Sprintf("%.1f", float64(core)) {
						t.Errorf("row %d column %d = %v, want core %d with its own usage",
							row, column, matches[column][1:], core)
					}
				}
				row++
			}
			if row != len(tt.want) {
				t.Errorf("got %d core rows, want %d", row, len(tt.want))
			}
			if tt.hidden > 0 && !strings.Contains(view, fmt.Sprintf("+%d more", tt.hidden)) {
				t.Errorf("missing hidden-core count:\n%s", view)
			}
		})
	}
}

func TestActivityNeverInventsStartupOrFailedUsage(t *testing.T) {
	for _, sample := range []metrics.CPU{
		{PhysicalCores: 2, LogicalCPUs: 4},
		{PhysicalCores: 2, LogicalCPUs: 4, Unavailable: true},
	} {
		text := ansi.Strip(RenderActivity(94, sample, nil, 10))
		if strings.Contains(text, "%") {
			t.Fatalf("displayed a fabricated percentage: %s", text)
		}
		if sample.Unavailable && !strings.Contains(text, "unavailable") {
			t.Fatalf("failed sample not identified: %s", text)
		}
	}
}

func TestCompactActivityRetainsCounts(t *testing.T) {
	view := RenderActivity(74, metrics.CPU{Valid: true, Usage: 12.5, LogicalCPUs: 8}, nil, 1)
	if lipgloss.Height(view) != 1 || lipgloss.Width(view) > 74 ||
		!strings.Contains(ansi.Strip(view), "8 logical CPUs") {
		t.Fatalf("compact CPU summary lost counts or overflowed: %q", view)
	}
}

func TestActivityHistoryMakesLowUsageVisibleWithLabeledScale(t *testing.T) {
	view := ansi.Strip(RenderActivity(94, metrics.CPU{Valid: true, Usage: 0.4, LogicalCPUs: 1},
		[]float64{0.1, 0.2, 0.4}, 8))
	if !strings.Contains(view, "scale 0-1%") || !strings.ContainsAny(view, "\u2581\u2582\u2583\u2584\u2585\u2586\u2587\u2588") {
		t.Fatalf("low measured usage vanished from the chart:\n%s", view)
	}
}
