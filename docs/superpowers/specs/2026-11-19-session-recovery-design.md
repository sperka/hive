# Session recovery

Persist the sessions Hive observes so restarting the computer does not lose the
names, tools, and project directories needed to reopen them.

## Approach

Maintain a readable JSON array at `~/.config/hive/sessions.json`. Each record has
`name`, `tool`, `path`, and `created_at`. Reuse Hive's existing creation behavior
when reopening. A tmux resurrection plugin would add an external dependency;
storing process state and arbitrary window layouts exceeds this simple catalog.
The user confirmed that name, tool, and directory recovery are enough.

## Behavior

- Reconcile live tmux sessions with saved records on startup and dashboard refresh.
- Add new live sessions and update their saved setup. Preserve missing records.
- Show each missing session with Open, Delete saved entry, and Later actions.
  Escape postpones all remaining choices until next startup.
- Open recreates the session in its saved directory using its saved executable.
  Failure leaves the record available and displays the error.
- Delete saved entry removes only that record, never an active tmux session.
- Explicit Hive create, rename, and delete operations update the catalog.
- An empty or absent tmux server is a normal state. Other tmux errors and invalid
  JSON must be reported and must not replace the last saved setup.
- Write atomically and serialize read/modify/write operations across Hive
  processes. Do not expire records. Skip unchanged writes.
- Keep existing CLI JSON output compatible. The compact switcher can update the
  catalog without interrupting switching with recovery prompts.

## Limits and constraints

Use existing Go, Bubble Tea, and styling dependencies. Restore fresh processes;
terminal history, running program state, command arguments, and custom tmux
layouts are outside this change. Live sessions not created by Hive use tmux's
reported tool and session directory. Reconcile while the dashboard is open and
after returning from an attached session; this is not an always-on daemon.

Existing unrelated working-tree edits must be preserved.
