package no_script_url

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoScriptUrlRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoScriptUrlRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// ================================================================
			// Basic non-matching strings
			// ================================================================
			{Code: `var a = 'Hello World!';`},
			{Code: `var a = 10;`},
			{Code: `var a = '';`},

			// ================================================================
			// Near misses — should NOT trigger
			// ================================================================
			// Prefix before "javascript:"
			{Code: `var url = 'xjavascript:'`},
			{Code: "var url = `xjavascript:`"},
			// No colon
			{Code: `var a = 'javascript';`},
			// Leading whitespace
			{Code: `var a = ' javascript:';`},
			// Other protocols
			{Code: `var a = 'about:blank';`},
			{Code: `var a = 'mailto:user@example.com';`},
			{Code: "var a = `https://example.com`;"},

			// ================================================================
			// Template literals with substitutions (TemplateExpression, not
			// NoSubstitutionTemplateLiteral) — static value unknown
			// ================================================================
			{Code: "var url = `${foo}javascript:`"},
			{Code: "var url = `javascript:${foo}`"},
			{Code: "var url = `${a}javascript:${b}`"},

			// ================================================================
			// Tagged templates — tag controls interpretation
			// ================================================================
			{Code: "var a = foo`javaScript:`;"},
			{Code: "var a = obj.tag`javascript:`;"},
			{Code: "var a = obj['tag']`javascript:`;"},
			{Code: "var a = tag()`javascript:`;"},
			// Nested tagged template — inner is also tagged, should NOT trigger
			{Code: "tag`${foo`javascript:`}`;"},

			// ================================================================
			// String concatenation — individual parts don't start with "javascript:"
			// ================================================================
			{Code: `var a = 'java' + 'script:';`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Unicode / hex escape sequences — AST .Text is the interpreted
			// value, so \u006a = 'j' and the string becomes "javascript:"
			// ================================================================
			{
				Code: `var a = '\u006aavascript:';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = '\x6aavascript:';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = `\\u006aavascript:`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},

			// ================================================================
			// Basic string literals
			// ================================================================
			{
				Code: `var a = 'javascript:void(0);';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = 'javascript:';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = "JavaScript:void(0)";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},

			// ================================================================
			// Case-insensitivity
			// ================================================================
			{
				Code: `var a = 'JAVASCRIPT:';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = 'jAvAsCrIpT:void(0)';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},

			// ================================================================
			// Template literals (NoSubstitutionTemplateLiteral)
			// ================================================================
			{
				Code: "var a = `javascript:`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = `JavaScript:`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},

			// ================================================================
			// Nesting contexts — assignment, property, argument, etc.
			// ================================================================
			{
				Code: "location.href = 'javascript:void(0)';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 17},
				},
			},
			{
				Code: "location.href = `javascript:void(0)`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 17},
				},
			},
			// Object property value
			{
				Code: "var obj = { href: 'javascript:void(0)' };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 19},
				},
			},
			// Array element
			{
				Code: "var arr = ['javascript:'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 12},
				},
			},
			// Function argument
			{
				Code: "fn('javascript:');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 4},
				},
			},
			// Return statement
			{
				Code: "function f() { return 'javascript:'; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 23},
				},
			},
			// Conditional expression
			{
				Code: "var a = x ? 'javascript:' : 'safe';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 13},
				},
			},
			// Default parameter
			{
				Code: "function f(url = 'javascript:') {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 18},
				},
			},
			// Class property
			{
				Code: "class A { url = 'javascript:' }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 17},
				},
			},
			// Enum initializer (TypeScript)
			{
				Code: "enum E { A = 'javascript:' }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 14},
				},
			},

			// ================================================================
			// Multiple errors in one statement
			// ================================================================
			{
				Code: "a('javascript:', 'javascript:void(0)');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 3},
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 18},
				},
			},

			// ================================================================
			// String literal inside tagged template expression — the string
			// literal itself is NOT tagged, only the outer template is
			// ================================================================
			{
				Code: "tag`${'javascript:'}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 7},
				},
			},

			// ================================================================
			// Template literal inside tagged template — parent is TemplateSpan,
			// NOT TaggedTemplateExpression, so it should trigger
			// ================================================================
			{
				Code: "tag`${`javascript:`}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 7},
				},
			},

			// ================================================================
			// JSX — the most common real-world trigger
			// ================================================================
			{
				Code: `var a = <a href="javascript:void(0)" />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 17},
				},
			},
			{
				Code: `var a = <a href={'javascript:void(0)'} />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 18},
				},
			},

			// ================================================================
			// Arrow function implicit return
			// ================================================================
			{
				Code: "const f = () => 'javascript:';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 17},
				},
			},

			// ================================================================
			// Logical / nullish operators
			// ================================================================
			{
				Code: "var a = x ?? 'javascript:';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 14},
				},
			},
			{
				Code: "var a = x || 'javascript:';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 14},
				},
			},

			// ================================================================
			// Computed property / element access
			// ================================================================
			{
				Code: "var a = obj['javascript:'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 13},
				},
			},
			{
				Code: "var obj = { ['javascript:']: 1 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 14},
				},
			},

			// ================================================================
			// TypeScript type position — string literal types are also checked
			// ================================================================
			{
				Code: "type T = 'javascript:';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 10},
				},
			},

			// ================================================================
			// Binary expression — matching side still triggers
			// ================================================================
			{
				Code: "var a = 'javascript:' + path;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 1, Column: 9},
				},
			},

			// ================================================================
			// Multi-line
			// ================================================================
			{
				Code: "var a =\n  'javascript:';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedScriptURL", Line: 2, Column: 3},
				},
			},
		},
	)
}
