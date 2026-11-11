package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdipanjan/hive/internal/config"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/logger"
	"github.com/mdipanjan/hive/internal/styles"
)

const (
	FocusTool = iota
	FocusCustomTool
	FocusPath
	FocusName
	FocusButtons
)

func (m Model) updateNewForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	m.err = nil

	switch key {
	case "esc":
		m.app.ReturnToSessionList()
		return m, nil

	case "tab", "down":
		return m.moveFormFocus(1)

	case "shift+tab", "up":
		return m.moveFormFocus(-1)

	case "left":
		switch m.form.Focus {
		case FocusTool:
			m.form.Tool = (m.form.Tool + len(m.form.Tools)) % (len(m.form.Tools) + 1)
		case FocusButtons:
			m.form.Button = 0
		}

	case "right":
		switch m.form.Focus {
		case FocusTool:
			m.form.Tool = (m.form.Tool + 1) % (len(m.form.Tools) + 1)
		case FocusButtons:
			m.form.Button = 1
		}

	case "enter":
		if m.form.Focus == FocusButtons {
			if m.form.Button == 1 {
				m.app.ReturnToSessionList()
				return m, nil
			}
			return m.createSession()
		}
		return m.moveFormFocus(1)

	case "backspace":
		if m.form.Focus == FocusPath && len(m.form.Path) > 0 {
			m.form.Path = m.form.Path[:len(m.form.Path)-1]
		} else if m.form.Focus == FocusName && len(m.form.Name) > 0 {
			m.form.Name = m.form.Name[:len(m.form.Name)-1]
		}

	default:
		if m.form.Focus == FocusTool && (msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace) {
			m.form.Tool = len(m.form.Tools)
			m.form.Focus = FocusCustomTool
			cmd := m.form.CustomTool.Focus()
			var inputCmd tea.Cmd
			m.form.CustomTool, inputCmd = m.form.CustomTool.Update(msg)
			return m, tea.Batch(cmd, inputCmd)
		}
		if len(key) == 1 {
			switch m.form.Focus {
			case FocusPath:
				if key == "b" {
					m.app.PickPath()
					return m, m.form.FilePicker.Init()
				}
				m.form.Path += key
			case FocusName:
				m.form.Name += key
			}
		}
	}

	if m.form.Focus == FocusCustomTool {
		var cmd tea.Cmd
		m.form.CustomTool, cmd = m.form.CustomTool.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) moveFormFocus(direction int) (tea.Model, tea.Cmd) {
	focus := m.form.Focus + direction
	if focus == FocusCustomTool && m.form.Tool < len(m.form.Tools) {
		focus += direction
	}
	m.form.Focus = max(FocusTool, min(FocusButtons, focus))
	m.form.CustomTool.Blur()
	if m.form.Focus == FocusCustomTool {
		return m, m.form.CustomTool.Focus()
	}
	return m, nil
}

func newForm() NewSessionForm {
	input := textinput.New()
	input.Placeholder = "Executable name or path"
	input.Width = 45
	return NewSessionForm{
		Tools:      config.Load().Tool,
		CustomTool: input,
		Path:       getDefaultPath(),
		Focus:      FocusTool,
	}
}

func newFilePicker() filepicker.Model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowPermissions = false
	fp.ShowSize = false
	fp.ShowHidden = false
	fp.Height = 15
	fp.CurrentDirectory, _ = os.UserHomeDir()

	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(styles.ColorCyan)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(styles.ColorCyan)
	fp.Styles.File = lipgloss.NewStyle().Foreground(styles.ColorWhite)
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(styles.ColorGreen).Bold(true)

	return fp
}

func (m Model) createSession() (tea.Model, tea.Cmd) {
	tool := m.form.CustomTool.Value()
	if m.form.Tool < len(m.form.Tools) {
		tool = m.form.Tools[m.form.Tool]
	}
	if strings.TrimSpace(tool) == "" {
		m.err = fmt.Errorf("enter a tool executable name or path")
		return m, nil
	}
	path := m.form.Path
	name := m.form.Name

	logger.Log.Debug("creating session", "name", name, "tool", tool, "path", path)
	createdName, err := lifecycle.New().Create(lifecycle.CreateRequest{Name: name, Tool: tool, Path: path})
	if err != nil {
		m.err = err
		return m, nil
	}
	logger.Log.Debug("created session", "name", createdName)

	m.sessions, _ = lifecycle.New().List()
	m.err = nil
	m.app.ReturnToSessionList()
	return m, nil
}
