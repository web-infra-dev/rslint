package no_empty_character_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHasEmptyCharacterClassLegacy(t *testing.T) {
	// Each test: pattern (the part between / and /flags), expected result.
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		// ── Basic non-class patterns ──
		{"no class at all", `abc`, false},
		{"empty pattern", ``, false},

		// ── Non-empty character classes ──
		{"single char class", `[a]`, false},
		{"range class", `[a-zA-Z]`, false},
		{"multi-char class", `[abc]`, false},

		// ── Empty character class ──
		{"basic empty class", `[]`, true},
		{"empty class at start", `[]abc`, true},
		{"empty class at end", `abc[]`, true},
		{"empty class in middle", `foo[]bar`, true},
		{"empty class with trailing ]", `[]]`, true},

		// ── Negated empty class [^] — allowed ──
		{"negated empty class", `[^]`, false},
		{"negated non-empty class", `[^a]`, false},
		{"negated range", `[^a-z]`, false},
		{"double caret", `[^^]`, false},

		// ── Escaped characters outside class ──
		{"escaped [ outside", `\[]`, false},           // \[ is literal [, ] is literal ]
		{"escaped \\ then [", `\\[]`, true},           // \\ is literal \, then [] is empty class
		{"escaped [ then empty class", `\[[]`, true},    // \[ literal [, then [] is empty class
		{"escaped [ escaped ]", `\[\]`, false},        // \[ and \] are both literals

		// ── Escaped characters inside class ──
		{"escaped ] inside class", `[\]]`, false},     // class contains literal ]
		{"escaped [ inside class", `[\[]`, false},     // class contains literal [
		{"escaped \\ inside class", `[\\]`, false},    // class contains literal \.
		{"escaped a inside class", `[\a]`, false},     // class contains literal a
		{"backslash at class start then ]", `[\]]`, false},

		// ── [ as literal inside class (legacy: no nesting) ──
		{"[ inside class", `[[]`, false},              // class contains [
		{"[ then a inside class", `[[a]`, false},      // class contains [ and a
		{"[ ] outside after class", `[a]]`, false},    // class [a], then literal ]

		// ── Multiple classes ──
		{"two non-empty classes", `[a][b]`, false},
		{"first empty second ok", `[][b]`, true},
		{"first ok second empty", `[a][]`, true},
		{"both empty", `[][]`, true},

		// ── Complex escape sequences ──
		{"complex escapes in class", `[\-\[\]\/\{\}\(\)\*\+\?\.\\^\$\|]`, false},
		{"whitespace class", `\s*:\s*`, false},

		// ── Classes with flags (flags stripped before calling) ──
		{"escaped ] with uy", `[\]]`, false},
		{"escaped ] with s", `[\]]`, false},
		{"escaped ] with d", `[\]]`, false},

		// ── Malformed patterns (should not panic) ──
		{"unterminated class", `[abc`, false},
		{"unterminated escape in class", `[\`, false},
		{"unterminated escape outside", `\`, false},
		{"single [", `[`, false},
		{"single [^", `[^`, false},
		{"[ then \\", `[\\`, false},

		// ── ESLint original test patterns ──
		{"eslint: ^abc[a-zA-Z]", `^abc[a-zA-Z]`, false},
		{"eslint: ^abc[]", `^abc[]`, true},
		{"eslint: foo[]bar", `foo[]bar`, true},
		{"eslint: \\[[]", `\[[]`, true},
		{"eslint: \\[\\[\\]a-z[]", `\[\[\]a-z[]`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEmptyCharacterClassLegacy(tt.pattern)
			if got != tt.want {
				t.Errorf("hasEmptyCharacterClassLegacy(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestHasEmptyCharacterClassV(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		// ══════════════════════════════════════════════
		// 1. Basic cases (same as legacy but under v-flag)
		// ══════════════════════════════════════════════
		{"no class", `abc`, false},
		{"empty pattern", ``, false},
		{"non-empty class", `[a]`, false},
		{"range class", `[a-z]`, false},
		{"empty class", `[]`, true},
		{"empty class at end", `abc[]`, true},
		{"empty class in middle", `foo[]bar`, true},

		// ══════════════════════════════════════════════
		// 2. Negated classes
		// ══════════════════════════════════════════════
		{"negated empty [^] allowed", `[^]`, false},
		{"negated non-empty", `[^a]`, false},
		{"negated with nested non-empty", `[^[a]]`, false},

		// ══════════════════════════════════════════════
		// 3. Escaped characters
		// ══════════════════════════════════════════════
		{"escaped [ outside", `\[]`, false},
		{"escaped \\\\ then []", `\\[]`, true},
		{"escaped ] inside class", `[\]]`, false},
		{"escaped [ inside class", `[\[]`, false},
		{"escaped \\\\ inside class", `[\\]`, false},

		// ══════════════════════════════════════════════
		// 4. Nested classes — depth 1
		// ══════════════════════════════════════════════
		{"nested empty [[]]", `[[]]`, true},
		{"nested non-empty [[a]]", `[[a]]`, false},
		{"nested after char [a[]]", `[a[]]`, true},
		{"nested before char [[]a]", `[[]a]`, true},
		{"nested negated empty [[^]]", `[[^]]`, false},
		{"nested negated non-empty [[^a]]", `[[^a]]`, false},
		{"negated outer with nested empty [^[]]", `[^[]]`, true},
		{"negated outer with nested negated empty [^[^]]", `[^[^]]`, false},
		{"negated outer with nested non-empty [^[a]]", `[^[a]]`, false},

		// ══════════════════════════════════════════════
		// 5. Nested classes — depth 2
		// ══════════════════════════════════════════════
		{"depth-2 empty [[[]]]", `[[[]]]`, true},
		{"depth-2 non-empty [[[a]]]", `[[[a]]]`, false},
		{"depth-2 empty innermost [a[b[]]]", `[a[b[]]]`, true},
		{"depth-2 all non-empty [a[b[c]]]", `[a[b[c]]]`, false},

		// ══════════════════════════════════════════════
		// 6. Nested classes — depth 3+
		// ══════════════════════════════════════════════
		{"depth-3 empty [[[[]]]]]", `[[[[]]]]`, true},
		{"depth-3 non-empty [[[[a]]]]", `[[[[a]]]]`, false},
		{"depth-4 empty [[[[[]]]]]]", `[[[[[]]]]]`, true},

		// ══════════════════════════════════════════════
		// 7. Multiple nested classes at same level
		// ══════════════════════════════════════════════
		{"two nested both non-empty [[a][b]]", `[[a][b]]`, false},
		{"first nested empty second ok [[]][b]]", `[[][b]]`, true},
		{"first nested ok second empty [[a][]]", `[[a][]]`, true},
		{"both nested empty [[][]]", `[[][]]`, true},
		{"three nested last empty [[a][b][]]", `[[a][b][]]`, true},
		{"three nested all ok [[a][b][c]]", `[[a][b][c]]`, false},

		// ══════════════════════════════════════════════
		// 8. Set subtraction (--) with nested classes
		// ══════════════════════════════════════════════
		{"subtraction non-empty [a--[b]]", `[a--[b]]`, false},
		{"subtraction empty RHS [a--[]]", `[a--[]]`, true},
		{"subtraction empty LHS [[]--b]", `[[]--b]`, true},
		{"subtraction both non-empty [[a]--[b]]", `[[a]--[b]]`, false},
		{"subtraction RHS nested empty [[a]--[[]]]", `[[a]--[[]]]`, true},
		{"subtraction no nested [a--b]", `[a--b]`, false},

		// ══════════════════════════════════════════════
		// 9. Set intersection (&&) with nested classes
		// ══════════════════════════════════════════════
		{"intersection non-empty [a&&[b]]", `[a&&[b]]`, false},
		{"intersection empty RHS [a&&[]]", `[a&&[]]`, true},
		{"intersection empty LHS [[]&&b]", `[[]&&b]`, true},
		{"intersection both non-empty [[a]&&[b]]", `[[a]&&[b]]`, false},
		{"intersection no nested [a&&b]", `[a&&b]`, false},

		// ══════════════════════════════════════════════
		// 10. Mixed operations with deep nesting
		// ══════════════════════════════════════════════
		{"deep in subtraction [a--[b--[c--[]]]]", `[a--[b--[c--[]]]]`, true},
		{"deep in intersection [a&&[b&&[c]]]", `[a&&[b&&[c]]]`, false},
		{"subtraction then intersection [a--[b]&&[c]]", `[a--[b]&&[c]]`, false},
		{"mixed with empty deep [a--[b&&[]]]", `[a--[b&&[]]]`, true},

		// ══════════════════════════════════════════════
		// 11. Escaped characters in nested classes
		// ══════════════════════════════════════════════
		{"nested with escaped ]", "[[\\" + "]]]", false},
		{"nested with escaped [", "[[\\" + "[]]", false},
		{"nested escape then empty", "[[\\\\][]]", true},
		{"deeply nested escape", "[a[b[\\" + "]c]]", false},

		// ══════════════════════════════════════════════
		// 12. Multiple top-level classes
		// ══════════════════════════════════════════════
		{"two top-level first ok [a][b]", `[a][b]`, false},
		{"two top-level first empty [][b]", `[][b]`, true},
		{"two top-level second empty [a][]", `[a][]`, true},
		{"top-level ok then nested empty [a][[]]", `[a][[]]`, true},
		{"top-level nested ok then empty [[a]][]", `[[a]][]`, true},

		// ══════════════════════════════════════════════
		// 13. Unicode property escapes (\p{...}, \P{...})
		// ══════════════════════════════════════════════
		{"unicode property in class [\\p{ASCII}]", `[\p{ASCII}]`, false},
		{"unicode property with nested [\\p{ASCII}[a]]", `[\p{ASCII}[a]]`, false},
		// \p skip 2, then {, A, S, C, I, I, }, [, a, ], ] — all fine

		// ══════════════════════════════════════════════
		// 14. \q{} string literals (only valid inside class with v-flag)
		// ══════════════════════════════════════════════
		{"empty \\q{} in class [\\q{}]", `[\q{}]`, false},
		// \q skip 2 → {, then }, then ] → close. Non-empty (has \q, {, }).

		// ══════════════════════════════════════════════
		// 15. Pattern with ] outside class (literal)
		// ══════════════════════════════════════════════
		{"literal ] outside", `a]b`, false},
		{"class then literal ]", `[a]]`, false},
		{"empty class then literal ]", `[]]`, true},

		// ══════════════════════════════════════════════
		// 16. Malformed patterns — must not panic
		// ══════════════════════════════════════════════
		{"unterminated outer class", `[abc`, false},
		{"unterminated nested class", `[[abc`, false},
		{"unterminated deeply nested", `[[[`, false},
		{"unterminated escape in class", `[\`, false},
		{"unterminated escape outside", `\`, false},
		{"single [", `[`, false},
		{"single [^", `[^`, false},
		{"[^ then [", `[^[`, false},
		{"escape at end of nested", `[[\`, false},

		// ══════════════════════════════════════════════
		// 17. ESLint v-flag test suite patterns
		// ══════════════════════════════════════════════
		// Valid:
		{"eslint v: [[^]]", `[[^]]`, false},
		{"eslint v: [[\\]]]", "[[\\" + "]]]", false},
		{"eslint v: [[\\[]]", "[[\\" + "[]]", false},
		{"eslint v: [a--b]", `[a--b]`, false},
		{"eslint v: [a&&b]", `[a&&b]`, false},
		{"eslint v: [[a][b]]", `[[a][b]]`, false},
		// Invalid:
		{"eslint v: []", `[]`, true},
		{"eslint v: [[]]", `[[]]`, true},
		{"eslint v: [[a][]]", `[[a][]]`, true},
		{"eslint v: [a[[b[]c]]d]", `[a[[b[]c]]d]`, true},
		{"eslint v: [a--[]]", `[a--[]]`, true},
		{"eslint v: [[]--b]", `[[]--b]`, true},
		{"eslint v: [a&&[]]", `[a&&[]]`, true},
		{"eslint v: [[]&&b]", `[[]&&b]`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEmptyCharacterClassV(tt.pattern)
			if got != tt.want {
				t.Errorf("hasEmptyCharacterClassV(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestNoEmptyCharacterClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyCharacterClassRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `var foo = /^abc[a-zA-Z]/;`},
			{Code: `var regExp = new RegExp("^abc[]");`},
			{Code: `var foo = /^abc/;`},
			{Code: `var foo = /[\[]/;`},
			{Code: `var foo = /[\]]/;`},
			{Code: `var foo = /[^]/;`},
			{Code: `var foo = /\[]/`},
			{Code: `var foo = /[[]/;`},
			// v-flag valid cases (ES2024 unicodeSets)
			{Code: `var foo = /[[^]]/v;`},
			{Code: `var foo = /[[\]]]/v;`},
			{Code: `var foo = /[[\[]]/v;`},
			{Code: `var foo = /[a--b]/v;`},
			{Code: `var foo = /[a&&b]/v;`},
			{Code: `var foo = /[[a][b]]/v;`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = /^abc[]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /foo[]bar/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[]]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /\[[]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /\[\[\]a-z[]/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			// v-flag invalid cases (ES2024 unicodeSets)
			{
				Code: `var foo = /[]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[[]]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[[a][]]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[a[[b[]c]]d]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[a--[]]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[[]--b]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[a&&[]]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
			{
				Code: `var foo = /[[]&&b]/v;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11},
				},
			},
		},
	)
}
