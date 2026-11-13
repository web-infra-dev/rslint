package rule

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/testutil/directivefixtures"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestDisableManagerSameLineBoundaries(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, testCase := range []struct {
			name          string
			disable       string
			enable        string
			otherDisabled bool
		}{
			{"specific", " react/no-namespace", " react/no-namespace", false},
			{"wildcard then specific", "", " react/no-namespace", true},
			{"specific then wildcard", " react/no-namespace", "", false},
		} {
			t.Run(prefix+"/"+testCase.name, func(t *testing.T) {
				code := "/* 😀 */ before(); /* " + prefix + "-disable" + testCase.disable + " */ inside(); /* " + prefix + "-enable" + testCase.enable + " */ after();"
				sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/directives.js", Path: "/directives.js"}, code, core.ScriptKindJS)
				dm := NewDisableManager(sf, NewCommentStore(sf))
				for _, query := range []struct {
					marker string
					want   bool
				}{
					{"before()", false}, {"inside()", true}, {"after()", false},
				} {
					pos := strings.Index(code, query.marker)
					if pos < 0 {
						t.Fatalf("missing marker %q", query.marker)
					}
					if got := dm.IsRuleDisabled("react/no-namespace", pos); got != query.want {
						t.Errorf("%s: disabled = %v, want %v", query.marker, got, query.want)
					}
				}
				if got := dm.IsRuleDisabled("no-console", strings.Index(code, "after()")); got != testCase.otherDisabled {
					t.Errorf("unrelated rule disabled = %v, want %v", got, testCase.otherDisabled)
				}
			})
		}
	}
}

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
// matchDirective
// ---------------------------------------------------------------------------

func TestMatchDirective(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedKind  directiveKind
		expectedRules []string
	}{
		// ---- rslint- prefix: block disable/enable -------------------------
		{"rslint-disable wildcard", "rslint-disable", directiveBlock, nil},
		{"rslint-disable single rule", "rslint-disable no-console", directiveBlock, []string{"no-console"}},
		{"rslint-disable multiple rules", "rslint-disable no-console, no-alert", directiveBlock, []string{"no-console", "no-alert"}},
		{"rslint-disable scoped rule", "rslint-disable @typescript-eslint/no-explicit-any", directiveBlock, []string{"@typescript-eslint/no-explicit-any"}},
		{"rslint-disable with -- description", "rslint-disable no-console -- reason", directiveBlock, []string{"no-console"}},
		{"rslint-disable wildcard with -- description", "rslint-disable -- reason", directiveBlock, nil},
		{"rslint-enable wildcard", "rslint-enable", directiveEnable, nil},
		{"rslint-enable single rule", "rslint-enable no-console", directiveEnable, []string{"no-console"}},
		{"rslint-enable multiple rules", "rslint-enable no-console, no-alert", directiveEnable, []string{"no-console", "no-alert"}},

		// ---- rslint- prefix: line directives ------------------------------
		{"rslint-disable-line wildcard", "rslint-disable-line", directiveLine, nil},
		{"rslint-disable-line single rule", "rslint-disable-line no-console", directiveLine, []string{"no-console"}},
		{"rslint-disable-line multiple rules", "rslint-disable-line no-console, no-alert", directiveLine, []string{"no-console", "no-alert"}},
		{"rslint-disable-line with -- description", "rslint-disable-line no-console -- reason", directiveLine, []string{"no-console"}},
		{"rslint-disable-next-line wildcard", "rslint-disable-next-line", directiveNextLine, nil},
		{"rslint-disable-next-line single rule", "rslint-disable-next-line no-console", directiveNextLine, []string{"no-console"}},
		{"rslint-disable-next-line multiple rules", "rslint-disable-next-line no-console, no-alert", directiveNextLine, []string{"no-console", "no-alert"}},
		{"rslint-disable-next-line with -- description", "rslint-disable-next-line no-console -- reason", directiveNextLine, []string{"no-console"}},
		{"rslint-disable-next-line scoped rule", "rslint-disable-next-line @typescript-eslint/no-explicit-any", directiveNextLine, []string{"@typescript-eslint/no-explicit-any"}},

		// ---- eslint- prefix: block disable/enable -------------------------
		{"eslint-disable wildcard", "eslint-disable", directiveBlock, nil},
		{"eslint-disable single rule", "eslint-disable no-console", directiveBlock, []string{"no-console"}},
		{"eslint-disable multiple rules", "eslint-disable no-console, no-alert", directiveBlock, []string{"no-console", "no-alert"}},
		{"eslint-disable scoped rule", "eslint-disable @typescript-eslint/no-explicit-any", directiveBlock, []string{"@typescript-eslint/no-explicit-any"}},
		{"eslint-disable with -- description", "eslint-disable no-console -- reason", directiveBlock, []string{"no-console"}},
		{"eslint-disable wildcard with -- description", "eslint-disable -- reason", directiveBlock, nil},
		{"eslint-enable wildcard", "eslint-enable", directiveEnable, nil},
		{"eslint-enable single rule", "eslint-enable no-console", directiveEnable, []string{"no-console"}},
		{"eslint-enable multiple rules", "eslint-enable no-console, no-alert", directiveEnable, []string{"no-console", "no-alert"}},

		// ---- eslint- prefix: line directives ------------------------------
		{"eslint-disable-line wildcard", "eslint-disable-line", directiveLine, nil},
		{"eslint-disable-line single rule", "eslint-disable-line no-console", directiveLine, []string{"no-console"}},
		{"eslint-disable-line multiple rules", "eslint-disable-line no-console, no-alert", directiveLine, []string{"no-console", "no-alert"}},
		{"eslint-disable-line with -- description", "eslint-disable-line no-console -- reason", directiveLine, []string{"no-console"}},
		{"eslint-disable-next-line wildcard", "eslint-disable-next-line", directiveNextLine, nil},
		{"eslint-disable-next-line single rule", "eslint-disable-next-line no-console", directiveNextLine, []string{"no-console"}},
		{"eslint-disable-next-line multiple rules", "eslint-disable-next-line no-console, no-alert", directiveNextLine, []string{"no-console", "no-alert"}},
		{"eslint-disable-next-line with -- description", "eslint-disable-next-line no-console -- reason", directiveNextLine, []string{"no-console"}},

		// ---- non-directive comments (should return directiveNone) ---------
		{"empty string", "", directiveNone, nil},
		{"random comment", "some random comment", directiveNone, nil},
		{"tslint prefix not recognized", "tslint-disable no-console", directiveNone, nil},
		{"partial prefix eslint", "eslint no-console", directiveNone, nil},
		{"partial prefix rslint", "rslint no-console", directiveNone, nil},
		{"eslint- without disable/enable", "eslint-check no-console", directiveNone, nil},
		{"rslint- without disable/enable", "rslint-check no-console", directiveNone, nil},

		// ---- edge cases ---------------------------------------------------
		{"rslint-enable with -- description", "rslint-enable no-console -- reason", directiveEnable, []string{"no-console"}},
		{"eslint-enable with -- description", "eslint-enable no-console -- reason", directiveEnable, []string{"no-console"}},
		{"multiple rules with extra spaces", "rslint-disable  no-console ,  no-alert ", directiveBlock, []string{"no-console", "no-alert"}},
		{"rule name with multiple -- separators", "rslint-disable no-console -- reason1 -- reason2", directiveBlock, []string{"no-console"}},
		{"-- without space is not separator", "rslint-disable no-console--not-stripped", directiveBlock, []string{"no-console--not-stripped"}},

		// ---- case sensitivity (directives are case-sensitive) -------------
		{"uppercase RSLINT-DISABLE not recognized", "RSLINT-DISABLE no-console", directiveNone, nil},
		{"uppercase ESLINT-DISABLE not recognized", "ESLINT-DISABLE no-console", directiveNone, nil},
		{"mixed case Rslint-Disable not recognized", "Rslint-Disable no-console", directiveNone, nil},
		{"mixed case Eslint-Enable not recognized", "Eslint-Enable no-console", directiveNone, nil},

		// ---- no space between prefix and rule name ------------------------
		{"rslint-disable no space before rule", "rslint-disable" + "no-console", directiveBlock, []string{"no-console"}}, // cspell:disable-line
		{"eslint-disable no space before rule", "eslint-disable" + "no-console", directiveBlock, []string{"no-console"}}, // cspell:disable-line

		// ---- typos / near-miss prefixes -----------------------------------
		{"rslint-disabled (extra d)", "rslint-disabled no-console", directiveBlock, []string{"d no-console"}},
		{"eslint-disabled (extra d)", "eslint-disabled no-console", directiveBlock, []string{"d no-console"}},

		// ---- enable-line / enable-next-line (not valid ESLint directives) --
		{"rslint-enable-line treated as enable rule -line", "rslint-enable-line", directiveEnable, []string{"-line"}},
		{"rslint-enable-next-line treated as enable rule -next-line", "rslint-enable-next-line", directiveEnable, []string{"-next-line"}},
		{"eslint-enable-line treated as enable rule -line", "eslint-enable-line", directiveEnable, []string{"-line"}},

		// ---- whitespace variants ------------------------------------------
		{"tab between prefix and rule", "rslint-disable\tno-console", directiveBlock, []string{"no-console"}},
		{"only whitespace after prefix is wildcard", "rslint-disable   ", directiveBlock, nil},
		{"only tab after prefix is wildcard", "eslint-disable\t", directiveBlock, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, rules, _ := matchDirective(tt.input)
			if kind != tt.expectedKind {
				t.Errorf("kind: expected %d, got %d", tt.expectedKind, kind)
			}
			if len(rules) != len(tt.expectedRules) {
				t.Fatalf("expected %d rules, got %d: %v", len(tt.expectedRules), len(rules), rules)
			}
			for i, expected := range tt.expectedRules {
				if rules[i] != expected {
					t.Errorf("rule[%d]: expected %q, got %q", i, expected, rules[i])
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
			name:       "query line 0 with no directives",
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
			dm := managerForDirectiveEvents(t, tt.directives, nil, nil)
			got := dm.IsRuleDisabled(tt.ruleName, int(dm.sourceFile.ECMALineMap()[tt.line+1])-1)
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
	dm := managerForDirectiveEvents(t, nil, map[int][]string{5: {"no-unused-vars"}, 10: {"*"}}, map[int][]string{8: {"no-debugger"}, 12: {"*"}})

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
			got := dm.IsRuleDisabled(tt.ruleName, int(dm.sourceFile.ECMALineMap()[tt.line+1])-1)
			if got != tt.want {
				t.Errorf("line %d, rule %q: got %v, want %v", tt.line, tt.ruleName, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Integration: block + line directives together
// ---------------------------------------------------------------------------

func TestDisableManagerBlockAndLineCombined(t *testing.T) {
	dm := managerForDirectiveEvents(t, []blockDirective{{line: 0, isDisable: true, rules: []string{"no-console"}}, {line: 10, rules: []string{"no-console"}}}, map[int][]string{15: {"no-alert"}}, map[int][]string{20: {"no-debugger"}})

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
			got := dm.IsRuleDisabled(tt.ruleName, int(dm.sourceFile.ECMALineMap()[tt.line+1])-1)
			if got != tt.want {
				t.Errorf("rule %q at line %d: got %v, want %v", tt.ruleName, tt.line, got, tt.want)
			}
		})
	}
}

func TestDisableManagerBlockDirectivesInLineComments(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, tracking := range []bool{false, true} {
			for _, tc := range []struct {
				name, source     string
				report, disabled bool
				unused           int
			}{
				{"specific disable", "// eslint-disable a\nx();", true, true, 0},
				{"wildcard disable", "// eslint-disable\nx();", true, true, 0},
				{"specific enable", "/* eslint-disable a */\n// eslint-enable a\nx();", true, false, 1},
				{"wildcard enable", "/* eslint-disable */\n// eslint-enable\nx();", true, false, 1},
				{"unused disable", "// eslint-disable a\nx();", false, false, 1},
				{"unmatched enable", "// eslint-enable a\nx();", false, false, 1},
			} {
				t.Run(fmt.Sprintf("%s/tracking=%v/%s", prefix, tracking, tc.name), func(t *testing.T) {
					source := strings.ReplaceAll(tc.source, "eslint-", prefix+"-")
					sf := parseCommentStoreSource(t, source)
					manager := NewSuppressionManager(sf, NewCommentStore(sf))
					if tracking {
						manager = NewDisableManager(sf, NewCommentStore(sf))
					}
					if tc.report {
						if got := manager.IsRuleDisabled("a", strings.Index(source, "x();")); got != tc.disabled {
							t.Fatalf("disabled = %v, want %v", got, tc.disabled)
						}
					}
					if !tracking {
						return
					}
					diagnostics := manager.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
					if len(diagnostics) != tc.unused {
						t.Fatalf("unused diagnostics = %d, want %d: %+v", len(diagnostics), tc.unused, diagnostics)
					}
					for _, diagnostic := range diagnostics {
						if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 {
							t.Fatalf("expected one removal fix: %+v", diagnostic)
						}
						fix := (*diagnostic.FixesPtr)[0]
						if fix.Range.Pos() != 0 || fix.Range.End() != strings.IndexByte(source, '\n') || fix.Text != " " {
							t.Fatalf("unexpected comment removal: %+v", fix)
						}
					}
				})
			}
		}
	}
}

func TestUnusedDisableDirectives(t *testing.T) {
	for _, tc := range []struct {
		name, source, usedRule, fixed string
		count                         int
	}{
		{"unused wildcard", "/* eslint-disable */\nx();", "", " \nx();", 1},
		{"last repeated directive wins", "/* eslint-disable a */\n/* eslint-disable a */\nx();", "a", " \n/* eslint-disable a */\nx();", 1},
		{"block precedes line", "/* eslint-disable a */\nx(); // eslint-disable-line a", "a", "/* eslint-disable a */\nx();  ", 1},
		{"partial first", "/* eslint-disable a, b -- reason */\nx();", "b", "/* eslint-disable b -- reason */\nx();", 1},
		{"rule name substring", "/* eslint-disable bar, a -- reason */\nx();", "bar", "/* eslint-disable bar -- reason */\nx();", 1},
		{"partial last", "/* eslint-disable a, b -- reason */\nx();", "a", "/* eslint-disable a -- reason */\nx();", 1},
		{"whole LF", "/* eslint-disable a\n */\nx();", "", " \n\nx();", 1},
		{"whole CRLF", "/* eslint-disable a\r\n */\nx();", "", "\r\n\nx();", 1},
		{"whole separated CR LF", "/* eslint-disable a\r -- reason\n */\nx();", "", "\r \n\nx();", 1},
		{"whole preceding CR", "\r/* eslint-disable a\n */x();", "", "\r \nx();", 1},
		{"whole following LF", "/* eslint-disable a\r */\nx();", "", "\r \nx();", 1},
		{"whole both boundaries", "\r/* eslint-disable a\n -- reason\r */\nx();", "", "\r \n\r \nx();", 1},
		{"partial first LF", "/* eslint-disable a,\nb */\nx();", "b", "/* eslint-disable  \nb */\nx();", 1},
		{"partial first CRLF", "/* eslint-disable a,\r\nb */\nx();", "b", "/* eslint-disable \r\nb */\nx();", 1},
		{"partial last LF", "/* eslint-disable b\n, a */\nx();", "b", "/* eslint-disable b\n */\nx();", 1},
		{"partial last CRLF", "/* eslint-disable b\r\n, a */\nx();", "b", "/* eslint-disable b\r\n */\nx();", 1},
		{"partial first preceding CR", "/* eslint-disable \ra,\nb */\nx();", "b", "/* eslint-disable \r \nb */\nx();", 1},
		{"partial last separated CR LF", "/* eslint-disable b\r,\na */\nx();", "b", "/* eslint-disable b\r \n */\nx();", 1},
		{"no space partial first", "/* eslint-disable" + "a, b */\nx();", "b", "/* eslint-disable" + "b */\nx();", 1},
		{"no space partial last", "/* eslint-disable" + "a, b */\nx();", "a", "/* eslint-disable" + "a */\nx();", 1},
		{"rslint no space partial first", "/* rslint-disable" + "a, b */\nx();", "b", "/* rslint-disable" + "b */\nx();", 1},
		{"rslint no space partial last", "/* rslint-disable" + "a, b */\nx();", "a", "/* rslint-disable" + "a */\nx();", 1},
		{"next line no space", "// eslint-disable-next-line" + "a, b -- reason\nx();", "b", "// eslint-disable-next-line" + "b -- reason\nx();", 1},
		{"same line no space", "x(); // rslint-disable-line" + "a,b", "a", "x(); // rslint-disable-line" + "a", 1},
		{"enable no space", "/* eslint-disable a */\nx();\n/* eslint-enable" + "a,b */", "a", "/* eslint-disable a */\nx();\n/* eslint-enable" + "a */", 1},
		{"empty explanation", "/* eslint-disable a, b -- */\nx();", "a", "/* eslint-disable a -- */\nx();", 1},
		{"used empty explanation", "/* eslint-disable a -- */\nx();", "a", "", 0},
		{"unused enable", "/* eslint-enable a */\nx();", "", " \nx();", 1},
		{"matched enable", "/* eslint-disable a */\nx();\n/* eslint-enable a */", "a", "", 0},
		{"rslint next line", "// rslint-disable-next-line a\nx();", "a", "", 0},
		{"multiline next line", "/* eslint-disable-next-line\na */\nx();", "a", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sf := parseCommentStoreSource(t, tc.source)
			dm := NewDisableManager(sf, NewCommentStore(sf))
			if tc.usedRule != "" && !dm.IsRuleDisabled(tc.usedRule, strings.Index(tc.source, "x();")) {
				t.Fatal("expected suppression")
			}
			diagnostics := dm.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
			if len(diagnostics) != tc.count {
				t.Fatalf("got %d diagnostics, want %d: %#v", len(diagnostics), tc.count, diagnostics)
			}
			if tc.count == 1 {
				d := diagnostics[0]
				if d.RuleName != "" || d.Severity != SeverityWarning {
					t.Fatalf("wrong linter diagnostic: %#v", d)
				}
				f := d.Fixes()[0]
				fixed := tc.source[:f.Range.Pos()] + f.Text + tc.source[f.Range.End():]
				if fixed != tc.fixed {
					t.Fatalf("fixed = %q, want %q", fixed, tc.fixed)
				}
			}
		})
	}
}

// blockDirective represents a block-level disable or enable event at a source location
type blockDirective struct {
	line      int
	isDisable bool     // true = disable, false = enable
	rules     []string // nil means all rules (wildcard)
}

func managerForDirectiveEvents(t *testing.T, events []blockDirective, lines, nextLines map[int][]string) *DisableManager {
	t.Helper()
	sourceLines := make([]string, 1024)
	formatRules := func(names []string) string {
		if len(names) == 1 && names[0] == "*" {
			return ""
		}
		return strings.Join(names, ", ")
	}
	for _, e := range events {
		kind := "enable"
		if e.isDisable {
			kind = "disable"
		}
		sourceLines[e.line] += "/* eslint-" + kind + " " + formatRules(e.rules) + " */"
	}
	for line, names := range lines {
		sourceLines[line] += "/* eslint-disable-line " + formatRules(names) + " */"
	}
	for line, names := range nextLines {
		sourceLines[line-1] += "/* eslint-disable-next-line " + formatRules(names) + " */"
	}
	sf := parseCommentStoreSource(t, strings.Join(sourceLines, "\n"))
	// Force the lazily computed map before the table makes positional queries.
	scanner.GetECMALineAndUTF16CharacterOfPosition(sf, 0)
	return NewDisableManager(sf, NewCommentStore(sf))
}

// Fixtures use ESLint 10.14.0's pinned directive engine; see the generator.
func TestUnusedDirectiveESLintConformance(t *testing.T) {
	var cases []struct {
		Name        string   `json:"name"`
		Source      string   `json:"source"`
		Fixed       string   `json:"fixed"`
		Reports     []string `json:"reports"`
		Diagnostics []struct {
			Message  string `json:"message"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
			Severity int    `json:"severity"`
			Fix      struct {
				Range [2]int `json:"range"`
				Text  string `json:"text"`
			} `json:"fix"`
		} `json:"diagnostics"`
	}
	data, err := os.ReadFile("testdata/unused-directives.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty conformance fixture")
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sf := parseCommentStoreSource(t, tc.Source)
			dm := NewDisableManager(sf, NewCommentStore(sf))
			for _, name := range tc.Reports {
				dm.IsRuleDisabled(name, strings.Index(tc.Source, "hit();"))
			}
			diagnostics := dm.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
			sort.SliceStable(diagnostics, func(i, j int) bool { return diagnostics[i].Range.Pos() < diagnostics[j].Range.Pos() })
			if len(diagnostics) != len(tc.Diagnostics) {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(diagnostics), len(tc.Diagnostics), diagnostics)
			}
			byteOffset := func(offset int) int {
				units := 0
				for pos, r := range tc.Source {
					if units == offset {
						return pos
					}
					units++
					if r > 0xffff {
						units++
					}
				}
				if units == offset {
					return len(tc.Source)
				}
				t.Fatalf("invalid UTF-16 offset %d", offset)
				return 0
			}
			var fixes []RuleFix
			for i, d := range diagnostics {
				want := tc.Diagnostics[i]
				line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, d.Range.Pos())
				if d.Message.Description != want.Message || line+1 != want.Line || int(column)+1 != want.Column || d.Severity != SeverityWarning || want.Severity != 1 {
					t.Errorf("diagnostic %d: %+v at %d:%d; want %+v", i, d, line+1, column+1, want)
				}
				fix := d.Fixes()[0]
				if fix.Range.Pos() != byteOffset(want.Fix.Range[0]) || fix.Range.End() != byteOffset(want.Fix.Range[1]) || fix.Text != want.Fix.Text {
					t.Errorf("fix: %+v, want %+v", fix, want.Fix)
				}
				fixes = append(fixes, fix)
			}
			sort.SliceStable(fixes, func(i, j int) bool { return fixes[i].Range.Pos() < fixes[j].Range.Pos() })
			var output strings.Builder
			end := 0
			for _, fix := range fixes {
				if fix.Range.Pos() < end {
					continue
				}
				output.WriteString(tc.Source[end:fix.Range.Pos()])
				output.WriteString(fix.Text)
				end = fix.Range.End()
			}
			output.WriteString(tc.Source[end:])
			if output.String() != tc.Fixed {
				t.Errorf("fixed %q, want %q", output.String(), tc.Fixed)
			}
		})
	}
}

func TestUnusedDirectiveSkippedRules(t *testing.T) {
	source := "/* eslint-disable skipped, off */\n/* eslint-disable */\nx();"
	sf := parseCommentStoreSource(t, source)
	dm := NewDisableManager(sf, NewCommentStore(sf))
	diagnostics := dm.UnusedDiagnostics("/test.ts", SeverityError, EditDemandAutofix, map[string]bool{"skipped": true})
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message.Description, "'off'") {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if got := dm.UnusedDiagnostics("/test.ts", SeverityOff, EditDemandAutofix, nil); len(got) != 0 {
		t.Fatal("off produced unused diagnostics")
	}
}

func TestUnusedDirectiveUsageLocations(t *testing.T) {
	name := "a"
	for _, tc := range []struct {
		name, source string
		usage        DisableDirectiveUsage
	}{
		{"partial with Unicode and CRLF", "// 😀\r\nconst x = 1; /* eslint-disable a, b */\r\nx();", DisableDirectiveUsage{Line: 1, Column: 13, RuleID: &name}},
		{"UTF-16 column", "\"😀\"; /* eslint-disable a, b */\nx();", DisableDirectiveUsage{Line: 0, Column: 6, RuleID: &name}},
		{"wildcard", "/* eslint-disable */\nx();", DisableDirectiveUsage{Line: 0, Column: 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sf := parseCommentStoreSource(t, tc.source)
			manager := NewDisableManager(sf, NewCommentStore(sf))
			manager.Freeze(sf)
			if err := manager.MarkUsedDirectives([]DisableDirectiveUsage{tc.usage, tc.usage}); err != nil {
				t.Fatal(err)
			}
			diagnostics := manager.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
			if tc.usage.RuleID == nil {
				if len(diagnostics) != 0 {
					t.Fatalf("used wildcard reported unused: %+v", diagnostics)
				}
			} else if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message.Description, "'b'") || strings.Contains(diagnostics[0].Message.Description, "'a'") {
				t.Fatalf("wrong partial usage: %+v", diagnostics)
			}
		})
	}
}

func TestUnusedDirectiveUsageRejectsUnknownEntriesAtomically(t *testing.T) {
	source := "/* eslint-disable a, b */\nx();\n/* eslint-enable a */"
	sf := parseCommentStoreSource(t, source)
	name := "a"
	for _, invalid := range []DisableDirectiveUsage{
		{Line: -1, Column: 0, RuleID: &name},
		{Line: 0, Column: 1, RuleID: &name},
		{Line: 2, Column: 0, RuleID: &name},
		{Line: 0, Column: 0},
	} {
		manager := NewDisableManager(sf, NewCommentStore(sf))
		if err := manager.MarkUsedDirectives([]DisableDirectiveUsage{{Line: 0, Column: 0, RuleID: &name}, invalid}); err == nil {
			t.Fatalf("accepted invalid summary: %+v", invalid)
		}
		diagnostics := manager.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
		if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message.Description, "'a' or 'b'") {
			t.Fatalf("invalid summary partially marked usage: %+v", diagnostics)
		}
	}
}

func TestUnusedDirectiveUpstreamConformance(t *testing.T) {
	corpus := directivefixtures.Load(t, "testdata/eslint-directives/upstream.json")
	if corpus.Manifest.UpstreamTests != 89 || len(corpus.Cases) != 88 || len(corpus.Manifest.Exceptions) != 1 {
		t.Fatal("upstream engine inventory changed")
	}
	runDirectiveConformance(t, corpus)
}

func TestUnusedDirectiveSourceConformance(t *testing.T) {
	corpus := directivefixtures.Load(t, "testdata/eslint-directives/source.json")
	if corpus.Manifest.UpstreamTests != 486 {
		t.Fatal("source test inventory changed")
	}
	runDirectiveConformance(t, corpus)
}

func runDirectiveConformance(t *testing.T, corpus directivefixtures.Corpus) {
	t.Helper()

	for index, tc := range corpus.Cases {
		t.Run(fmt.Sprintf("%03d/%s", index, tc.Name), func(t *testing.T) {
			for _, prefix := range []string{"eslint", "rslint"} {
				source := tc.Source
				if prefix == "rslint" {
					source = tc.RslintSource
				}
				sf := parseCommentStoreSource(t, source)
				for _, variant := range tc.Variants {
					for _, reverse := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%s/noFix=%v/reverse=%v", prefix, variant.Severity, variant.DisableFixes, reverse), func(t *testing.T) {
							manager := NewDisableManager(sf, NewCommentStore(sf))
							suppression := NewSuppressionManager(sf, NewCommentStore(sf))
							invalid := map[string]bool{}
							for _, name := range tc.InvalidRules {
								invalid[name] = true
							}
							manager.ValidateRules(func(name string) bool { return !invalid[name] })
							for n := range tc.Reports {
								i := n
								if reverse {
									i = len(tc.Reports) - 1 - n
								}
								report := tc.Reports[i]
								got := manager.IsRuleDisabled(report.RuleID, directivefixtures.ByteOffset(t, source, report.Pos))
								// Unavailable rules cannot emit native reports. Every real
								// report must agree in the compact and reporting paths.
								if !invalid[report.RuleID] {
									if suppressed := suppression.IsRuleDisabled(report.RuleID, directivefixtures.ByteOffset(t, source, report.Pos)); suppressed != report.Suppressed {
										t.Errorf("suppression-only report %d (%s at %d): suppressed=%v, want %v", i, report.RuleID, report.Pos, suppressed, report.Suppressed)
									}
								}
								if got != report.Suppressed {
									t.Errorf("report %d (%s at %d): suppressed=%v, want %v", i, report.RuleID, report.Pos, got, report.Suppressed)
								}
							}
							ignored := map[string]bool{}
							for _, name := range tc.Ignored {
								ignored[name] = true
							}
							demand := EditDemandAutofix
							if variant.DisableFixes {
								demand = 0
							}
							diagnostics := manager.UnusedDiagnostics("/test.ts", ParseSeverity(variant.Severity), demand, ignored)
							sort.SliceStable(diagnostics, func(i, j int) bool { return diagnostics[i].Range.Pos() < diagnostics[j].Range.Pos() })
							if len(diagnostics) != len(variant.Diagnostics) {
								t.Fatalf("got %d unused diagnostics, want %d: %+v", len(diagnostics), len(variant.Diagnostics), diagnostics)
							}
							for i, got := range diagnostics {
								want := variant.Diagnostics[i]
								line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, got.Range.Pos())
								severity := SeverityError
								if want.Severity == 1 {
									severity = SeverityWarning
								}
								if got.RuleName != "" || got.Message.Description != want.Message || line+1 != want.Line || int(column)+1 != want.Column || got.Severity != severity {
									t.Errorf("diagnostic %d: %q at %d:%d severity %v, want %+v", i, got.Message.Description, line+1, column+1, got.Severity, want)
								}
								fixes := got.Fixes()
								if want.Fix == nil {
									if len(fixes) != 0 {
										t.Error("unexpected fix")
									}
									continue
								}
								if len(fixes) != 1 {
									t.Fatalf("got %d fixes, want one", len(fixes))
								}
								fix := fixes[0]
								wantFix := directivefixtures.LinePreservingFix(t, source, *want.Fix)
								if fix.Range.Pos() != directivefixtures.ByteOffset(t, source, wantFix.Range[0]) || fix.Range.End() != directivefixtures.ByteOffset(t, source, wantFix.Range[1]) || fix.Text != wantFix.Text {
									t.Errorf("fix: %+v, want %+v", fix, wantFix)
								}
							}
						})
					}
				}
			}
		})
	}
}

// Compare complete line-terminator tokens, rather than just the total count:
// splitting a CRLF and merging a different CR/LF must not cancel each other out.
func TestUnusedDirectiveLineTerminatorComposition(t *testing.T) {
	terminators := regexp.MustCompile("\r\n|[\r\n\u2028\u2029]")
	breaks := []string{"", "\r", "\n", "\r\n", "\n\r", "\r x\n", "\u2028", "\u2029"}
	boundaries := []string{"", "\r", "\n", "\r\n"}
	for _, prefix := range boundaries {
		for _, suffix := range boundaries {
			for _, first := range breaks {
				for _, second := range breaks {
					for _, third := range breaks {
						comments := []string{"/* eslint-disable a" + first + "*/", "/* eslint-disable b" + second + "*/", "/* eslint-disable c" + third + "*/"}
						source := prefix + strings.Join(comments, "") + suffix + "debugger;"
						want := terminators.FindAllString(source, -1)
						for mask := range 8 {
							var fixed strings.Builder
							fixed.WriteString(prefix)
							offset := len(prefix)
							for i, comment := range comments {
								end := offset + len(comment)
								if mask&(1<<i) != 0 {
									replacement := directiveLineTerminators(source, offset, end)
									if replacement == "" {
										replacement = " "
									}
									fixed.WriteString(replacement)
								} else {
									fixed.WriteString(comment)
								}
								offset = end
							}
							fixed.WriteString(suffix + "debugger;")
							if got := terminators.FindAllString(fixed.String(), -1); !slices.Equal(got, want) {
								t.Fatalf("mask=%d source=%q fixed=%q; line terminators got=%q want=%q", mask, source, fixed.String(), got, want)
							}
						}
					}
				}
			}
		}
	}
}

func TestUnusedDirectiveFixturePrecedingCR(t *testing.T) {
	const source = "/* eslint-disable \ra,b */x();"
	sf := parseCommentStoreSource(t, source)
	manager := NewDisableManager(sf, NewCommentStore(sf))
	if !manager.IsRuleDisabled("b", strings.Index(source, "x();")) {
		t.Fatal("expected b to be suppressed")
	}
	diagnostics := manager.UnusedDiagnostics("/test.ts", SeverityWarning, EditDemandAutofix, nil)
	if len(diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
	}
	fix := diagnostics[0].Fixes()[0]
	want := directivefixtures.LinePreservingFix(t, source, directivefixtures.Fix{Range: [2]int{fix.Range.Pos(), fix.Range.End()}, Text: ""})
	if fix.Text != " " || want.Text != " " {
		t.Fatalf("production fix = %q, fixture fix = %q, want space", fix.Text, want.Text)
	}
}
