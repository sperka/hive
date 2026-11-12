# Changelog

All notable changes to this project will be documented in this file.

## 0.3.1

- Find updates through GitHub's public release redirect instead of the
  anonymous REST API, avoiding its shared-IP rate limit.
- Publish a checksum-listed installer with each release so installation no
  longer depends on raw.githubusercontent.com or the GitHub API.
- Keep the full keyboard legend directly below the dashboard, wrapping
  shortcuts on narrow terminals.
- Make the legend readable with bright labels, bold keys, and a filled
  background that follows the selected theme.

## 0.3.0

- Redesigned the dashboard around a 100-column frame with Sessions and Details
  side by side and CPU activity below.
- Restored the original 3-4-3 hexagons, with a bold HIVE/version label and
  right-aligned session totals and state icons.
- Replaced simulated CPU history with real measurements, including average,
  peak, graph scale, sample count, and physical/logical core counts.
- Added color-coded per-core percentages in four or eight columns, numbered
  top to bottom, with shared alignment and subdued statistics.
- Added optional Kitty graphics for Ghostty/Kitty, including SSH and existing
  tmux passthrough support; text hexagons remain the default.
- Improved narrow-terminal layouts, long Unicode fields, overlay resizing,
  and selection after returning from an attached session.
- Updated installation and release validation for v0.3.0. Existing v0.2.0
  installations can upgrade with `hive update`.

## 0.2.0

- Added `hive update` and `hive update --check` using `sperka/hive` releases.
- Verify release checksums and atomically replace the executable during updates.
- Publish Linux and macOS binaries for amd64 and arm64 from this fork.
- Updated the installer with checksum verification, custom install directories,
  and support for installing a specific release with `HIVE_VERSION`.
- Added configurable tool choices and custom executable input.
- Added session renaming while preserving running processes and session state.
- Added the `prefix+h` session switcher popup and `hive switch` command.

## 0.1.2

- Added runtime state tracking for Hive sessions.
- Added internal session runner.
- Improved nvim session resume behavior.
- Preserved original tool/path metadata for sessions.
- Improved status labels: active, running, ready, completed, failed.

## 0.1.1

- Fix directory browser when creating new sessions (press `b` on path field)
- Hide tmux status bar in created sessions for cleaner UI
- Enable mouse support in created sessions
- Add OG image for social sharing

## 0.1.0

- TUI for managing tmux sessions
- CLI mode with JSON output for AI agents
- Commands: list, create, attach, delete
- Search sessions with `/`
- Delete confirmation dialog
- Help popup with `?`
- 12 color themes with `t` to cycle
- CPU activity sparkline
- Session details panel
- Config persistence at `~/.config/hive/config.json`
