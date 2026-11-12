package tui

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/graphics"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/logger"
	"github.com/mdipanjan/hive/internal/metrics"
	"github.com/mdipanjan/hive/internal/session"
	"golang.org/x/term"
)

type Mode string

const (
	ModeDashboard Mode = "dashboard"
	ModeSwitch    Mode = "switch"
)

type Model struct {
	mode            Mode
	width           int
	height          int
	sessions        []session.Session
	cursor          int
	app             AppState
	form            NewSessionForm
	searchInput     textinput.Model
	searchResults   []int
	searchCursor    int
	deleteButton    int
	renameInput     textinput.Model
	renameName      string
	cpuUsageHistory []float64
	cpu             metrics.CPU
	cpuSampler      *metrics.Sampler
	mark            *graphics.Mark
	err             error
}

type NewSessionForm struct {
	Tool       int
	Tools      []string
	CustomTool textinput.Model
	FilePicker filepicker.Model
	Path       string
	Name       string
	Focus      int
	Button     int
}

type cpuTickMsg time.Time
type cpuSampleMsg struct{ sample metrics.CPU }
type sessionAttachedMsg struct{ err error }

func New() Model {
	return newModel(ModeDashboard)
}

func NewSwitch() Model {
	return newModel(ModeSwitch)
}

func newModel(mode Mode) Model {
	logger.Log.Debug("initializing model", "mode", mode)
	sessions, err := lifecycle.New().List()
	if err != nil {
		logger.Log.Error("failed to list sessions", "err", err)
	}
	logger.Log.Debug("found sessions", "count", len(sessions))

	m := Model{
		mode:     mode,
		sessions: sessions,
		cursor:   0,
		app:      NewAppState(),
		err:      err,
	}
	if mode == ModeDashboard {
		m.cpuSampler = metrics.NewSampler()
		m.mark = graphics.NewMark()
	}

	if w, h, sizeErr := term.GetSize(int(os.Stdout.Fd())); sizeErr == nil && w > 0 && h > 0 {
		m.width = w
		m.height = h
	}

	if mode == ModeSwitch {
		m.searchInput = newSearchInput()
		m.app.Search()
		m.searchResults = m.getIndices("")
		m.searchCursor = 0
	}

	return m
}

func (m Model) Init() tea.Cmd {
	if m.mode == ModeSwitch {
		return textinput.Blink
	}
	return m.sampleCPU()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Viewport and sampling messages must not be swallowed by overlay inputs.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case cpuTickMsg:
		return m, m.sampleCPU()
	case cpuSampleMsg:
		m.cpu = msg.sample
		if msg.sample.Valid {
			m.cpuUsageHistory = append(m.cpuUsageHistory, msg.sample.Usage)
			if len(m.cpuUsageHistory) > 60 {
				m.cpuUsageHistory = m.cpuUsageHistory[len(m.cpuUsageHistory)-60:]
			}
		}
		return m, cpuTick()
	}
	if m.app.PickingPath() {
		return m.updateFilePicker(msg)
	}

	if m.app.Searching() {
		if _, ok := msg.(tea.KeyMsg); !ok {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case sessionAttachedMsg:
		logger.Log.Debug("returned from attach", "err", msg.err)
		if m.mode == ModeSwitch {
			m.err = msg.err
			if msg.err != nil {
				return m, nil
			}
			return m, tea.Quit
		}
		sessions, _ := lifecycle.New().List()
		m.sessions = sessions
		m.cursor = max(0, min(m.cursor, len(sessions)-1))
		return m, tea.ClearScreen

	case tea.KeyMsg:
		if m.app.Searching() {
			return m.updateSearch(msg)
		}
		if m.app.ShowingHelp() {
			if msg.String() == "?" || msg.String() == "esc" {
				m.app.CloseOverlay()
			}
			return m, nil
		}
		if m.app.ConfirmingDelete() {
			return m.updateDeleteConfirm(msg)
		}
		if m.app.RenamingSession() {
			return m.updateRename(msg)
		}
		if m.app.CreatingSession() {
			return m.updateNewForm(msg)
		}
		return m.updateList(msg)

	}
	if m.app.CreatingSession() && m.form.Focus == FocusCustomTool {
		var cmd tea.Cmd
		m.form.CustomTool, cmd = m.form.CustomTool.Update(msg)
		return m, cmd
	}
	if m.app.RenamingSession() {
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	return RenderView(m)
}
