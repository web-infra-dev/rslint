// cspell:ignore dgimsuy dgimsvy dgimuvy giig Cpmn Ougr Tnsa Vith
package no_invalid_regexp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidRegexpRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidRegexpRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `RegExp('.')`},
			{Code: `new RegExp('.')`},
			{Code: `new RegExp('.', 'im')`},
			{Code: `new RegExp('.', 'gmi')`},
			{Code: `new RegExp('.', 'dgimsuy')`},
			{Code: `new RegExp(pattern, 'g')`}, // non-literal pattern, skip
			{Code: `new RegExp('.', flags)`},   // non-literal flags, skip
			{Code: `RegExp('')`},               // empty pattern is valid
			{Code: `RegExp('a|b')`},            // alternation
			{Code: `new RegExp('\\\\d+')`},     // escaped digits
			{Code: `new RegExp('[abc]')`},      // character class
			{Code: `new RegExp('(?:a)')`},      // non-capturing group
			{Code: `RegExp('a{1,2}')`},         // quantifier
			{Code: `new RegExp('.', 'v')`},     // v flag alone is valid
			{Code: `new RegExp('.', 'u')`},     // u flag alone is valid
			// No arguments
			{Code: `RegExp()`},
			{Code: `new RegExp`},
			// Non-string pattern types — skip pattern validation
			{Code: "RegExp(`pattern`)"},             // template literal
			{Code: `RegExp(pattern)`},               // variable
			{Code: `RegExp('[' + '')`},              // binary expression
			{Code: `RegExp(cond ? '[' : '.')`},      // conditional
			{Code: `RegExp(123)`},                   // number
			{Code: `RegExp(null)`},                  // null
			{Code: `RegExp(undefined)`},             // undefined
			{Code: `RegExp(/abc/)`},                 // regex literal
			// Non-RegExp callee — should not match
			{Code: `global.RegExp('.', 'z')`},       // member expression
			{Code: `window.RegExp('.', 'z')`},       // member expression
			{Code: `this.RegExp('.', 'z')`},         // member expression
			{Code: `foo.RegExp('.', 'z')`},          // member expression
			{Code: `regexp('.', 'z')`},              // case sensitivity
			// Non-literal flags — skip flag validation
			{Code: "new RegExp('.', `g`)"},           // template literal flags
			{Code: `new RegExp('.', 'g' + 'i')`},    // binary expression flags
			// Non-literal pattern + non-literal flags — skip both
			{Code: `new RegExp(pattern, flags)`},
			// Non-literal pattern + valid flags
			{Code: `new RegExp(foo, 'gi')`},
			{Code: `new RegExp(pattern, '')`},
			// Empty flags string
			{Code: `new RegExp('.', '')`},
			// All valid flags without u (with v)
			{Code: `new RegExp('.', 'dgimsvy')`},
			// Each individual valid flag
			{Code: `new RegExp('.', 'd')`},
			{Code: `new RegExp('.', 'g')`},
			{Code: `new RegExp('.', 'i')`},
			{Code: `new RegExp('.', 'm')`},
			{Code: `new RegExp('.', 's')`},
			{Code: `new RegExp('.', 'y')`},
			// allowConstructorFlags option
			{
				Code:    `new RegExp('.', 'z')`,
				Options: map[string]interface{}{"allowConstructorFlags": "z"},
			},
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": "az"},
			},
			// allowConstructorFlags as array with multi-char string
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"az"}},
			},
			// allowConstructorFlags as array with single-char strings
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a", "z"}},
			},
			// allowConstructorFlags: standard flag in list (no-op)
			{
				Code:    `new RegExp('.', 'g')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"u"}},
			},
			// allowConstructorFlags: multiple custom flags
			{
				Code:    `new RegExp('.', 'agz')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a", "z"}},
			},
			// allowConstructorFlags: case sensitive
			{
				Code:    `new RegExp('.', 'A')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"A"}},
			},
			// allowConstructorFlags: empty array (no effect)
			{
				Code:    `new RegExp('.', 'g')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{}},
			},
			// === Skipped: regexp2 engine limitations (FP — valid in ESLint but regexp2 rejects) ===
			// Unicode property long names
			{Code: `new RegExp('\\p{Letter}', 'u')`, Skip: true},
			// Unicode Script= syntax
			{Code: `new RegExp('\\p{Script=Nandinagari}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Cpmn}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Cypro_Minoan}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Old_Uyghur}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Ougr}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Tangsa}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Tnsa}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Toto}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Vith}', 'u')`, Skip: true},
			{Code: `new RegExp('\\p{Script=Vithkuqi}', 'u')`, Skip: true},
			// v-flag set notation
			{Code: `new RegExp('[A--B]', 'v')`, Skip: true},
			{Code: `new RegExp('[A--[0-9]]', 'v')`, Skip: true},
			{Code: `new RegExp('[\\p{Basic_Emoji}--\\q{a|bc|def}]', 'v')`, Skip: true},
			{Code: `new RegExp('[A--B]', flags)`, Skip: true},
			// Surrogate pair named capture groups
			{Code: `new RegExp('(?<\\ud835\\udc9c>.)', 'g')`, Skip: true},
			{Code: `new RegExp('(?<\\u{1d49c}>.)', 'g')`, Skip: true},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === Flag validation: invalid flags ===
			{
				Code: `RegExp('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'x');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Case-sensitive: uppercase flags are invalid
			{
				Code: `new RegExp('.', 'G');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'I');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// All invalid flags
			{
				Code: `new RegExp('.', 'abc');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Mixed valid + invalid
			{
				Code: `new RegExp('.', 'gz');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'gia');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Flag validation: duplicate flags ===
			{
				Code: `new RegExp('.', 'aa');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'gg');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Duplicate each standard flag
			{
				Code: `new RegExp('.', 'dd');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'ii');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'mm');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'ss');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'uu');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'vv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'yy');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Multiple duplicates
			{
				Code: `new RegExp('.', 'giig');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Mixed invalid + duplicate
			{
				Code: `new RegExp('.', 'ggz');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Flag validation: u+v mutually exclusive ===
			{
				Code: `RegExp('.', 'uv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// u+v in longer string
			{
				Code: `new RegExp('.', 'dgimuvy');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// u+v with other valid flags
			{
				Code: `new RegExp('.', 'guv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Pattern validation ===
			{
				Code: `RegExp('[');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('(');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('\\');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Pattern with valid flags — still validate pattern
			{
				Code: `RegExp('[', 'g', 'extra');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Non-literal flags + invalid pattern ===
			{
				Code: `new RegExp('[', flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Non-literal pattern + flag errors ===
			{
				Code: `RegExp(foo, 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp(foo, 'gg');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp(foo, 'uv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Template literal pattern + invalid flags
			{
				Code: "RegExp(`test`, 'z');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Binary expression pattern + invalid flags
			{
				Code: `RegExp('a' + 'b', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Parenthesized callee ===
			{
				Code: `(RegExp)('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Nesting / composition ===
			// Inside function call
			{
				Code: `foo(new RegExp('['));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 5},
				},
			},
			// Inside conditional
			{
				Code: `cond ? new RegExp('[') : null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 8},
				},
			},
			// Inside array
			{
				Code: `[RegExp('[')];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 2},
				},
			},
			// Inside object
			{
				Code: `({ re: RegExp('[') });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 8},
				},
			},
			// Inside arrow function
			{
				Code: `const fn = () => RegExp('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 18},
				},
			},
			// Nested RegExp: inner is invalid
			{
				Code: `RegExp(RegExp('['));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 8},
				},
			},
			// In template literal expression
			{
				Code: "`${RegExp('[')}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 4},
				},
			},
			// IIFE
			{
				Code: `(function() { RegExp('.', 'z'); })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 15},
				},
			},
			// Class method
			{
				Code: `class C { m() { new RegExp('['); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 17},
				},
			},
			// Default parameter
			{
				Code: `function f(x = new RegExp('[')) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 16},
				},
			},
			// For loop init
			{
				Code: `for (let r = new RegExp('[');;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 14},
				},
			},
			// Assignment
			{
				Code: `let r = new RegExp('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 9},
				},
			},
			// Return statement
			{
				Code: `function g() { return new RegExp('['); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 23},
				},
			},
			// Logical expression
			{
				Code: `true && new RegExp('.', 'zz');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 9},
				},
			},
			// === allowConstructorFlags: still-invalid cases ===
			{
				Code:    `new RegExp('.', 'z')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Case-sensitive: "A" allowed but "a" used
			{
				Code:    `RegExp('.', 'a')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"A"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Duplicate custom flag with allowConstructorFlags
			{
				Code:    `new RegExp('.', 'aa')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// === Skipped: regexp2 engine limitations (FN — invalid in ESLint but regexp2 misses) ===
			// Invalid escape in unicode mode
			{
				Code: `new RegExp('\\a', 'u');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// v-flag specific parsing
			{
				Code: `new RegExp('[[]', 'v');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('[[]\\u{0}*', 'v');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Duplicate named capture groups outside alternatives
			{
				Code: `new RegExp('(?<k>a)(?<k>b)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Inline modifier validation
			{
				Code: `new RegExp('(?ii:foo)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-ii:foo)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?i-i:foo)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-:foo)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-u:foo)');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Trailing backslash
			{
				Code: `new RegExp('\\');`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
		},
	)
}
