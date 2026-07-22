package linter

import (
	"testing"
	"time"
)

func TestTimingCollectorAddFile(t *testing.T) {
	c := NewTimingCollector()
	rules := []ConfiguredRule{{Name: "x"}, {Name: "y"}}
	c.addFile(rules, []time.Duration{time.Millisecond, 2 * time.Millisecond})
	c.addFile(rules[:1], []time.Duration{3 * time.Millisecond})

	got := c.Timings()
	if got["x"].Time != 4*time.Millisecond || got["x"].Files != 2 || got["x"].Kind != RuleKindNative {
		t.Errorf("x = %+v, want {Kind:native Time:4ms Files:2}", got["x"])
	}
	if got["y"].Time != 2*time.Millisecond || got["y"].Files != 1 || got["y"].Kind != RuleKindNative {
		t.Errorf("y = %+v, want {Kind:native Time:2ms Files:1}", got["y"])
	}
}

func TestTimingCollectorAddFileRuleTimesMS(t *testing.T) {
	c := NewTimingCollector()
	c.addFileRuleTimesMS(map[string]float64{"a": 1.5})
	c.addFileRuleTimesMS(map[string]float64{"a": 0.5, "b": 2})

	got := c.Timings()
	if got["a"].Time != 2*time.Millisecond || got["a"].Files != 2 || got["a"].Kind != RuleKindJS {
		t.Errorf("a = %+v, want {Kind:js Time:2ms Files:2}", got["a"])
	}
	if got["b"].Time != 2*time.Millisecond || got["b"].Files != 1 || got["b"].Kind != RuleKindJS {
		t.Errorf("b = %+v, want {Kind:js Time:2ms Files:1}", got["b"])
	}
}
