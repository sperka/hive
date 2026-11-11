package lifecycle

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mdipanjan/hive/internal/session"
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
	return tmux.List()
}

func (s Service) Create(req CreateRequest) (string, error) {
	tool := req.Tool
	if tool == "" {
		tool = "bash"
	}
	if _, err := exec.LookPath(tool); err != nil {
		return "", fmt.Errorf("tool %q is not an executable on PATH or an executable path: %w", tool, err)
	}

	name := req.Name
	if name == "" {
		name = "hive-" + s.generateID(6)
	}

	path, err := s.resolvePath(req.Path)
	if err != nil {
		return "", err
	}

	return name, tmux.Create(name, tool, path)
}

func (s Service) AttachCmd(name string) *exec.Cmd {
	return tmux.AttachCmd(name)
}

func (s Service) Delete(name string) error {
	return tmux.Kill(name)
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
