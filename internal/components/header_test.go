package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mdipanjan/hive/internal/session"
)

func TestHeaderKeepsOriginalHexagonsAlignedWithBrand(t *testing.T) {
	before := Version
	Version = "0.2.0-5-gbd8173b"
	t.Cleanup(func() { Version = before })
	if got := ansi.Strip(RenderLogo()); got != "HIVE" {
		t.Errorf("wordmark = %q, want single-line HIVE", got)
	}
	view := RenderHeader(nil, 94, "")
	lines := strings.Split(ansi.Strip(view), "\n")
	if len(lines) != 5 || lipgloss.Width(view) != 94 {
		t.Fatalf("header should fill 94x5, got %dx%d", lipgloss.Width(view), len(lines))
	}
	if strings.TrimSpace(lines[0]) != "" || strings.TrimSpace(lines[4]) != "" {
		t.Errorf("header lost its top or bottom padding:\n%s", view)
	}
	if !strings.HasSuffix(lines[1], "HIVE v"+Version) {
		t.Errorf("wordmark and full version are not together on their original row: %q", lines[1])
	}
	for row, count := range []int{3, 4, 3} {
		padding := []int{2, 1, 2}[row]
		if strings.Count(lines[row+1], "\u2b21") != count ||
			!strings.HasPrefix(lines[row+1], strings.Repeat(" ", padding)+"\u2b21") {
			t.Errorf("row %d has incorrect hexagon count or padding: %q", row, lines[row+1])
		}
	}
}

func TestHeaderRightAlignsSessionTotalAndStateCounts(t *testing.T) {
	var sessions []session.Session
	for status := session.StatusIdle; status <= session.StatusUnknown; status++ {
		sessions = append(sessions, session.Session{Status: status})
	}
	view := ansi.Strip(RenderHeader(sessions, 94, ""))
	lines := strings.Split(view, "\n")
	if len(lines) != 5 {
		t.Fatalf("header height = %d, want 5", len(lines))
	}
	for _, want := range []string{
		"\u25cf 1 active", "\u25d0 1 running", "\u25cc 1 ready",
		"\u2713 1 done", "\u2715 1 failed", "\u25cb 1 idle", "? 1 unknown",
	} {
		if !strings.Contains(lines[3], want) {
			t.Errorf("missing %q on the bottom hexagon row:\n%s", want, view)
		}
	}
	middleMark := strings.Split(ansi.Strip(RenderHoneycomb(3, 4, 3)), "\n")[1]
	if got := strings.TrimSpace(strings.TrimPrefix(lines[2], middleMark)); got != "7 sessions" {
		t.Errorf("middle row content = %q, want session total", got)
	}
	if !strings.HasSuffix(lines[2], "7 sessions") || !strings.HasSuffix(lines[3], "? 1 unknown") {
		t.Errorf("session total and breakdown are not right-aligned:\n%s", view)
	}
	if strings.Contains(lines[3], "sessions") {
		t.Errorf("session total spilled onto the bottom hexagon row: %q", lines[3])
	}
}

func TestNarrowHeaderDoesNotOverflowWithLongVersionAndManySessions(t *testing.T) {
	before := Version
	Version = "0.2.0-dashboard-development"
	t.Cleanup(func() { Version = before })
	view := RenderHeader(make([]session.Session, 100), 34, "########\n########\n########")
	if got := lipgloss.Width(view); got > 34 {
		t.Fatalf("header width = %d, want at most 34", got)
	}
}
