package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type RuntimeState string

const (
	StateStarting  RuntimeState = "starting"
	StateRunning   RuntimeState = "running"
	StateCompleted RuntimeState = "completed"
	StateFailed    RuntimeState = "failed"
	StateReady     RuntimeState = "ready"
	StateUnknown   RuntimeState = "unknown"
)

type Metadata struct {
	Name      string    `json:"name"`
	Tool      string    `json:"tool"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

type Runtime struct {
	State     RuntimeState `json:"state"`
	StartedAt time.Time    `json:"started_at"`
	EndedAt   *time.Time   `json:"ended_at,omitempty"`
	ExitCode  *int         `json:"exit_code,omitempty"`
}

func WriteMetadata(meta Metadata) error {
	return WriteMetadataForID(meta.Name, meta)
}

// WriteMetadataForID keeps storage independent of a session's editable name.
func WriteMetadataForID(id string, meta Metadata) error {
	return writeJSON(metadataPath(id), meta)
}

func ReadMetadata(name string) (Metadata, error) {
	var meta Metadata
	err := readJSON(metadataPath(name), &meta)
	return meta, err
}

func WriteRuntime(name string, runtime Runtime) error {
	return writeJSON(runtimePath(name), runtime)
}

func ReadRuntime(name string) (Runtime, error) {
	var runtime Runtime
	err := readJSON(runtimePath(name), &runtime)
	if errors.Is(err, os.ErrNotExist) {
		return Runtime{State: StateUnknown}, nil
	}
	return runtime, err
}

func DeleteSessionState(name string) error {
	return os.RemoveAll(sessionDir(name))
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

func metadataPath(name string) string {
	return filepath.Join(sessionDir(name), "metadata.json")
}

func runtimePath(name string) string {
	return filepath.Join(sessionDir(name), "state.json")
}

func sessionDir(name string) string {
	return filepath.Join(baseDir(), name)
}

func baseDir() string {
	if dir := os.Getenv("HIVE_STATE_DIR"); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".hive", "sessions")
	}

	return filepath.Join(home, ".hive", "sessions")
}
