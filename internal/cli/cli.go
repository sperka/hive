package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdipanjan/hive/internal/config"
	"github.com/mdipanjan/hive/internal/lifecycle"
	"github.com/mdipanjan/hive/internal/runner"
	"github.com/mdipanjan/hive/internal/styles"
	"github.com/mdipanjan/hive/internal/tui"
)

type SessionOutput struct {
	Name   string `json:"name"`
	Tool   string `json:"tool"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

func Run(args []string) bool {
	if len(args) < 2 {
		return false
	}

	switch args[1] {
	case "list", "ls":
		return runList(args[2:])
	case "create", "new":
		return runCreate(args[2:])
	case "attach", "a":
		return runAttach(args[2:])
	case "delete", "rm":
		return runDelete(args[2:])
	case "switch", "sw", "s":
		return runSwitch(args[2:])
	case "run-session":
		return runSession(args[2:])
	}

	return false
}

func runList(args []string) bool {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	sessions, err := lifecycle.New().List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		output := make([]SessionOutput, len(sessions))
		for i, s := range sessions {
			output[i] = SessionOutput{
				Name:   s.Name,
				Tool:   s.Tool,
				Path:   s.Path,
				Status: s.Status.String(),
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
	} else {
		if len(sessions) == 0 {
			fmt.Println("No sessions")
			return true
		}
		for _, s := range sessions {
			fmt.Printf("%s\t%s\t%s\t%s\n", s.Name, s.Status.String(), s.Tool, s.Path)
		}
	}

	return true
}

func runCreate(args []string) bool {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	tool := fs.String("tool", "bash", "Tool to use (pi, claude, nvim, bash)")
	path := fs.String("path", ".", "Working directory")
	name := fs.String("name", "", "Session name (auto-generated if empty)")
	fs.Parse(args)

	createdName, err := lifecycle.New().Create(lifecycle.CreateRequest{Name: *name, Tool: *tool, Path: *path})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created session: %s\n", createdName)
	return true
}

func runAttach(args []string) bool {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: hive attach <session-name>\n")
		os.Exit(1)
	}

	name := args[0]
	cmd := lifecycle.New().AttachCmd(name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return true
}

func runDelete(args []string) bool {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: hive delete <session-name>\n")
		os.Exit(1)
	}

	name := args[0]
	err := lifecycle.New().Delete(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted session: %s\n", name)
	return true
}

func runSwitch(args []string) bool {
	fs := flag.NewFlagSet("switch", flag.ExitOnError)
	fs.Parse(args)

	cfg := config.Load()
	styles.ApplyTheme(styles.GetThemeByKey(cfg.Theme))

	p := tea.NewProgram(tui.NewSwitch(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return true
}

func runSession(args []string) bool {
	fs := flag.NewFlagSet("run-session", flag.ExitOnError)
	name := fs.String("name", "", "Session name")
	stateID := fs.String("state-id", "", "Stable session state ID (defaults to name)")
	tool := fs.String("tool", "", "Tool to run")
	path := fs.String("path", "", "Working directory")
	fs.Parse(args)

	if err := runner.Run(runner.Options{Name: *name, StateID: *stateID, Tool: *tool, Path: *path}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return true
}
