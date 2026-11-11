# Hive

A lightweight TUI for managing tmux sessions, optimized for AI coding agents.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey)
![License](https://img.shields.io/badge/License-MIT-green)

<p align="center">
  <img src="assets/demo.gif" alt="Hive Demot " width="700">
</p>

## Why Hive?

Managing multiple AI coding agents (Claude, Pi, Cursor) across projects gets messy fast. Hive solves this.

**Problem:** Juggling tmux sessions manually - creating, naming, finding, switching.

**Solution:** A single interface to manage all your agent sessions.

## Features

| Feature               | What it solves                                        |
| --------------------- | ----------------------------------------------------- |
| **Unified Dashboard** | See all sessions at a glance - status, tool, path     |
| **Quick Search**      | Find sessions instantly with `/` - no more `tmux ls`  |
| **Instant Switcher**  | Jump between sessions with `prefix+h` - no detach      |
| **CLI + JSON Output** | AI agents can create/manage sessions programmatically |
| **12 Themes**         | Match your terminal aesthetic                         |
| **Lightweight**       | ~4MB binary, instant startup, no runtime deps         |

## Requirements

- **tmux** - Terminal multiplexer

## Installation

**Works on:** macOS, Linux

```bash
curl -fsSL https://raw.githubusercontent.com/mdipanjan/hive/main/install.sh | bash
```

Then run: `hive`

<details>
<summary>Other install methods</summary>

**From Source**

```bash
git clone https://github.com/mdipanjan/hive.git
cd hive
make build
sudo mv hive /usr/local/bin/
```

**Go**

```bash
go install github.com/mdipanjan/hive@latest
```

</details>

## Usage

### TUI Mode

```bash
hive
```

**Keyboard Shortcuts:**

| Key     | Action            |
| ------- | ----------------- |
| `n`     | New session       |
| `enter` | Attach to session |
| `prefix + h` | Open session switcher (from inside any session) |
| `d`     | Delete session    |
| `/`     | Search sessions   |
| `t`     | Cycle themes      |
| `?`     | Help              |
| `q`     | Quit              |

### CLI Mode (AI-Agent Compatible)

```bash
# List sessions
hive list              # Plain text
hive list --json       # JSON output

# Create session
hive create --tool pi --path /projects/myapp --name my-session

# Create Neovim + terminal split session
hive create --tool nvim --path /projects/myapp --name my-editor

# In nvim sessions:
# - left pane opens Neovim
# - right pane stays as a shell
# - prefix+h opens the Hive session switcher popup
# - quitting Neovim detaches from tmux, but keeps the session alive
# - attaching again opens Neovim again with the right shell preserved

# Attach to session
hive attach my-session

# Open compact session switcher
hive switch

# Delete session
hive delete my-session
```

**Aliases:**

- `list` / `ls`
- `create` / `new`
- `attach` / `a`
- `switch` / `sw` / `s`
- `delete` / `rm`

**JSON Output Example:**

```json
[
  {
    "name": "hive-abc123",
    "tool": "pi",
    "path": "/projects/myapp",
    "status": "running"
  }
]
```

## Session Switcher

Switch between sessions without leaving the one you're in.

From inside any Hive-created tmux session, press:

```
prefix + h   (e.g. Ctrl+b h)
```

This opens a centered switcher popup. Type to filter, `↑↓` to select, `enter`
to switch, `esc` to close. Because Hive uses `tmux switch-client` under the
hood, switching is instant - no detach/reattach cycle.

You can also open it directly:

```bash
hive switch
```

> The `prefix + h` binding is installed automatically when Hive creates a
> session. To add it to existing sessions, run:
>
> ```bash
> tmux bind-key h display-popup -B -w 90% -h 80% -E "hive switch"
> ```

## Configuration

Config file: `~/.config/hive/config.json`

```json
{
  "theme": "tokyo-night",
  "tool": ["pi", "claude", "codex", "zsh"]
}
```

Hive creates this file with defaults if it does not exist. Existing configs without
`tool` use the default choices above. Edit the array to set the choices and their
order; an empty array leaves only custom input. The first choice is selected when
you open the new-session form.

Press `n`, then use left/right to choose a tool. Select `Custom...` and press `tab`,
or just start typing in the tool selector, to enter any executable name or path.
The tool must be installed and executable. Arguments and shell expressions are
not supported in this field. The CLI's `--tool` flag accepts the same executables;
`nvim` and `vim` retain the editor split layout.

## Themes

Press `t` to cycle through themes:

- Tokyo Night (default)
- Tokyo Storm
- Dracula
- Nord
- Gruvbox
- Catppuccin
- One Dark
- Solarized Dark
- GitHub Dark
- Rose Pine
- Monokai
- Zinc Dark

## Debug Mode

```bash
DEBUG=1 hive
```

Logs are written to `debug.log`.

## Project Structure

```
hive/
  main.go
  internal/
    cli/          # CLI commands
    components/   # UI components
    config/       # Configuration
    logger/       # Logging
    session/      # Session types
    styles/       # Themes & styles
    tmux/         # tmux operations
    tui/          # TUI logic
  assets/         # Demo GIF
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.
