package main

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

const mib = 1024 * 1024

// These are cold-start process estimates, not a table of runner sizes. Never
// lower them after observing a small package: a later package can be much larger.
var initialEstimates = map[string]int64{
	"compile": 4096, "link": 8192, "test": 4096, "other": 4096,
}

type config struct {
	CPUQuota       string `json:"cpu_quota"`
	MemoryQuotaGiB string `json:"memory_quota_gib"`
	CPU            int    `json:"cpu_parallelism"`
	BudgetMiB      int64  `json:"memory_budget_mib"`
	HeadroomMiB    int64  `json:"memory_headroom_mib"`
}

func readConfig(getenv func(string) string) config {
	c := config{CPUQuota: getenv("cpu_limit"), MemoryQuotaGiB: getenv("memory_Limit"), CPU: 16}
	if n, err := strconv.Atoi(getenv("GOMAXPROCS")); err == nil && n > 0 {
		c.CPU = n
	}
	if n, err := strconv.ParseFloat(c.CPUQuota, 64); err == nil && n > 0 && n < float64(math.MaxInt32) {
		c.CPU = max(1, int(n))
	}
	if n, err := strconv.ParseFloat(c.MemoryQuotaGiB, 64); err == nil && n >= 1.0/1024 && n < float64(math.MaxInt64/mib/1024) {
		quota := int64(n * 1024)
		c.BudgetMiB = quota * 4 / 5
		c.HeadroomMiB = quota - c.BudgetMiB
	}
	return c
}

type usage struct {
	CurrentMiB int64 `json:"current_mib"`
	PeakMiB    int64 `json:"peak_mib"`
	Finished   bool  `json:"finished"`
	ExitCode   int   `json:"exit_code"`
}

type reservation struct {
	Phase      string
	MiB        int64
	CurrentMiB int64
}

type phaseStats struct {
	Processes       int   `json:"processes"`
	MaxActive       int   `json:"max_active"`
	PeakMiB         int64 `json:"peak_commit_mib"`
	EstimateMiB     int64 `json:"next_reservation_mib"`
	WaitMillis      int64 `json:"wait_ms"`
	ExecutionMillis int64 `json:"execution_ms"`
}

type budget struct {
	mu             sync.Mutex
	config         config
	active         map[*reservation]bool
	stats          map[string]*phaseStats
	maxActive      int
	maxReservedMiB int64
	// Available system commit, rather than free physical RAM. Tests inject it.
	commitAvailable func() (int64, error)
}

func newBudget(c config, available func() (int64, error)) *budget {
	b := &budget{config: c, active: make(map[*reservation]bool), stats: make(map[string]*phaseStats), commitAvailable: available}
	for phase, estimate := range initialEstimates {
		b.stats[phase] = &phaseStats{EstimateMiB: estimate}
	}
	return b
}

// Check and reserve under one lock: simultaneous tools must not all see the
// same free memory. Account for growth already promised to running processes.
func (b *budget) tryAcquire(phase string) (*reservation, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.active) >= b.config.CPU {
		return nil, nil
	}
	s, ok := b.stats[phase]
	if !ok {
		return nil, fmt.Errorf("unknown phase %q", phase)
	}
	reserved, growth, samePhase := int64(0), int64(0), 0
	for r := range b.active {
		reserved += r.MiB
		growth += max(0, r.MiB-r.CurrentMiB)
		if r.Phase == phase {
			samePhase++
		}
	}
	amount := int64(0)
	if b.config.BudgetMiB > 0 {
		// A runner smaller than an initial estimate can still run one process,
		// with a correspondingly smaller Go soft limit. This is not a hard cap.
		amount = min(s.EstimateMiB, b.config.BudgetMiB)
		if reserved+amount > b.config.BudgetMiB {
			return nil, nil
		}
		available, err := b.commitAvailable()
		if err != nil {
			return nil, fmt.Errorf("read system commit headroom: %w", err)
		}
		if growth+amount+b.config.HeadroomMiB > available {
			return nil, nil
		}
	}
	r := &reservation{Phase: phase, MiB: amount}
	b.active[r] = true
	b.maxActive = max(b.maxActive, len(b.active))
	b.maxReservedMiB = max(b.maxReservedMiB, reserved+amount)
	s.MaxActive = max(s.MaxActive, samePhase+1)
	return r, nil
}

func (b *budget) observe(r *reservation, u usage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	r.CurrentMiB = u.CurrentMiB
	s := b.stats[r.Phase]
	s.PeakMiB = max(s.PeakMiB, u.PeakMiB)
	// Peak includes the wrapper and descendants. Allow 50% above the largest
	// observed peak and retain the cold-start floor for unseen packages.
	s.EstimateMiB = max(s.EstimateMiB, (u.PeakMiB*3+1)/2)
	if b.config.BudgetMiB > 0 {
		r.MiB = max(r.MiB, min(s.EstimateMiB, b.config.BudgetMiB), u.PeakMiB)
		reserved := int64(0)
		for active := range b.active {
			reserved += active.MiB
		}
		b.maxReservedMiB = max(b.maxReservedMiB, reserved)
	}
}

func (b *budget) release(r *reservation, wait, elapsed time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.active, r)
	s := b.stats[r.Phase]
	s.Processes++
	s.WaitMillis += wait.Milliseconds()
	s.ExecutionMillis += elapsed.Milliseconds()
}
