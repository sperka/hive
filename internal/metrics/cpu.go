package metrics

import (
	"fmt"
	"math"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
)

type CPU struct {
	Usage         float64
	PerCore       []float64
	PhysicalCores int
	LogicalCPUs   int
	Valid         bool
	Unavailable   bool
}

// Sampler compares consecutive counters; its first sample is deliberately unmeasured.
// A dashboard owns one sampler and runs its commands sequentially.
type Sampler struct {
	read     func() ([]cpu.TimesStat, error)
	previous []cpu.TimesStat
	physical int
	logical  int
}

func NewSampler() *Sampler {
	physical, _ := cpu.Counts(false)
	logical, _ := cpu.Counts(true)
	if logical <= 0 {
		logical = runtime.NumCPU()
	}
	return &Sampler{
		read:     func() ([]cpu.TimesStat, error) { return cpu.Times(true) },
		physical: physical, logical: logical,
	}
}

func (s *Sampler) Sample() (CPU, error) {
	result := CPU{PhysicalCores: s.physical, LogicalCPUs: s.logical}
	current, err := s.read()
	if err != nil || len(current) == 0 {
		s.previous = nil
		result.Unavailable = true
		if err == nil {
			err = fmt.Errorf("CPU counters unavailable")
		}
		return result, err
	}
	result.LogicalCPUs = len(current)
	previous := make(map[string]cpu.TimesStat, len(s.previous))
	for _, item := range s.previous {
		previous[item.CPU] = item
	}
	s.previous = current
	result.PerCore = make([]float64, len(current))
	var total, busy float64
	for i, now := range current {
		result.PerCore[i] = math.NaN()
		before, ok := previous[now.CPU]
		if !ok {
			continue
		}
		elapsed := counterTotal(now) - counterTotal(before)
		idle := (now.Idle + now.Iowait) - (before.Idle + before.Iowait)
		work := elapsed - idle
		if elapsed <= 0 || idle < 0 || work < 0 || math.IsNaN(elapsed) || math.IsInf(elapsed, 0) {
			continue
		}
		result.PerCore[i] = math.Max(0, math.Min(100, work/elapsed*100))
		total += elapsed
		busy += work
	}
	if total > 0 {
		result.Usage = math.Max(0, math.Min(100, busy/total*100))
		result.Valid = true
	}
	return result, nil
}

func counterTotal(t cpu.TimesStat) float64 {
	// Guest and GuestNice are already included in User/Nice on Linux.
	return t.User + t.Nice + t.System + t.Idle + t.Iowait + t.Irq + t.Softirq + t.Steal
}
