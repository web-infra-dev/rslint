package main

import (
	"strings"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/linter"
)

func TestTimingEnvEnabled(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"FALSE", false},
		{"1", true},
		{"all", true},
		{"true", true},
	}
	for _, c := range cases {
		if got := timingEnvEnabled(c.value); got != c.want {
			t.Errorf("timingEnvEnabled(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestFormatRuleTimingTableEmpty(t *testing.T) {
	if got := formatRuleTimingTable(nil); got != "" {
		t.Errorf("expected empty output for no timings, got %q", got)
	}
}

func TestFormatRuleTimingTable(t *testing.T) {
	table := formatRuleTimingTable(map[string]linter.RuleTiming{
		"no-console":                {Time: 500 * time.Microsecond, Files: 3},
		"@typescript-eslint/no-var": {Time: 1500 * time.Microsecond, Files: 2},
	})

	lines := strings.Split(strings.TrimSuffix(table, "\n"), "\n")
	// Leading blank line, header, separator, two rule rows, footnote.
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), table)
	}
	if !strings.Contains(lines[1], "Rule") ||
		!strings.Contains(lines[1], "Time (ms)") ||
		!strings.Contains(lines[1], "Files") ||
		!strings.Contains(lines[1], "Relative") {
		t.Errorf("unexpected header line: %q", lines[1])
	}
	// Sorted by time descending.
	if !strings.HasPrefix(lines[3], "@typescript-eslint/no-var") {
		t.Errorf("expected slowest rule first, got %q", lines[3])
	}
	if !strings.Contains(lines[3], "1.5") || !strings.Contains(lines[3], "75.0%") {
		t.Errorf("unexpected slowest-rule row: %q", lines[3])
	}
	if !strings.HasPrefix(lines[4], "no-console") {
		t.Errorf("expected no-console second, got %q", lines[4])
	}
	if !strings.Contains(lines[4], "0.5") || !strings.Contains(lines[4], "25.0%") ||
		!strings.Contains(lines[4], "3") {
		t.Errorf("unexpected no-console row: %q", lines[4])
	}
}
