package lifecycle

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/tmux"
)

type CreateRequest struct {
	Name string
	Tool string
	Path string
}

type Service struct {
	generateID func(int) string
	workingDir func() (string, error)
}

func New() Service {
	return Service{generateID: randomID, workingDir: os.Getwd}
}

func (s Service) List() ([]session.Session, error) {
	live, _, err := s.Reconcile()
	return live, err
}

func (s Service) Create(req CreateRequest) (string, error) {
	req, err := s.prepareCreate(req)
	if err != nil {
		return "", err
	}
	err = state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		if err := tmux.Create(req.Name, req.Tool, req.Path); err != nil {
			return nil, err
		}
		return upsertSaved(saved, state.Metadata{
			Name: req.Name, Tool: req.Tool, Path: req.Path, CreatedAt: time.Now().UTC(),
		}), nil
	})
	return req.Name, err
}

func (s Service) prepareCreate(req CreateRequest) (CreateRequest, error) {
	tool := req.Tool
	if tool == "" {
		tool = "bash"
	}
	if _, err := exec.LookPath(tool); err != nil {
		return req, fmt.Errorf("tool %q is not an executable on PATH or an executable path: %w", tool, err)
	}

	name := req.Name
	if name == "" {
		name = "hive-" + s.generateID(6)
	}
	if err := validateName(name); err != nil {
		return req, err
	}

	path, err := s.resolvePath(req.Path)
	if err != nil {
		return req, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return req, fmt.Errorf("session directory: %w", err)
	}
	if !info.IsDir() {
		return req, fmt.Errorf("session path %q is not a directory", path)
	}
	return CreateRequest{Name: name, Tool: tool, Path: path}, nil
}

func (s Service) AttachCmd(name string) *exec.Cmd {
	return tmux.AttachCmd(name)
}

func (s Service) Delete(name string) error {
	return state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		if err := tmux.Kill(name); err != nil {
			return nil, err
		}
		return removeSaved(saved, name), nil
	})
}

func (s Service) Rename(name, newName string) error {
	if err := validateName(newName); err != nil {
		return err
	}
	return state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		live, err := tmux.List()
		if err != nil {
			return nil, err
		}
		if err := tmux.Rename(name, newName); err != nil {
			return nil, err
		}
		saved = removeSaved(saved, name)
		for _, sess := range live {
			if sess.Name == name {
				sess.Name = newName
				saved = upsertSaved(saved, metadata(sess))
				break
			}
		}
		return saved, nil
	})
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("session name cannot be empty")
	}
	if strings.ContainsAny(name, ".:/\\|") || strings.ContainsFunc(name, unicode.IsControl) {
		return fmt.Errorf("session name cannot contain dots, colons, slashes, pipes, or control characters")
	}
	return nil
}

func (s Service) resolvePath(value string) (string, error) {
	if value == "" || value == "." {
		return s.workingDir()
	}
	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		value = filepath.Join(home, value[2:])
	}
	return filepath.Abs(value)
}

func randomID(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}
