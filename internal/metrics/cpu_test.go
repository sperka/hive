package metrics

import (
	"errors"
	"math"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
)

func TestSamplerUsesCounterDeltas(t *testing.T) {
	snapshots := [][]cpu.TimesStat{
		{{CPU: "cpu0", User: 10, Idle: 90}, {CPU: "cpu1", User: 20, Idle: 80}},
		{{CPU: "cpu0", User: 20, Idle: 100}, {CPU: "cpu1", User: 20, Idle: 100}},
	}
	index := 0
	s := &Sampler{physical: 1, logical: 2, read: func() ([]cpu.TimesStat, error) {
		value := snapshots[index]
		index++
		return value, nil
	}}
	first, err := s.Sample()
	if err != nil || first.Valid || first.PhysicalCores != 1 || first.LogicalCPUs != 2 {
		t.Fatalf("first sample must show counts but no invented usage: %+v, %v", first, err)
	}
	next, err := s.Sample()
	if err != nil || !next.Valid || next.Usage != 25 || next.PerCore[0] != 50 || next.PerCore[1] != 0 {
		t.Fatalf("expected 25%% aggregate, 50%%/0%% per CPU: %+v, %v", next, err)
	}
}

func TestSamplerGuestTimeIsNotCountedTwice(t *testing.T) {
	// Linux includes guest time in user; iowait is idle.
	s := &Sampler{
		previous: []cpu.TimesStat{{CPU: "cpu0"}},
		read: func() ([]cpu.TimesStat, error) {
			return []cpu.TimesStat{{CPU: "cpu0", User: 30, Guest: 20, Idle: 60, Iowait: 10}}, nil
		},
	}
	got, err := s.Sample()
	if err != nil || got.Usage != 30 {
		t.Fatalf("expected 30%% without double-counting guest time: %+v, %v", got, err)
	}
}

func TestSamplerCounterResetAndReadFailureAreNotIdle(t *testing.T) {
	s := &Sampler{
		previous: []cpu.TimesStat{{CPU: "cpu0", User: 100, Idle: 100}},
		read: func() ([]cpu.TimesStat, error) {
			return []cpu.TimesStat{{CPU: "cpu0", User: 1, Idle: 1}}, nil
		},
	}
	got, err := s.Sample()
	if err != nil || got.Valid || !math.IsNaN(got.PerCore[0]) {
		t.Fatalf("reset counters must be unavailable: %+v, %v", got, err)
	}
	s.read = func() ([]cpu.TimesStat, error) { return nil, errors.New("read failed") }
	got, err = s.Sample()
	if err == nil || got.Valid || !got.Unavailable {
		t.Fatalf("failed read must be unavailable: %+v, %v", got, err)
	}
}
