package linter

import (
	"sync"
	"time"
)

// Rule implementation kinds reported in the timing table.
const (
	RuleKindNative = "native" // built-in Go rule
	RuleKindJS     = "js"     // rule run by the Node.js ESLint plugin worker
)

// RuleTiming aggregates one rule's execution cost across all linted files.
type RuleTiming struct {
	// Kind is RuleKindNative or RuleKindJS, depending on which side
	// executed the rule.
	Kind string
	// Time is the total time spent in the rule: building its listeners
	// (Run) plus every listener invocation during AST traversal, including
	// any diagnostics/fix construction those listeners perform. Files are
	// linted by parallel workers, so the sum over all rules can exceed the
	// run's wall-clock time.
	Time time.Duration
	// Files is the number of files the rule executed on.
	Files int
}

// TimingCollector accumulates per-rule timings across the concurrent lint
// workers. Workers merge once per linted file (not per node/listener), so
// lock contention stays negligible.
type TimingCollector struct {
	mu      sync.Mutex
	timings map[string]RuleTiming
}

func NewTimingCollector() *TimingCollector {
	return &TimingCollector{timings: make(map[string]RuleTiming)}
}

// addFile folds one file's per-rule durations (parallel to rules) into the
// aggregate.
func (c *TimingCollector) addFile(rules []ConfiguredRule, durations []time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, configuredRule := range rules {
		t := c.timings[configuredRule.Name]
		t.Kind = RuleKindNative
		t.Time += durations[i]
		t.Files++
		c.timings[configuredRule.Name] = t
	}
}

// addFileRuleTimesMS folds one file's per-rule times — in MILLISECONDS, as
// the Node plugin worker reports them on the wire — into the aggregate.
func (c *TimingCollector) addFileRuleTimesMS(times map[string]float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, ms := range times {
		t := c.timings[name]
		t.Kind = RuleKindJS
		t.Time += time.Duration(ms * float64(time.Millisecond))
		t.Files++
		c.timings[name] = t
	}
}

// Timings returns a snapshot of the accumulated per-rule timings.
func (c *TimingCollector) Timings() map[string]RuleTiming {
	c.mu.Lock()
	defer c.mu.Unlock()
	snapshot := make(map[string]RuleTiming, len(c.timings))
	for name, t := range c.timings {
		snapshot[name] = t
	}
	return snapshot
}
