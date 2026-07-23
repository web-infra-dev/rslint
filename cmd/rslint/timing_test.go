package main

import (
	"strings"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/linter"
)

func TestTimingFlagSet(t *testing.T) {
	cases := []struct {
		value       string
		wantErr     bool
		wantEnabled bool
		wantLimit   int
	}{
		{value: "true", wantEnabled: true},
		{value: "all", wantEnabled: true},
		{value: "ALL", wantEnabled: true},
		{value: "false"},
		{value: "10", wantEnabled: true, wantLimit: 10},
		{value: "0", wantErr: true},
		{value: "-3", wantErr: true},
		{value: "ten", wantErr: true},
	}
	for _, c := range cases {
		var enabled bool
		var limit int
		err := (&timingFlag{enabled: &enabled, limit: &limit}).Set(c.value)
		if (err != nil) != c.wantErr {
			t.Errorf("Set(%q) error = %v, wantErr %v", c.value, err, c.wantErr)
			continue
		}
		if err == nil && (enabled != c.wantEnabled || limit != c.wantLimit) {
			t.Errorf("Set(%q) = enabled %v limit %d, want enabled %v limit %d",
				c.value, enabled, limit, c.wantEnabled, c.wantLimit)
		}
	}
}

func TestFormatRuleTimingTableEmpty(t *testing.T) {
	if got := formatRuleTimingTable(nil, 0); got != "" {
		t.Errorf("expected empty output for no timings, got %q", got)
	}
}

func TestFormatRuleTimingTable(t *testing.T) {
	table := formatRuleTimingTable(map[string]linter.RuleTiming{
		"no-console":                {Kind: linter.RuleKindNative, Time: 500 * time.Microsecond, Files: 3},
		"@typescript-eslint/no-var": {Kind: linter.RuleKindJS, Time: 1500 * time.Microsecond, Files: 2},
	}, 0)

	lines := strings.Split(strings.TrimSuffix(table, "\n"), "\n")
	// Leading blank line, header, separator, two rule rows, footnote.
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), table)
	}
	if !strings.Contains(lines[1], "Rule") ||
		!strings.Contains(lines[1], "Source") ||
		!strings.Contains(lines[1], "Time (ms)") ||
		!strings.Contains(lines[1], "Files") ||
		!strings.Contains(lines[1], "Relative") {
		t.Errorf("unexpected header line: %q", lines[1])
	}
	// Sorted by time descending.
	if !strings.HasPrefix(lines[3], "@typescript-eslint/no-var") {
		t.Errorf("expected slowest rule first, got %q", lines[3])
	}
	if !strings.Contains(lines[3], "js") || !strings.Contains(lines[3], "1.5") ||
		!strings.Contains(lines[3], "75.0%") {
		t.Errorf("unexpected slowest-rule row: %q", lines[3])
	}
	if !strings.HasPrefix(lines[4], "no-console") {
		t.Errorf("expected no-console second, got %q", lines[4])
	}
	if !strings.Contains(lines[4], "native") || !strings.Contains(lines[4], "0.5") ||
		!strings.Contains(lines[4], "25.0%") || !strings.Contains(lines[4], "3") {
		t.Errorf("unexpected no-console row: %q", lines[4])
	}
}

func TestFormatRuleTimingTableLimit(t *testing.T) {
	table := formatRuleTimingTable(map[string]linter.RuleTiming{
		"no-console":                {Kind: linter.RuleKindNative, Time: 500 * time.Microsecond, Files: 3},
		"@typescript-eslint/no-var": {Kind: linter.RuleKindJS, Time: 1500 * time.Microsecond, Files: 2},
	}, 1)

	if strings.Contains(table, "no-console") {
		t.Errorf("expected no-console to be cut by limit 1:\n%s", table)
	}
	// Relative stays the share of ALL rule time, not just the shown rows.
	if !strings.Contains(table, "@typescript-eslint/no-var") || !strings.Contains(table, "75.0%") {
		t.Errorf("unexpected top-1 table:\n%s", table)
	}
}
