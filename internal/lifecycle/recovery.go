package lifecycle

import (
	"fmt"
	"slices"
	"time"

	"github.com/mdipanjan/hive/internal/session"
	"github.com/mdipanjan/hive/internal/state"
	"github.com/mdipanjan/hive/internal/tmux"
)

// Reconcile records live setup without forgetting sessions lost on reboot.
func (s Service) Reconcile() (live []session.Session, missing []state.Metadata, err error) {
	err = state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		var listErr error
		live, listErr = tmux.List()
		if listErr != nil {
			return nil, listErr
		}
		present := make(map[string]bool, len(live))
		for _, sess := range live {
			present[sess.Name] = true
			saved = upsertSaved(saved, metadata(sess))
		}
		for _, record := range saved {
			if !present[record.Name] {
				missing = append(missing, record)
			}
		}
		return saved, nil
	})
	return
}

// Restore reopens a saved session, or leaves an already-live namesake alone.
func (s Service) Restore(name string) error {
	return state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		live, err := tmux.List()
		if err != nil {
			return nil, err
		}
		for _, sess := range live {
			if sess.Name == name {
				return upsertSaved(saved, metadata(sess)), nil
			}
		}
		for _, record := range saved {
			if record.Name != name {
				continue
			}
			// Incomplete records must not silently fall back to the current
			// directory or a different tool.
			if record.Tool == "" || record.Path == "" {
				return nil, fmt.Errorf("saved session %q has no tool or directory", name)
			}
			req, err := s.prepareCreate(CreateRequest{Name: name, Tool: record.Tool, Path: record.Path})
			if err != nil {
				return nil, err
			}
			if err := tmux.Create(req.Name, req.Tool, req.Path); err != nil {
				return nil, err
			}
			record.CreatedAt = time.Now().UTC()
			return upsertSaved(saved, record), nil
		}
		return nil, fmt.Errorf("saved session %q no longer exists", name)
	})
}

// Forget only removes a saved entry. It never kills a live tmux session.
func (s Service) Forget(name string) error {
	return state.UpdateSavedSessions(func(saved []state.Metadata) ([]state.Metadata, error) {
		return removeSaved(saved, name), nil
	})
}

func metadata(sess session.Session) state.Metadata {
	return state.Metadata{Name: sess.Name, Tool: sess.Tool, Path: sess.Path, CreatedAt: sess.CreatedAt}
}

func upsertSaved(saved []state.Metadata, record state.Metadata) []state.Metadata {
	for i := range saved {
		if saved[i].Name == record.Name {
			saved[i] = record
			return saved
		}
	}
	return append(saved, record)
}

func removeSaved(saved []state.Metadata, name string) []state.Metadata {
	return slices.DeleteFunc(saved, func(record state.Metadata) bool { return record.Name == name })
}
