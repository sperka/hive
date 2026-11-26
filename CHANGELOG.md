# Changelog

Release history from the v0.1.0 baseline through v0.4.1, reconstructed from the
repository's commits and release tags. Versions 0.1.0-0.1.2 are inherited from
[upstream Hive](https://github.com/mdipanjan/hive). Fork-specific development
first shipped in 0.2.0, together with upstream's subsequent lifecycle and
session-switcher work. Each comparison links to the full set of changes.

## 0.4.1

[Changes since 0.4.0](https://github.com/sperka/hive/compare/v0.4.0...v0.4.1)

- Restore live CPU usage and per-core percentages in macOS release builds.
- Show performance and efficiency core counts on Macs that report both types.
- Match each verified core number to its summary label using magenta for performance
  and cyan for efficiency. Load percentages keep their green, yellow, and red scale.

## 0.4.0

[Changes since 0.3.2](https://github.com/sperka/hive/compare/v0.3.2...v0.4.0)

### Session recovery

- Save each observed session's name, tool, working directory, and creation time
  in `~/.config/hive/sessions.json`, including sessions created outside Hive.
- Offer to reopen missing sessions at startup, delete their saved entries, or
  defer recovery until later. Reopened sessions start fresh instances of their
  saved tools in their saved directories.
- Keep saved session entries in sync when creating, renaming, or deleting
  sessions, while the dashboard is open, and when reading the session list.
  Retain missing entries when the tmux server stops.
- Lock catalog updates across Hive processes, write snapshots atomically with
  private permissions, and skip writes when nothing changed.
- Preserve saved entries on invalid JSON, tmux errors, or failed restoration.
  Reject missing tools and directories, and leave existing live sessions alone.
- Treat an absent tmux server as an empty session list so recovery works after
  a reboot. Keep recovery prompts out of the compact session switcher.
- Preserve selection and open overlays during refreshes, reject stale refresh
  results, and keep delete confirmations bound to their original session.
- Fit recovery prompts to smaller terminals and show restoration errors in place.
- Add tests for catalog concurrency, corrupt data, lifecycle operations,
  recovery actions, viewport sizing, and real tmux restoration in quoted paths.

### Installation and documentation

- Install to `~/.local/bin` by default without sudo, and explain how to migrate
  older system installations to a user-writable directory for future updates.
- Fail with a clear error for unwritable destinations instead of invoking sudo.
  Point updater permission errors to the user-owned installation instructions.
- Test default installation with spaces in the home path, reject attempts to
  stage outside the test directory, and verify the installer never invokes sudo.
- Document recovery storage and limits, current installation commands, release
  validation, and rollback to 0.3.2. Recovery saves setup, not terminal history,
  running process state, arguments, or custom pane layouts.

## 0.3.2

[Changes since 0.3.1](https://github.com/sperka/hive/compare/v0.3.1...v0.3.2)

- Remove the keyboard legend's background fill while keeping its text bright.
- Show the active theme name dimmed in the dashboard's bottom-right corner.
- Keep the theme name and cycling order in sync with the saved theme on startup.
- Add theme regression tests and update the release and rollback instructions.

## 0.3.1

[Changes since 0.3.0](https://github.com/sperka/hive/compare/v0.3.0...v0.3.1)

- Find updates through GitHub's public release redirect instead of the
  anonymous REST API, avoiding its shared-IP rate limit.
- Validate that discovery resolves to a stable release tag in this repository,
  and pin archive and checksum downloads to that same tag.
- Publish a checksum-listed installer with each release so installation no
  longer depends on raw.githubusercontent.com or the GitHub API.
- Verify the published installer matches the tagged source before promoting the
  draft release. Add bootstrap instructions for older updaters affected by 403s.
- Keep the full keyboard legend directly below the dashboard, wrapping
  shortcuts on narrow terminals.
- Make the legend readable with bright labels, bold keys, and a filled
  background that follows the selected theme.
- Add redirect, installer, and dashboard regression coverage.

## 0.3.0

[Changes since 0.2.0](https://github.com/sperka/hive/compare/v0.2.0...v0.3.0)

- Redesign the dashboard around a 100-column frame with Sessions and Details
  side by side and CPU activity below.
- Restore the original 3-4-3 hexagons, with a bold HIVE/version label and
  right-aligned session totals and state icons.
- Replace simulated CPU history with real measurements, including average,
  peak, graph scale, sample count, and physical/logical core counts.
- Add color-coded per-core percentages in four or eight columns, numbered
  top to bottom, with shared alignment and subdued statistics.
- Identify startup and unavailable CPU samples without presenting them as zero.
- Add optional Kitty graphics for Ghostty/Kitty, including SSH and existing
  tmux passthrough support; text hexagons remain the default.
- Improve narrow-terminal layouts, long Unicode fields, overlay resizing,
  and selection after returning from an attached session.
- Share the TUI startup path between dashboard and switcher so graphics and
  terminal cleanup run consistently.
- Add CPU sampler, activity panel, header, graphics, and dashboard tests, plus
  design and implementation notes.
- Update installation and release validation, derive installer test versions
  from the changelog, and verify updates from 0.2.0.

## 0.2.0

[Changes since 0.1.2](https://github.com/sperka/hive/compare/v0.1.2...v0.2.0)

### Tools and sessions

- Configure ordered tool choices in `config.json` with defaults of `pi`,
  `claude`, `codex`, and `zsh`. Create a default config on first use and preserve
  defaults for older configs that have no tool list.
- Add `Custom...` input for executable names or paths in the new-session form.
  Accept the same custom executables through `hive create --tool`, validate
  executability, and retain the editor split layout for `nvim` and `vim`.
- Rename sessions with `r`, preserving running processes, tool, directory,
  creation time, and runtime status.
- Assign stable state IDs so renaming does not break running tools' state
  writes. Preserve legacy state keys, target exact tmux sessions, and roll back
  a rename if its metadata update fails.
- Resolve the current owning session when an editor exits so detaching still
  works after a rename. Clean up state after session deletion or failed creation.

### Installation and updates

- Add `hive update` and `hive update --check` using `sperka/hive` releases.
  Leave equal or newer installed versions in place and let development builds
  update to the latest stable release.
- Verify archive SHA-256 checksums, reject invalid archives, and atomically
  replace the executable, including installations launched through symlinks.
- Publish Linux and macOS binaries for amd64 and arm64 from this fork, with
  the release version embedded in each binary.
- Add installer checksum verification, custom destinations through
  `HIVE_INSTALL_DIR`, and pinned releases through `HIVE_VERSION`. At this
  version, the default destination is `/usr/local/bin` with sudo when needed.
- Stage installs in the destination directory and preserve an existing binary
  if downloads, verification, or extraction fail.
- Add release instructions and rollback guidance. Test and verify draft release
  assets before publishing them as the latest stable release.
- Pin Go, GoReleaser, and ShellCheck through mise; derive build versions from
  Git tags and retain the upstream Go module import path.
- Add config, form, rename, runtime, updater, installer, and isolated tmux tests.

### Upstream work included after 0.1.2

- Introduce a shared session lifecycle service for CLI and TUI operations,
  including common name, tool, and working-directory handling.
- Separate TUI screens and overlays into explicit application state and add
  lifecycle and state-transition tests.
- Add the `prefix+h` session switcher popup and `hive switch` command, with
  `sw` and `s` aliases, live filtering, and switching inside existing tmux clients.
- Install the popup binding on Hive-created sessions and document how to add
  it to existing sessions.

## 0.1.2

[Upstream changes since 0.1.1](https://github.com/sperka/hive/compare/v0.1.1...v0.1.2)

- Add editor workspaces for the `nvim` and `vim` tool choices, with Neovim on the
  left and a shell on the right. Configure pane borders and window backgrounds.
- Detach when the editor exits and reopen it on the next attachment while
  keeping the shell pane alive.
- Introduce the internal `run-session` runner and persisted metadata and runtime
  records, including start/end times and exit codes.
- Preserve the original tool, working directory, and creation time instead of
  replacing them with the current pane command.
- Distinguish active, running, ready, completed, and failed sessions, and
  update dashboard status labels and totals.
- Quote tool paths and commands for tmux, and add tests for editor layout,
  session commands, state storage, and runtime status.

## 0.1.1

[Upstream changes since 0.1.0](https://github.com/sperka/hive/compare/v0.1.0...v0.1.1)

- Hide the tmux status bar and enable mouse support in newly created sessions.
- Add the tag-triggered GitHub Actions release workflow.
- Add a Jekyll landing page, demo, Open Graph images, and a GitHub Pages workflow
  limited to website changes. Remove the initial custom-domain setting.
- Improve README formatting and ignore local development notes and build output.

## 0.1.0

[Original upstream release](https://github.com/sperka/hive/tree/v0.1.0)

- Introduce a centered TUI for managing tmux sessions, with a honeycomb logo,
  session list, detail panel, activity sparkline, and session totals.
- Create, attach to, and delete sessions; choose a tool, browse for a working
  directory, and set a session name in the new-session form.
- Add CLI commands `list`, `create`, `attach`, and `delete`, with aliases and
  JSON session output for scripts and AI agents.
- Search with `/`, confirm deletions, open help with `?`, and scroll the session
  list when it exceeds the visible area.
- Provide 12 themes with `t` to cycle and save the selected theme in
  `~/.config/hive/config.json`.
- Generate session names using `crypto/rand` for cross-platform support.
- Add version and help flags, Makefile build/version injection, GoReleaser
  configuration, and an installation script.
- Include debug logging, the README and demo, MIT license, changelog, and
  initial design notes.
