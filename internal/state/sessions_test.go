package state

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func savedPath(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return filepath.Join(home, ".config", "hive", "sessions.json")
}

func TestSavedSessionsRoundTripAndConcurrentUpdates(t *testing.T) {
	path := savedPath(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := UpdateSavedSessions(func(saved []Metadata) ([]Metadata, error) {
				return append(saved, Metadata{Name: fmt.Sprintf("session-%02d", i), Tool: "sh", Path: "/tmp"}), nil
			})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	saved, err := ReadSavedSessions()
	if err != nil || len(saved) != 20 {
		t.Fatalf("saved = %#v, err = %v", saved, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("saved file should be private: %v, %v", info, err)
	}
	for i, record := range saved {
		if record.Name != fmt.Sprintf("session-%02d", i) || record.Path != "/tmp" || record.Tool != "sh" {
			t.Fatalf("record %d = %#v", i, record)
		}
	}
	if err := UpdateSavedSessions(func(saved []Metadata) ([]Metadata, error) { return saved, nil }); err != nil {
		t.Fatal(err)
	}
	unchanged, _ := os.Stat(path)
	if !os.SameFile(info, unchanged) {
		t.Fatal("unchanged catalog was unnecessarily replaced")
	}
}

func TestSavedSessionsRejectCorruptInputWithoutChangingFile(t *testing.T) {
	for _, data := range []string{`{`, `null`, `{}`, `[{"name":"same"},{"name":"same"}]`, `[{}]`} {
		t.Run(data, func(t *testing.T) {
			path := savedPath(t)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			called := false
			err := UpdateSavedSessions(func(saved []Metadata) ([]Metadata, error) {
				called = true
				return nil, nil
			})
			if err == nil || called {
				t.Fatalf("invalid catalog was accepted: called=%v, err=%v", called, err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != data {
				t.Fatal("invalid catalog was overwritten")
			}
		})
	}
}

func TestSavedSessionsFailedUpdatePreservesLastSnapshot(t *testing.T) {
	savedPath(t)
	if err := UpdateSavedSessions(func([]Metadata) ([]Metadata, error) {
		return []Metadata{{Name: "work", Tool: "sh", Path: "/project"}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	err := UpdateSavedSessions(func(saved []Metadata) ([]Metadata, error) {
		saved[0].Name = "changed"
		return saved, fmt.Errorf("tmux failed")
	})
	if err == nil {
		t.Fatal("expected update failure")
	}
	got, err := ReadSavedSessions()
	if err != nil || len(got) != 1 || got[0].Name != "work" {
		t.Fatalf("last snapshot lost: %#v, %v", got, err)
	}
}
