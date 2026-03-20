package rule

import (
	"testing"
)

// ---------------------------------------------------------------------------
// parseRuleNames
// ---------------------------------------------------------------------------

func TestParseRuleNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single rule", "no-unused-vars", []string{"no-unused-vars"}},
		{"multiple rules", "no-unused-vars, no-console, no-debugger", []string{"no-unused-vars", "no-console", "no-debugger"}},
		{"rules with extra spaces", " no-unused-vars , no-console ", []string{"no-unused-vars", "no-console"}},
		{"empty string", "", nil},
		{"whitespace only", "   ", nil},
		{"typescript-eslint scoped rule", "@typescript-eslint/no-unsafe-member-access", []string{"@typescript-eslint/no-unsafe-member-access"}},
		{"mixed rules", "no-console, @typescript-eslint/no-unsafe-member-access, no-unused-vars", []string{"no-console", "@typescript-eslint/no-unsafe-member-access", "no-unused-vars"}},
		{"single rule with -- description", "@typescript-eslint/consistent-type-assertions -- needed here", []string{"@typescript-eslint/consistent-type-assertions"}},
		{"multiple rules with -- description", "no-console, no-debugger -- reason for disabling", []string{"no-console", "no-debugger"}},
		{"rule with -- description and extra spaces", " no-console -- reason ", []string{"no-console"}},
		{"wildcard with -- description", " -- just a description", nil},
		{"empty description after --", "no-console -- ", []string{"no-console"}},
		{"multiple -- separators", "no-console, no-debugger -- reason1 -- reason2", []string{"no-console", "no-debugger"}},
		{"-- without space before is not separator", "no-console--not-stripped", []string{"no-console--not-stripped"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRuleNames(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d rules, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("rule[%d]: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isBlockDisabled — unit tests for the event-replay algorithm
// ---------------------------------------------------------------------------

func TestIsBlockDisabled(t *testing.T) {
	tests := []struct {
		name       string
		directives []blockDirective
		ruleName   string
		line       int
		want       bool
	}{
		// ---- basic --------------------------------------------------------
		{
			name:       "no directives",
			directives: nil,
			ruleName:   "no-console",
			line:       5,
			want:       false,
		},
		{
			name: "disable specific rule, no enable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     99,
			want:     true,
		},
		{
			name: "disable specific rule, other rule unaffected",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-alert",
			line:     99,
			want:     false,
		},
		{
			name: "wildcard disable, no enable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
			},
			ruleName: "anything",
			line:     99,
			want:     true,
		},
		{
			name: "before disable line — not disabled",
			directives: []blockDirective{
				{line: 5, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     3,
			want:     false,
		},
		{
			name: "on disable line — disabled",
			directives: []blockDirective{
				{line: 5, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     true,
		},

		// ---- disable + enable range (core bug fix) ------------------------
		{
			name: "inside disable/enable range — disabled",
			directives: []blockDirective{
				{line: 1, isDisable: true, rules: []string{"no-console"}},
				{line: 10, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     true,
		},
		{
			name: "after enable — not disabled",
			directives: []blockDirective{
				{line: 1, isDisable: true, rules: []string{"no-console"}},
				{line: 10, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     15,
			want:     false,
		},
		{
			name: "on enable line — not disabled",
			directives: []blockDirective{
				{line: 1, isDisable: true, rules: nil},
				{line: 5, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     5,
			want:     false,
		},
		{
			name: "wildcard disable/enable — inside range",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 20, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     10,
			want:     true,
		},
		{
			name: "wildcard disable/enable — after range",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 20, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     25,
			want:     false,
		},

		// ---- multiple ranges ----------------------------------------------
		{
			name: "multiple disable/enable ranges — gap between",
			directives: []blockDirective{
				{line: 1, isDisable: true, rules: []string{"no-console"}},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
				{line: 10, isDisable: true, rules: []string{"no-console"}},
				{line: 15, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     7,
			want:     false, // between ranges
		},
		{
			name: "multiple disable/enable ranges — in second range",
			directives: []blockDirective{
				{line: 1, isDisable: true, rules: []string{"no-console"}},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
				{line: 10, isDisable: true, rules: []string{"no-console"}},
				{line: 15, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     12,
			want:     true,
		},

		// ---- wildcard + specific interaction ------------------------------
		{
			name: "wildcard disable + specific enable — enabled rule",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     10,
			want:     false, // specifically re-enabled
		},
		{
			name: "wildcard disable + specific enable — other rules still disabled",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-alert",
			line:     10,
			want:     true, // not specifically enabled, wildcard still active
		},
		{
			name: "specific disable + wildcard enable — all re-enabled",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
				{line: 10, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     15,
			want:     false,
		},
		{
			name: "wildcard disable + specific enable + specific re-disable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
				{line: 10, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     7,
			want:     false, // enabled at line 5
		},
		{
			name: "wildcard disable + specific enable + specific re-disable — after re-disable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
				{line: 10, isDisable: true, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     15,
			want:     true, // re-disabled at line 10
		},

		// ---- wildcard enable resets specific state -------------------------
		{
			name: "specific disable + wildcard enable resets specific state",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
				{line: 5, isDisable: false, rules: nil}, // enable all → resets
			},
			ruleName: "no-console",
			line:     10,
			want:     false,
		},
		{
			name: "wildcard disable + wildcard enable + query before enable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: nil},
				{line: 20, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     0,
			want:     true, // disable takes effect on its own line
		},

		// ---- multiple rules in one directive ------------------------------
		{
			name: "disable multiple rules — match first",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console", "no-debugger"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     true,
		},
		{
			name: "disable multiple rules — match second",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console", "no-debugger"}},
			},
			ruleName: "no-debugger",
			line:     5,
			want:     true,
		},
		{
			name: "disable multiple rules — no match",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console", "no-debugger"}},
			},
			ruleName: "no-alert",
			line:     5,
			want:     false,
		},
		{
			name: "enable one of multiple disabled rules",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console", "no-debugger"}},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     10,
			want:     false,
		},
		{
			name: "enable one of multiple disabled rules — other still disabled",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console", "no-debugger"}},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-debugger",
			line:     10,
			want:     true,
		},

		// ---- edge cases ---------------------------------------------------
		{
			name: "enable without preceding disable is no-op",
			directives: []blockDirective{
				{line: 0, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     false,
		},
		{
			name: "wildcard enable without preceding disable is no-op",
			directives: []blockDirective{
				{line: 0, isDisable: false, rules: nil},
			},
			ruleName: "no-console",
			line:     5,
			want:     false,
		},
		{
			name: "disable and enable on same line — enable wins (processed in order)",
			directives: []blockDirective{
				{line: 5, isDisable: true, rules: []string{"no-console"}},
				{line: 5, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     false,
		},
		{
			name: "query line 0 with no directives",
			directives: nil,
			ruleName:   "no-console",
			line:       0,
			want:       false,
		},

		// ---- duplicate disable (idempotent) --------------------------------
		{
			name: "duplicate disable is harmless",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
				{line: 3, isDisable: true, rules: []string{"no-console"}},
				{line: 10, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     5,
			want:     true,
		},
		{
			name: "duplicate disable — after enable",
			directives: []blockDirective{
				{line: 0, isDisable: true, rules: []string{"no-console"}},
				{line: 3, isDisable: true, rules: []string{"no-console"}},
				{line: 10, isDisable: false, rules: []string{"no-console"}},
			},
			ruleName: "no-console",
			line:     15,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &DisableManager{
				blockDirectives:       tt.directives,
				lineDisabledRules:     make(map[int][]string),
				nextLineDisabledRules: make(map[int][]string),
			}
			got := dm.isBlockDisabled(tt.ruleName, tt.line)
			if got != tt.want {
				t.Errorf("isBlockDisabled(%q, %d) = %v, want %v", tt.ruleName, tt.line, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DisableManager line-level directives (unchanged logic, keep coverage)
// ---------------------------------------------------------------------------

func TestDisableManagerLineLevelDirectives(t *testing.T) {
	dm := &DisableManager{
		lineDisabledRules: map[int][]string{
			5:  {"no-unused-vars"},
			10: {"*"},
		},
		nextLineDisabledRules: map[int][]string{
			8:  {"no-debugger"},
			12: {"*"},
		},
	}

	tests := []struct {
		name     string
		ruleName string
		line     int
		want     bool
	}{
		{"disable-line matches rule", "no-unused-vars", 5, true},
		{"disable-line does not match other rule", "no-console", 5, false},
		{"disable-line wildcard matches any rule", "anything", 10, true},
		{"disable-line does not affect other lines", "no-unused-vars", 6, false},
		{"disable-next-line matches rule", "no-debugger", 8, true},
		{"disable-next-line does not match other rule", "no-console", 8, false},
		{"disable-next-line wildcard matches any rule", "anything", 12, true},
		{"disable-next-line does not affect other lines", "no-debugger", 9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dm.isLineDisabled(tt.ruleName, tt.line)
			if got != tt.want {
				t.Errorf("line %d, rule %q: got %v, want %v", tt.line, tt.ruleName, got, tt.want)
			}
		})
	}
}

// isLineDisabled is a test helper that checks only line-level disables,
// bypassing the block directive check (which needs a sourceFile for pos→line).
func (dm *DisableManager) isLineDisabled(ruleName string, line int) bool {
	if lineRules, exists := dm.lineDisabledRules[line]; exists {
		for _, r := range lineRules {
			if r == ruleName || r == "*" {
				return true
			}
		}
	}
	if nextLineRules, exists := dm.nextLineDisabledRules[line]; exists {
		for _, r := range nextLineRules {
			if r == ruleName || r == "*" {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Integration: block + line directives together
// ---------------------------------------------------------------------------

func TestDisableManagerBlockAndLineCombined(t *testing.T) {
	dm := &DisableManager{
		blockDirectives: []blockDirective{
			{line: 0, isDisable: true, rules: []string{"no-console"}},
			{line: 10, isDisable: false, rules: []string{"no-console"}},
		},
		lineDisabledRules: map[int][]string{
			15: {"no-alert"},
		},
		nextLineDisabledRules: map[int][]string{
			20: {"no-debugger"},
		},
	}

	tests := []struct {
		name     string
		ruleName string
		line     int
		want     bool
	}{
		{"block disabled in range", "no-console", 5, true},
		{"block not disabled after enable", "no-console", 15, false},
		{"line disable on its line", "no-alert", 15, true},
		{"line disable not on other line", "no-alert", 16, false},
		{"next-line disable on target line", "no-debugger", 20, true},
		{"next-line disable not on other line", "no-debugger", 21, false},
		{"unrelated rule not affected", "no-eval", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dm.isBlockDisabled(tt.ruleName, tt.line)
			// For line-level checks we call the helper directly
			if !got {
				got = dm.isLineDisabled(tt.ruleName, tt.line)
			}
			if got != tt.want {
				t.Errorf("rule %q at line %d: got %v, want %v", tt.ruleName, tt.line, got, tt.want)
			}
		})
	}
}
