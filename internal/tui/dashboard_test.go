package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/components"
	"github.com/mdipanjan/hive/internal/metrics"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/testutil"
)

func dashboardFixture() Model {
	now := time.Now()
	return Model{
		app: NewAppState(),
		sessions: []session.Session{{
			Name: "agent-one", Tool: "codex", Path: "/projects/hive",
			Status: session.StatusRunning, CreatedAt: now, LastActivity: now,
		}},
	}
}

func TestCPUSamplingContinuesInOverlaysAndKeepsRealHistory(t *testing.T) {
	m := dashboardFixture()
	m.app.Search()
	m.searchInput = newSearchInput()
	for i := 0; i < 61; i++ {
		updated, cmd := m.Update(cpuSampleMsg{sample: metrics.CPU{Usage: float64(i), Valid: true}})
		m = updated.(Model)
		if cmd == nil {
			t.Fatal("sampling stopped inside search")
		}
	}
	updated, _ := m.Update(cpuSampleMsg{sample: metrics.CPU{Unavailable: true}})
	m = updated.(Model)
	if len(m.cpuUsageHistory) != 60 || m.cpuUsageHistory[0] != 1 || m.cpuUsageHistory[59] != 60 {
		t.Fatalf("expected the last 60 measured samples: %v", m.cpuUsageHistory)
	}
}

func TestDashboardErrorFitsViewport(t *testing.T) {
	m := dashboardFixture()
	m.width, m.height = 80, 24
	m.err = errors.New(strings.Repeat("a long error\n", 20))
	view := renderListView(m)
	if w, h := lipgloss.Size(view); w > 80 || h > 24 {
		t.Fatalf("error overflowed viewport: %dx%d", w, h)
	}
}

func TestTallDashboardShowsAllCPUMeters(t *testing.T) {
	m := dashboardFixture()
	m.width, m.height = 120, 42
	m.cpu = metrics.CPU{
		Valid: true, PhysicalCores: 32, LogicalCPUs: 32,
		PerCore: make([]float64, 32),
	}
	m.cpuUsageHistory = []float64{0}
	view := ansi.Strip(renderListView(m))
	if !strings.Contains(view, "#32") {
		t.Fatal("a tall dashboard unnecessarily hides CPU meters")
	}
	if lipgloss.Height(view) > 42 {
		t.Fatal("CPU meters overflow the viewport")
	}
}

func TestResizeUpdatesViewportDuringSearch(t *testing.T) {
	m := dashboardFixture()
	m.app.Search()
	m.searchInput = newSearchInput()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	got := updated.(Model)
	if got.width != 80 || got.height != 24 {
		t.Fatalf("search swallowed resize: %dx%d", got.width, got.height)
	}
}

func TestReturnFromAttachKeepsSelectionInRefreshedList(t *testing.T) {
	testutil.Tmux(t)
	testutil.TmuxCommand(t, "new-session", "-d", "-s", "remaining", "/bin/sh")
	m := dashboardFixture()
	m.sessions = []session.Session{{Name: "removed"}, {Name: "remaining"}}
	m.cursor = 1

	updated, _ := m.Update(sessionAttachedMsg{})
	m = updated.(Model)
	if len(m.sessions) != 1 || m.cursor != 0 {
		t.Fatalf("refreshed selection is invalid: cursor=%d, sessions=%v", m.cursor, m.sessions)
	}
	_, attach := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if attach == nil {
		t.Fatal("the displayed selection cannot be attached")
	}
}

func dashboardBox(m Model) string {
	view := renderListView(m)
	box, _, _ := strings.Cut(view, "\n\n")
	return box
}

func TestDashboardWidthIncludesBorders(t *testing.T) {
	for _, width := range []int{0, 160, 100, 80, 60, 40} {
		m := dashboardFixture()
		m.width = width
		want := 100
		if width > 0 && width < want {
			want = width
		}
		if got := lipgloss.Width(dashboardBox(m)); got != want {
			t.Errorf("terminal width %d: box width = %d, want %d", width, got, want)
		}
	}
}

func TestDashboardPanelsAreSideBySide(t *testing.T) {
	m := dashboardFixture()
	lines := strings.Split(ansi.Strip(dashboardBox(m)), "\n")
	found := false
	for _, line := range lines {
		left, right := strings.Index(line, "SESSIONS"), strings.Index(line, "DETAILS")
		if left >= 0 && right > left {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Sessions and Details titles on the same row:\n%s", strings.Join(lines, "\n"))
	}
	logoRow := strings.Split(ansi.Strip(components.RenderLogo()), "\n")[0]
	if !strings.Contains(strings.Join(lines[:6], "\n"), logoRow) {
		t.Fatal("HIVE wordmark missing from the header")
	}
}

func TestDashboardEmptySelectionKeepsDetails(t *testing.T) {
	m := Model{app: NewAppState(), cursor: 99}
	view := ansi.Strip(dashboardBox(m))
	if !strings.Contains(view, "DETAILS") || !strings.Contains(view, "No session selected") {
		t.Fatalf("empty dashboard lost the details panel:\n%s", view)
	}
}

func TestDashboardLongUnicodeFieldsStayWithinFrame(t *testing.T) {
	m := dashboardFixture()
	m.sessions[0].Name = strings.Repeat("\u754c\U0001f680", 30)
	m.sessions[0].Tool = strings.Repeat("very-long-executable-", 10)
	m.sessions[0].Path = "/" + strings.Repeat("\u65e5\u672c\u8a9e/", 30)
	m.cursor = 50
	for _, width := range []int{100, 80, 60, 40} {
		m.width = width
		view := dashboardBox(m)
		if !utf8.ValidString(view) {
			t.Fatalf("invalid UTF-8 at width %d", width)
		}
		if got := lipgloss.Width(view); got != width {
			t.Fatalf("long fields expanded %d-column frame to %d", width, got)
		}
	}
}

func TestDashboardFitsShortViewports(t *testing.T) {
	for _, size := range [][2]int{{100, 24}, {80, 20}, {60, 24}, {40, 24}, {20, 8}} {
		m := dashboardFixture()
		m.width, m.height = size[0], size[1]
		view := renderListView(m)
		if w, h := lipgloss.Size(view); w > m.width || h > m.height {
			t.Errorf("viewport %dx%d: rendered %dx%d", m.width, m.height, w, h)
		}
	}
}

func TestDashboardLegendStaysBelowFrameAcrossViewportSizes(t *testing.T) {
	for _, size := range [][2]int{{160, 50}, {100, 40}, {100, 24}, {80, 20}, {60, 24}, {40, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			m := dashboardFixture()
			m.width, m.height = size[0], size[1]
			view := ansi.Strip(renderListView(m))
			if w, h := lipgloss.Size(view); w > m.width || h > m.height {
				t.Fatalf("legend overflowed viewport: %dx%d", w, h)
			}
			lines := strings.Split(view, "\n")
			frameBottom := -1
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "\u2570") {
					frameBottom = i
				}
			}
			if frameBottom < 0 {
				t.Fatal("dashboard frame missing")
			}
			legend := strings.Join(lines[frameBottom+1:], "\n")
			for _, shortcut := range []string{
				"j/k: select", "enter: attach", "n: new", "d: delete",
				"r: rename", "/: search", "t: theme", "?: help", "q: quit",
			} {
				if !strings.Contains(legend, shortcut) {
					t.Errorf("legend lost %q:\n%s", shortcut, legend)
				}
			}
			firstLegendRow := frameBottom + 1
			for firstLegendRow < len(lines) && strings.TrimSpace(lines[firstLegendRow]) == "" {
				firstLegendRow++
			}
			if firstLegendRow > frameBottom+2 {
				t.Errorf("legend is %d rows below the frame", firstLegendRow-frameBottom)
			}
		})
	}
}
