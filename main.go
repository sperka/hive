package main

import (
	"fmt"
	"os"

	"github.com/mdipanjan/hive/internal/cli"
	"github.com/mdipanjan/hive/internal/components"
	"github.com/mdipanjan/hive/internal/config"
	"github.com/mdipanjan/hive/internal/styles"
	"github.com/mdipanjan/hive/internal/tui"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("hive", Version)
			return
		case "-h", "--help", "help":
			printHelp()
			return
		case "update":
			if err := cli.Update(os.Args[2:], Version); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			return
		}

		if cli.Run(os.Args) {
			return
		}
	}

	components.Version = Version

	cfg := config.Load()
	theme := styles.GetThemeByKey(cfg.Theme)
	styles.ApplyTheme(theme)

	if err := tui.Run(tui.New()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`hive - lightweight TUI for managing tmux sessions

Usage:
  hive                              Start the TUI
  hive -v, --version                Show version
  hive -h, --help                   Show this help

Commands:
  hive list [--json]                List all sessions
  hive create [options]             Create a new session
      --tool <tool>                 Executable name or path (default: bash)
      --path <path>                 Working directory (default: .)
      --name <name>                 Session name (auto-generated if empty)
  hive attach <name>                Attach to a session
  hive switch                       Open compact session switcher
  hive delete <name>                Delete a session
  hive update [--check]             Update from sperka/hive releases

Aliases:
  list   → ls
  create → new
  attach → a
  switch → sw, s
  delete → rm

TUI Keys:
  n         New session
  enter     Attach to session
  prefix+h  Open switcher from inside tmux
  d         Delete session
  r         Rename session
  /         Search sessions
  t         Switch theme
  ?         Help
  q         Quit`)
}
