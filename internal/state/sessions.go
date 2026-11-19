package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/sys/unix"
)

func savedSessionsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "hive", "sessions.json"), nil
}

// ReadSavedSessions reads an atomically published snapshot.
func ReadSavedSessions() ([]Metadata, error) {
	path, err := savedSessionsPath()
	if err != nil {
		return nil, err
	}
	saved, _, err := readSavedSessions(path)
	return saved, err
}

func readSavedSessions(path string) ([]Metadata, []byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Metadata{}, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	var saved []Metadata
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	if saved == nil {
		return nil, nil, fmt.Errorf("read %s: expected a JSON array of sessions", path)
	}
	if err := validateSavedSessions(saved); err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	return saved, data, nil
}

func validateSavedSessions(saved []Metadata) error {
	names := make(map[string]bool, len(saved))
	for _, record := range saved {
		if record.Name == "" || names[record.Name] {
			return fmt.Errorf("saved sessions must have nonempty, unique names: %q", record.Name)
		}
		names[record.Name] = true
	}
	return nil
}

// UpdateSavedSessions serializes read/modify/write operations across Hive
// processes. The callback also runs under the lock, so a stale live-session
// observation cannot resurrect an entry another Hive process just deleted.
func UpdateSavedSessions(update func([]Metadata) ([]Metadata, error)) error {
	path, err := savedSessionsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	for {
		err = unix.Flock(int(lock.Fd()), unix.LOCK_EX)
		if !errors.Is(err, unix.EINTR) {
			break
		}
	}
	if err != nil {
		return err
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)

	saved, previous, err := readSavedSessions(path)
	if err != nil {
		return err
	}
	saved, err = update(saved)
	if err != nil {
		return err
	}
	if saved == nil {
		saved = []Metadata{}
	}
	if err := validateSavedSessions(saved); err != nil {
		return err
	}
	sort.Slice(saved, func(i, j int) bool { return saved[i].Name < saved[j].Name })
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if bytes.Equal(previous, data) {
		return nil
	}
	if err := replaceSavedSessions(path, data); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}

func replaceSavedSessions(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".sessions-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
