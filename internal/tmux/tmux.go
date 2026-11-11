package tmux

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mdipanjan/hive/internal/logger"
	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
)

func List() ([]session.Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}|#{session_created}|#{session_activity}|#{session_attached}|#{session_path}|#{pane_current_command}|#{@hive_state_id}")
	output, err := cmd.Output()

	logger.Log.Debug("tmux list-sessions", "output", string(output), "err", err)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	sessions := make([]session.Session, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		created, _ := strconv.ParseInt(parts[1], 10, 64)
		activity, _ := strconv.ParseInt(parts[2], 10, 64)
		attached, _ := strconv.ParseInt(parts[3], 10, 64)

		path := ""
		tool := ""
		if len(parts) >= 5 {
			path = parts[4]
		}
		if len(parts) >= 6 {
			tool = parts[5]
		}

		createdAt := time.Unix(created, 0)
		lastActivity := time.Unix(activity, 0)

		stateID := name
		if len(parts) >= 7 && parts[6] != "" {
			stateID = parts[6]
		}
		meta, metaErr := state.ReadMetadata(stateID)
		if metaErr == nil {
			tool = meta.Tool
			path = meta.Path
			createdAt = meta.CreatedAt
		}

		runtime, runtimeErr := state.ReadRuntime(stateID)
		status := statusFromRuntime(runtime, runtimeErr, attached, lastActivity)

		sessions = append(sessions, session.Session{
			Name:         name,
			Tool:         tool,
			Path:         path,
			Status:       status,
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
		})
	}

	logger.Log.Debug("parsed sessions", "count", len(sessions))
	return sessions, nil
}

func statusFromRuntime(runtime state.Runtime, err error, attached int64, lastActivity time.Time) session.Status {
	if attached > 0 {
		return session.StatusActive
	}

	if err != nil {
		if time.Since(lastActivity) < 5*time.Minute {
			return session.StatusRunning
		}
		return session.StatusIdle
	}

	switch runtime.State {
	case state.StateStarting, state.StateRunning:
		return session.StatusRunning
	case state.StateReady:
		return session.StatusReady
	case state.StateCompleted:
		return session.StatusCompleted
	case state.StateFailed:
		return session.StatusFailed
	case state.StateUnknown:
		return session.StatusUnknown
	default:
		return session.StatusIdle
	}
}

func Create(name, tool, path string) error {
	logger.Log.Debug("creating tmux session", "name", name, "tool", tool, "path", path)

	stateID := "session-" + rand.Text()
	if err := state.WriteMetadataForID(stateID, state.Metadata{Name: name, Tool: tool, Path: path, CreatedAt: time.Now().UTC()}); err != nil {
		return err
	}

	if err := state.WriteRuntime(stateID, state.Runtime{State: state.StateStarting, StartedAt: time.Now().UTC()}); err != nil {
		_ = state.DeleteSessionState(stateID)
		return err
	}

	err := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", path,
		runnerCommand(name, tool, path, stateID),
		";", "set-option", "-t", "="+name+":", "@hive_state_id", stateID).Run()
	if err != nil {
		logger.Log.Error("tmux create failed", "err", err)
		_ = state.DeleteSessionState(stateID)
		return err
	}

	if tool == "nvim" || tool == "vim" {
		for _, args := range editorLayoutCommands(name, path) {
			if err := exec.Command("tmux", args...).Run(); err != nil {
				logger.Log.Error("tmux editor layout failed", "args", args, "err", err)
				return err
			}
		}
	}

	configureSession(name)
	return nil
}

func runnerCommand(name, tool, path, stateID string) string {
	executable, err := os.Executable()
	if err != nil {
		executable = "hive"
	}

	command := shellQuote(executable) + " run-session --name " + shellQuote(name) + " --tool " + shellQuote(tool) + " --path " + shellQuote(path)
	command += " --state-id " + shellQuote(stateID)
	if stateDir := os.Getenv("HIVE_STATE_DIR"); stateDir != "" {
		command = "HIVE_STATE_DIR=" + shellQuote(stateDir) + " " + command
	}
	return command
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func editorLayoutCommands(name, path string) [][]string {
	return [][]string{
		{"split-window", "-h", "-p", "35", "-t", name + ":0", "-c", path},
		{"select-pane", "-t", name + ":0.0"},
	}
}

func configureSession(name string) {
	for _, args := range sessionOptions(name) {
		exec.Command("tmux", args...).Run()
	}
	for _, args := range sessionKeyBindings() {
		exec.Command("tmux", args...).Run()
	}
}

func sessionKeyBindings() [][]string {
	return [][]string{
		{"bind-key", "h", "display-popup", "-B", "-w", "90%", "-h", "80%", "-E", hiveSwitchCommand()},
	}
}

func hiveSwitchCommand() string {
	executable, err := os.Executable()
	if err != nil {
		executable = "hive"
	}
	return shellQuote(executable) + " switch"
}

func sessionOptions(name string) [][]string {
	return [][]string{
		{"set-option", "-t", name, "status", "off"},
		{"set-option", "-t", name, "mouse", "on"},
		{"set-option", "-t", name, "pane-border-style", "fg=#45475a"},
		{"set-option", "-t", name, "pane-active-border-style", "fg=#a6e3a1"},
		{"set-option", "-t", name, "window-style", "bg=#1e1e2e"},
		{"set-option", "-t", name, "window-active-style", "bg=#1e1e2e"},
	}
}

func Attach(name string) error {
	cmd := exec.Command("tmux", "attach", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func AttachCmd(name string) *exec.Cmd {
	if IsInsideTmux() {
		logger.Log.Debug("using switch-client (inside tmux)", "target", name)
		return exec.Command("tmux", "switch-client", "-t", name)
	}
	logger.Log.Debug("using attach (outside tmux)", "target", name)
	return exec.Command("tmux", "attach", "-t", name)
}

func Kill(name string) error {
	logger.Log.Debug("killing tmux session", "name", name)
	target, stateID, err := sessionIdentity(name)
	if err != nil {
		return err
	}
	err = exec.Command("tmux", "kill-session", "-t", target).Run()
	if err != nil {
		logger.Log.Error("tmux kill failed", "err", err)
		return err
	}
	if stateErr := state.DeleteSessionState(stateID); stateErr != nil {
		logger.Log.Error("session state delete failed", "err", stateErr)
		return stateErr
	}
	return nil
}

// sessionIdentity also supports sessions created before stable state IDs existed.
func sessionIdentity(name string) (target, stateID string, err error) {
	output, err := exec.Command("tmux", "display-message", "-p", "-t", "="+name+":",
		"#{session_id}|#{@hive_state_id}").CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("find session %q: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	parts := strings.SplitN(strings.TrimSpace(string(output)), "|", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "$") {
		return "", "", fmt.Errorf("invalid tmux session identity %q", output)
	}
	stateID = parts[1]
	if stateID == "" {
		stateID = name
	}
	return parts[0], stateID, nil
}

func Rename(name, newName string) error {
	if name == newName {
		return nil
	}
	target, stateID, err := sessionIdentity(name)
	if err != nil {
		return err
	}
	meta, metaErr := state.ReadMetadata(stateID)
	if metaErr != nil && !errors.Is(metaErr, os.ErrNotExist) {
		return fmt.Errorf("read session metadata: %w", metaErr)
	}

	// Legacy runners continue writing under their original name. Keep that key
	// on the tmux session instead of moving files out from under the process.
	if err := exec.Command("tmux", "set-option", "-t", target, "@hive_state_id", stateID).Run(); err != nil {
		return fmt.Errorf("preserve session state: %w", err)
	}
	if err := renameTarget(target, newName); err != nil {
		return err
	}
	if metaErr == nil {
		meta.Name = newName
		if err := state.WriteMetadataForID(stateID, meta); err != nil {
			rollbackErr := renameTarget(target, name)
			return errors.Join(fmt.Errorf("update session metadata: %w", err), rollbackErr)
		}
	}
	return nil
}

func renameTarget(target, name string) error {
	output, err := exec.Command("tmux", "rename-session", "-t", target, "--", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("rename session: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
