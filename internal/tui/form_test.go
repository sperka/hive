package tui

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/config"
)

func formModel(t *testing.T, tools []string) Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := config.Save(config.Config{Theme: "nord", Tool: tools}); err != nil {
		t.Fatal(err)
	}
	m := Model{app: NewAppState()}
	model, _ := m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	return model.(Model)
}

func press(m Model, key tea.KeyType) Model {
	model, _ := m.Update(tea.KeyMsg{Type: key})
	return model.(Model)
}

func TestFormUsesConfiguredToolsAndCustomInput(t *testing.T) {
	m := formModel(t, []string{"zsh", "custom-agent"})
	if !reflect.DeepEqual(m.form.Tools, []string{"zsh", "custom-agent"}) {
		t.Fatalf("tools = %#v", m.form.Tools)
	}
	m = press(m, tea.KeyRight)
	if m.form.Tool != 1 {
		t.Fatal("right did not select second tool")
	}
	m = press(m, tea.KeyTab)
	if m.form.Focus != FocusPath {
		t.Fatal("preset should skip custom input")
	}
	m = press(m, tea.KeyShiftTab)
	m = press(m, tea.KeyRight)
	m = press(m, tea.KeyEnter)
	if m.form.Focus != FocusCustomTool || !m.form.CustomTool.Focused() {
		t.Fatal("Custom should focus editable input")
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/opt/my agent"), Paste: true})
	m = model.(Model)
	m = press(m, tea.KeyLeft)
	m = press(m, tea.KeyBackspace)
	if got := m.form.CustomTool.Value(); got != "/opt/my aget" {
		t.Fatalf("custom input editing = %q", got)
	}
	m = press(m, tea.KeyTab)
	m = press(m, tea.KeyShiftTab)
	if m.form.Focus != FocusCustomTool {
		t.Fatal("backward navigation skipped custom input")
	}
}

func TestTypingSelectsCustomTool(t *testing.T) {
	m := formModel(t, []string{"pi"})
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("agent")})
	m = model.(Model)
	if m.form.Tool != len(m.form.Tools) || m.form.CustomTool.Value() != "agent" {
		t.Fatal("typing should select and populate custom input")
	}
	m = press(m, tea.KeyEsc)
	if m.app.CreatingSession() {
		t.Fatal("escape did not cancel form")
	}
}

func TestEmptyToolListAndBlankCustomTool(t *testing.T) {
	m := formModel(t, []string{})
	m = press(m, tea.KeyLeft)
	m = press(m, tea.KeyRight)
	m = press(m, tea.KeyTab)
	if m.form.Focus != FocusCustomTool {
		t.Fatal("empty list should offer custom input")
	}
	m.form.Focus = FocusButtons
	m = press(m, tea.KeyEnter)
	if m.err == nil || !strings.Contains(m.View(), "enter a tool") || !m.app.CreatingSession() {
		t.Fatal("blank custom tool should show an error and keep the form open")
	}
}

func TestThemeChangePreservesConfiguredTools(t *testing.T) {
	m := formModel(t, []string{"my-agent", "zsh"})
	m.app.ReturnToSessionList()
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if got := config.Load().Tool; !reflect.DeepEqual(got, []string{"my-agent", "zsh"}) {
		t.Fatalf("theme change replaced tools: %#v", got)
	}
}
