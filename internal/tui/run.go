package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Run keeps terminal image resources scoped to this program invocation.
func Run(model Model) error {
	defer func() {
		if cleanup := model.mark.Close(); cleanup != "" {
			fmt.Fprint(os.Stdout, cleanup)
		}
	}()
	_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	return err
}
