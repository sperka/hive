# Hive

A lightweight TUI for managing tmux sessions, optimized for AI coding agents.

This is the [sperka/hive](https://github.com/sperka/hive) fork of
[mdipanjan/hive](https://github.com/mdipanjan/hive). It adds configurable tools,
session renaming and recovery, a compact dashboard with real CPU metrics, and
verified installation and self-updates.

![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey)
![License](https://img.shields.io/badge/License-MIT-green)

Current release: [v0.4.0](https://github.com/sperka/hive/releases/tag/v0.4.0).
See [CHANGELOG.md](CHANGELOG.md) for the complete release history from v0.1.0.

<p align="center">
  <img src="assets/demo.gif" alt="Original Hive dashboard demo" width="700">
</p>

The demo shows the original dashboard; the current layout is described below.

## Why Hive?

Managing multiple AI coding agents (Claude, Pi, Cursor) across projects gets messy fast. Hive solves this.

**Problem:** Juggling tmux sessions manually - creating, naming, finding, switching.

**Solution:** A single interface to manage all your agent sessions.

## Features

| Feature | What it does |
| ------- | ------------ |
| **Compact Dashboard** | Shows sessions, runtime status, tool, path, and real CPU usage |
| **Session Recovery** | Saves session setup and offers to reopen missing sessions at startup |
| **Configurable Tools** | Lists your preferred tools and accepts custom executable names or paths |
| **Session Renaming** | Changes names without interrupting running tools or losing their state |
| **Search and Switcher** | Filters with `/` and switches from inside tmux with `prefix+h` |
| **Editor Workspaces** | Opens Neovim beside a persistent shell pane |
| **CLI + JSON Output** | Lets scripts and AI agents create, list, attach, switch, and delete sessions |
| **Verified Updates** | Installs and updates user-owned binaries with SHA-256 verification |
| **12 Themes** | Saves your theme and displays its name on the dashboard |
| **Single Binary** | Runs on Linux and macOS, on amd64 and arm64 |

## Requirements

- **tmux** - Terminal multiplexer
- The tools you want to launch, installed and executable.
- For the installer: Bash, curl, tar, and either `sha256sum` or `shasum`.

## Installation

**Works on:** macOS, Linux

```bash
curl -fsSL https://github.com/sperka/hive/releases/latest/download/install.sh | bash
```

Then run: `hive`

The installer downloads a release from `sperka/hive`, verifies its SHA-256
checksum, and installs to `~/.local/bin` as your user. Installation and updates
do not use sudo. Rerun the installer to upgrade an older Hive installation that
does not yet have `hive update`.

Put `~/.local/bin` before system directories on your `PATH`. Add this to your
shell configuration if needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Set `HIVE_INSTALL_DIR` to choose another user-writable directory.
To install a particular release instead of the latest stable version:

```bash
curl -fsSL https://github.com/sperka/hive/releases/download/v0.4.0/install.sh |
  HIVE_VERSION=v0.4.0 bash
```

<details>
<summary>Other install methods</summary>

**From Source**

```bash
git clone https://github.com/sperka/hive.git
cd hive
mise install
mise exec -- make build
mkdir -p "$HOME/.local/bin"
install -m 755 hive "$HOME/.local/bin/hive"
```

The Go module retains the upstream import path. Use the checkout above to build
this fork; `go install github.com/mdipanjan/hive@latest` installs upstream Hive.

</details>

## Updating

```bash
hive update --check  # Check the latest stable release
hive update          # Download, verify, and install it
```

Updates come from [sperka/hive releases](https://github.com/sperka/hive/releases).
Starting with v0.3.1, release discovery uses GitHub's public latest-release
redirect. Updates and installation do not require the anonymous REST API or
raw.githubusercontent.com. Both binary and checksum downloads use the same
resolved release tag.
Hive verifies the archive's SHA-256 checksum and atomically replaces its own
executable, including when launched through a symlink. Config and tmux sessions
are preserved. Reopen Hive after updating. Equal or newer installed versions are
left in place; development builds can update to the latest stable release.

The executable's directory must be writable by your user. To migrate an older
installation from `/usr/local/bin`, install a user-owned copy:

```bash
curl -fsSL https://github.com/sperka/hive/releases/latest/download/install.sh |
  HIVE_INSTALL_DIR="$HOME/.local/bin" bash
export PATH="$HOME/.local/bin:$PATH"
hash -r
hive --version
```

Future `hive update` commands update the user-owned copy. The old system copy
can stay in place. Older versions without `hive update` also use the installer
once to gain the command.

If an older updater returns HTTP 403 from `api.github.com`, bootstrap v0.4.0
with the release-hosted installer once:

```bash
curl -fsSL https://github.com/sperka/hive/releases/download/v0.4.0/install.sh |
  HIVE_VERSION=v0.4.0 HIVE_INSTALL_DIR="$HOME/.local/bin" bash
```

Future updates use `hive update`. A 403 from the release
download itself still requires checking that machine's network or proxy access
to GitHub; changing the updater cannot override a network block.

## Usage

The main dashboard is 100 terminal columns wide, including its border. A compact
hexagon/wordmark/status header sits above Sessions and Details, with CPU activity
below. Smaller terminals use a compact layout. CPU activity shows real usage
history, a subdued average/peak summary, and physical/logical core counts.
On Macs that report multiple CPU core types, the summary shows the count of each
type, such as `12 performance + 4 efficiency` or `12P + 4E` in a narrow terminal.
When macOS identifies each logical CPU, its `#` label matches the summary color:
magenta for performance and cyan for efficiency.
Per-logical-CPU entries show a core number and current percentage in eight columns
when space permits, otherwise four. Core numbers run down each column before
continuing in the next. The title, statistics, and grid share one left edge.
Percentages use green below 50%, yellow from 50%, and red from 80%. Startup and
unavailable samples are identified instead of drawn as zero.
The keyboard legend stays directly below the frame, with bright labels, bold
keys, and no background fill. It wraps instead of hiding shortcuts on narrow
terminals. The active theme name appears dimmed in the bottom-right corner.

All terminals, including Ghostty and Kitty, use the original text hexagons by
default. A generated image is available by explicitly enabling Kitty graphics.
Inside tmux, graphics require the existing `allow-passthrough` option to be `on`
or `all`; Hive does not change it. Image data is transmitted directly, so the
optional image also works over SSH.

```bash
HIVE_GRAPHICS=off hive    # Always use the text mark
HIVE_GRAPHICS=kitty hive  # Opt into an image in Ghostty or Kitty
```

The override requires a terminal and, inside tmux, enabled passthrough.

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
| `r`     | Rename session    |
| `/`     | Search sessions   |
| `t`     | Cycle themes      |
| `?`     | Help              |
| `q`     | Quit              |

To rename a session, select it and press `r`. Edit the prefilled name, then press
`enter` to save or `esc` to cancel. The session keeps running, and Hive preserves
its tool, working directory, creation time, and runtime status.

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

This opens a centered switcher popup. Type to filter, use up/down to select, `enter`
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

## Session Recovery

Hive maintains `~/.config/hive/sessions.json` with each observed session's name,
tool, working directory, and creation time. It updates the file when you create,
rename, or delete a session, while the dashboard is open, and when it reads the
session list. Sessions discovered outside Hive are saved too.

When you open Hive, saved sessions that are missing from tmux appear in a recovery
prompt. Choose **Open** to recreate a session, **Delete saved entry** to forget it,
or **Later** to keep it for another startup. Press `esc` to postpone all remaining
sessions. Use `tab` and `enter` to choose an action; `o`, `d`, and `l` are shortcuts.
Deleting a session through Hive also removes its saved entry. Closing one outside
Hive leaves it available for recovery.

Recovery starts a fresh instance of the saved tool in the saved directory using
Hive's normal session setup. It does not restore terminal history, running program
state, command arguments, or custom windows and pane layouts. The directory and
tool must still exist. The compact `hive switch` popup does not show recovery
prompts. Hive does not monitor tmux while it is closed.

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
    graphics/     # Optional Kitty graphics
    lifecycle/    # Session creation, renaming, deletion, and recovery
    logger/       # Logging
    metrics/      # CPU sampling
    runner/       # Tool execution and editor resume behavior
    session/      # Session types
    state/        # Runtime state and saved session catalog
    styles/       # Themes & styles
    testutil/     # Isolated tmux integration tests
    tmux/         # tmux operations
    tui/          # TUI logic
    update/       # Release discovery and verified self-updates
  assets/         # Demo GIF
  docs/           # Website and design notes
  scripts/        # Installer tests
  CHANGELOG.md    # Release history
  RELEASING.md    # Validation, publication, and rollback
```

## Contributing

See [RELEASING.md](RELEASING.md) for versioning, release validation, and rollback.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'feat(sessions): add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.
