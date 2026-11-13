// cspell:ignore dgimsuy dgimsvy dgimuvy giig Cpmn Ougr Tnsa Vith
package no_invalid_regexp

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidRegexpExtras(t *testing.T) {
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
			{Code: `new RegExp('(?i:a)')`},     // modifier group
			{Code: `new RegExp('(?i-m:a)')`},   // modifier group turning one on and one off
			{Code: `new RegExp('(?-ims:a)')`},  // modifier group turning three off
			{Code: `RegExp('a{1,2}')`},         // quantifier
			// Annex B lets a lookahead be quantified, and a `{` that no bound
			// closes stands for itself.
			{Code: `RegExp('(?=a)*')`},
			{Code: `RegExp('^{')`},
			{Code: `new RegExp('.', 'v')`}, // v flag alone is valid
			{Code: `new RegExp('.', 'u')`}, // u flag alone is valid
			// No arguments
			{Code: `RegExp()`},
			{Code: `new RegExp`},
			// Non-string pattern types — skip pattern validation
			{Code: "RegExp(`pattern`)"},        // template literal
			{Code: `RegExp(pattern)`},          // variable
			{Code: `RegExp('[' + '')`},         // binary expression
			{Code: `RegExp(cond ? '[' : '.')`}, // conditional
			{Code: `RegExp(123)`},              // number
			{Code: `RegExp(null)`},             // null
			{Code: `RegExp(undefined)`},        // undefined
			{Code: `RegExp(/abc/)`},            // regex literal
			// Non-RegExp callee — should not match
			{Code: `global.RegExp('.', 'z')`}, // member expression
			{Code: `window.RegExp('.', 'z')`}, // member expression
			{Code: `this.RegExp('.', 'z')`},   // member expression
			{Code: `foo.RegExp('.', 'z')`},    // member expression
			{Code: `regexp('.', 'z')`},        // case sensitivity
			// Non-literal flags — skip flag validation
			{Code: "new RegExp('.', `g`)"},       // template literal flags
			{Code: `new RegExp('.', 'g' + 'i')`}, // binary expression flags
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
			// allowConstructorFlags as array with multi-char string
			{
				Code:    `new RegExp('.', 'az')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"az"}}},
			},
			// allowConstructorFlags as array with single-char strings
			{
				Code:    `new RegExp('.', 'az')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"a", "z"}}},
			},
			// allowConstructorFlags: standard flag in list (no-op)
			{
				Code:    `new RegExp('.', 'g')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"u"}}},
			},
			// allowConstructorFlags: multiple custom flags
			{
				Code:    `new RegExp('.', 'agz')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"a", "z"}}},
			},
			// allowConstructorFlags: case sensitive
			{
				Code:    `new RegExp('.', 'A')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"A"}}},
			},
			// allowConstructorFlags: empty array (no effect)
			{
				Code:    `new RegExp('.', 'g')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{}}},
			},
			// Previously skipped valid syntax; all cases now run.
			// Unicode property long names
			{Code: `new RegExp('\\p{Letter}', 'u')`},
			// Unicode Script= syntax
			{Code: `new RegExp('\\p{Script=Nandinagari}', 'u')`},
			{Code: `new RegExp('\\p{Script=Cpmn}', 'u')`},
			{Code: `new RegExp('\\p{Script=Cypro_Minoan}', 'u')`},
			{Code: `new RegExp('\\p{Script=Old_Uyghur}', 'u')`},
			{Code: `new RegExp('\\p{Script=Ougr}', 'u')`},
			{Code: `new RegExp('\\p{Script=Tangsa}', 'u')`},
			{Code: `new RegExp('\\p{Script=Tnsa}', 'u')`},
			{Code: `new RegExp('\\p{Script=Toto}', 'u')`},
			{Code: `new RegExp('\\p{Script=Vith}', 'u')`},
			{Code: `new RegExp('\\p{Script=Vithkuqi}', 'u')`},
			// v-flag set notation
			{Code: `new RegExp('[A--B]', 'v')`},
			{Code: `new RegExp('[A--[0-9]]', 'v')`},
			{Code: `new RegExp('[\\p{Basic_Emoji}--\\q{a|bc|def}]', 'v')`},
			{Code: `new RegExp('[A--B]', flags)`},
			// Surrogate pair named capture groups
			{Code: `new RegExp('(?<\\ud835\\udc9c>.)', 'g')`},
			{Code: `new RegExp('(?<\\u{1d49c}>.)', 'g')`},
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
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Case-sensitive: "A" allowed but "a" used
			{
				Code:    `RegExp('.', 'a')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"A"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Duplicate custom flag with allowConstructorFlags
			{
				Code:    `new RegExp('.', 'aa')`,
				Options: []any{map[string]interface{}{"allowConstructorFlags": []interface{}{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Invalid escape in unicode mode
			{
				Code: `new RegExp('\\a', 'u');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// A quantifier repeating an assertion. Without `u` a `\p` is a `p`,
			// which leaves the `^` an anchor with a `?` after it.
			{
				Code: `new RegExp('\\p^?');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('^*');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('\\b{1}');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?=a)*', 'u');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Previously skipped invalid syntax; all cases now run.
			// v-flag specific parsing
			{
				Code: `new RegExp('[[]', 'v');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('[[]\\u{0}*', 'v');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Duplicate named capture groups outside alternatives
			{
				Code: `new RegExp('(?<k>a)(?<k>b)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Inline modifier validation
			{
				Code: `new RegExp('(?ii:foo)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-ii:foo)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?i-i:foo)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-:foo)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?-u:foo)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Trailing backslash
			{
				Code: `new RegExp('\\');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
		},
	)
}

// Cases retained from the previous JS suite.
func TestNoInvalidRegexpLegacyIntegrationExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidRegexpRule, []rule_tester.ValidTestCase{
		{Code: "new RegExp('\\\\p{Letter}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Nandinagari}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Cpmn}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Cypro_Minoan}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Old_Uyghur}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Ougr}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Tangsa}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Tnsa}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Toto}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Vith}', 'u');"},
		{Code: "new RegExp('\\\\p{Script=Vithkuqi}', 'u');"},
		{Code: "new RegExp('[A--B]', 'v');"},
		{Code: "new RegExp('[A--[0-9]]', 'v');"},
		{Code: "new RegExp('[\\\\p{Basic_Emoji}--\\\\q{a|bc|def}]', 'v');"},
		{Code: "new RegExp('[A--B]', flags);"},
		{Code: "new RegExp('(?<\\\\ud835\\\\udc9c>.)', 'g');"},
		{Code: "new RegExp('(?<\\\\u{1d49c}>.)', 'g');"},
	}, []rule_tester.InvalidTestCase{})
}

func TestNoInvalidRegexpRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidRegexpRule, []rule_tester.ValidTestCase{
		{Code: "new RegExp('.', 'é💩')", Options: []any{map[string]any{"allowConstructorFlags": []any{"é", "💩"}}}},
		{Code: "new (RegExp)(('a'), ('g'))"},
		{Code: "function f(RegExp) { return (RegExp)('[', 'z'); }"},
		{Code: "const RegExp = () => {}; new RegExp('[', 'z');"},
		{Code: "RegExp('[');", Globals: map[string]any{"RegExp": "off"}},
		{Code: "new RegExp('[a--b]', flags)"},
	}, []rule_tester.InvalidTestCase{
		{Code: "new RegExp('.', '💩💩')", Options: []any{map[string]any{"allowConstructorFlags": []any{"💩"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Duplicate flags ('💩') supplied to RegExp constructor.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}},
		{Code: "new RegExp('.', '💀')", Options: []any{map[string]any{"allowConstructorFlags": []any{"💩"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid flags supplied to RegExp constructor '💀'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}},
		{Code: "new RegExp('.', 'ê')", Options: []any{map[string]any{"allowConstructorFlags": []any{"é"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid flags supplied to RegExp constructor 'ê'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21}}},
		{Code: "(RegExp)(('['), ('g'))", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid regular expression: /[/: Unterminated character class.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23}}},
		{Code: "RegExp?.('[')", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid regular expression: /[/: Unterminated character class.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14}}},
		{Code: "RegExp(...patterns, 'z')", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid flags supplied to RegExp constructor 'z'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25}}},
		{Code: "RegExp('[', ...flags)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid regular expression: /[/: Unterminated character class.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}},
		{Code: "const text = '💩';\n  new RegExp(\n   '[',\n   'u'\n  );", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regexMessage", Message: "Invalid regular expression: /[/u: Unterminated character class.", Line: 2, Column: 3, EndLine: 5, EndColumn: 4}}},
	})
}

// The compiler-backed Go RuleTester rejects legacy octal string escapes before
// running rules, even in a script. Exercise the unchanged input through the
// parser and rule directly, including edit demand and the full report range.
func TestNoInvalidRegexpSloppyOctal(t *testing.T) {
	const code = `new RegExp('\133')`
	sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/octal.js", Path: "/octal.js"}, code, core.ScriptKindJS)
	comments := rule.NewCommentStore(sf)
	var diags []rule.RuleDiagnostic
	ctx := rule.RuleContext{SourceFile: sf, Comments: comments, DisableManager: rule.NewDisableManager(sf, comments), LanguageOptions: rule.LanguageOptions{SourceType: "script"}}.WithDiagnosticConsumer(NoInvalidRegexpRule.Name, rule.SeverityError, rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diags = append(diags, d) }})
	call := sf.Statements.Nodes[0].AsExpressionStatement().Expression
	NoInvalidRegexpRule.Run(ctx, nil)[call.Kind](call)
	if len(diags) != 1 {
		t.Fatalf("diagnostics=%#v", diags)
	}
	d := diags[0]
	if d.Message.Id != "regexMessage" || d.Message.Description != "Invalid regular expression: /[/: Unterminated character class." || d.Range.Pos() != 0 || d.Range.End() != len(code) {
		t.Fatalf("diagnostic=%#v", d)
	}
	if d.FixesPtr != nil || d.Suggestions != nil {
		t.Fatalf("unexpected edits: %#v", d)
	}
}
