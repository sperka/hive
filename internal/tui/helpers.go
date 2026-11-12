package tui

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/styles"
)

func newSearchInput() textinput.Model {
	in := textinput.New()
	in.Placeholder = "Search..."
	in.Focus()
	in.CharLimit = 30
	in.Width = 30
	in.PromptStyle = lipgloss.NewStyle().Foreground(styles.ColorCyan)
	in.TextStyle = lipgloss.NewStyle().Foreground(styles.ColorWhite)
	return in
}

func getDefaultPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return "~"
	}
	return dir
}

func cpuTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return cpuTickMsg(t)
	})
}

func (m Model) sampleCPU() tea.Cmd {
	if m.cpuSampler == nil {
		return nil
	}
	return func() tea.Msg {
		sample, _ := m.cpuSampler.Sample()
		return cpuSampleMsg{sample: sample}
	}
}
