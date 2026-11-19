# Session Recovery Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans to implement task by task.

**Goal:** Save observed sessions and let users recover or discard missing sessions.

**Architecture:** Add a locked, atomic catalog to state, reconcile it in lifecycle,
and expose recovery as a dashboard overlay. Keep tmux as the live-session source.

**Tech Stack:** Go 1.25.6, Bubble Tea, Lip Gloss, tmux.

**Spec:** `docs/superpowers/specs/2026-11-19-session-recovery-design.md`

## Global Constraints

- File: `~/.config/hive/sessions.json`; JSON array of name/tool/path/created_at.
- No expiry, extra dependencies, or arbitrary process-state restoration.
- Preserve missing records and unrelated checkout changes.
- Perform implementation and review locally; do not publish.

## Review Focus

- Empty tmux server after reboot retains all saved sessions.
- Invalid JSON or failed tmux reads never overwrite saved sessions.
- Concurrent Hive operations do not lose catalog entries.
- Rename/delete do not leave stale recovery entries.
- Failed restoration and a name becoming live while prompted retain safe choices.

## Task 1: Durable catalog and lifecycle

**Files:** `internal/state/sessions.go`, `internal/state/sessions_test.go`,
`internal/lifecycle/recovery.go`, `internal/lifecycle/recovery_test.go`,
`internal/lifecycle/lifecycle.go`, `internal/tmux/tmux.go`,
`internal/tmux/tmux_test.go`, `internal/testutil/tmux.go`.

**Interfaces:** `state.ReadSavedSessions() ([]Metadata, error)`,
`state.UpdateSavedSessions(func([]Metadata) ([]Metadata, error)) error`,
`Service.Reconcile() ([]session.Session, []state.Metadata, error)`,
`Service.Restore(name string) error`, `Service.Forget(name string) error`.

- [x] Write tests for reboot retention, startup adoption, concurrent updates,
  corrupt files, failed reads, restore, rename, and explicit deletion.
- [x] Run `mise exec -- go test ./internal/state ./internal/lifecycle ./internal/tmux`.
  Confirm missing behavior fails.
- [x] Implement atomic replacement with file locking and lifecycle reconciliation.
- [x] Run the same tests and confirm success.

## Task 2: Recovery prompt and maintained snapshots

**Files:** `internal/tui/recovery.go`, `internal/tui/recovery_test.go`,
`internal/tui/model.go`, `internal/tui/appstate.go`, `internal/tui/view.go`,
`internal/tui/delete.go`, `internal/tui/form.go`, `README.md`.

**Interfaces:** Recovery overlay uses `Service.Restore` and `Service.Forget`.
Dashboard refresh messages reconcile snapshots without swallowing overlay input.

- [x] Write tests for startup prompts, skip retention, delete-only-saved behavior,
  successful/failed opening, and refresh during overlays.
- [x] Run `mise exec -- go test ./internal/tui`; confirm missing behavior fails.
- [x] Implement prompt and periodic reconciliation, preserving selection and errors.
- [x] Document storage and recovery limits.
- [x] Run `mise exec -- go test -race ./...`, `mise exec -- go vet ./...`, and build.
- [x] Review full diff and verify recovery against an isolated real tmux server.

## Verification notes

- All package tests pass, including the race detector; `go vet ./...` is clean.
- Built Hive created a real session, retained it after `kill-server`, and reopened
  it through the recovery prompt with the exact saved tool and quoted directory.
- Review found that refresh could move an open delete confirmation to a different
  session. A failing regression test verified the issue; confirmation now keeps
  the original name. Small-viewport recovery overflow was also reproduced and
  fixed with a regression test.
- Scope confirmed by the user: name, tool, and directory are sufficient.
