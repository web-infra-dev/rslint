package no_nested_ternary_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_nested_ternary"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	tooDeepMessage       = "Do not nest ternary expressions."
	tooDeepMessageID     = "no-nested-ternary/too-deep"
	shouldParenMessage   = "Nested ternary expression should be parenthesized."
	shouldParenMessageID = "no-nested-ternary/should-parenthesized"
)

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code}
}

func lines(parts ...string) string {
	return strings.Join(parts, "\n")
}

// TestNoNestedTernaryUpstream migrates the full valid/invalid suite from
// upstream test/no-nested-ternary.js (v73.0.0) 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_nested_ternary_extras_test.go file.
func TestNoNestedTernaryUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_nested_ternary.NoNestedTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- test valid ----
			jsValid(`const foo = i > 5 ? true : false;`),
			jsValid(`const foo = i > 5 ? true : (i < 100 ? true : false);`),
			jsValid(`const foo = i > 5 ? (i < 100 ? true : false) : true;`),
			jsValid(`const foo = i > 5 ? (i < 100 ? true : false) : (i < 100 ? true : false);`),
			jsValid(`const foo = i > 5 ? true : (i < 100 ? FOO(i > 50 ? false : true) : false);`),
			// Parenthesized ternary in the test position
			jsValid(`const foo = (a ? b : c) ? d : e;`),
			jsValid(`const foo = (a ? b : c) ? (d ? e : f) : g;`),
			jsValid(`foo ? doBar() : doBaz();`),
			jsValid(`var foo = bar === baz ? qux : quxx;`),

			// ---- test.typescript valid ----
			// #663 — multi-line ternary with paren-wrapped alternate
			tsValid(lines(
				`const pluginName = isAbsolute ?`,
				`	pluginPath.slice(pluginPath.lastIndexOf('/') + 1) :`,
				`	(`,
				`		isNamespaced ?`,
				`		pluginPath.split('@')[1].split('/')[1] :`,
				`		pluginPath`,
				`	);`,
			)),
		},
		[]rule_tester.InvalidTestCase{
			// ---- test invalid ----
			// 3 levels deep, paren-wrapped everywhere: innermost reports too-deep.
			{
				Code:   `const foo = i > 5 ? true : (i < 100 ? true : (i < 1000 ? true : false));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: tooDeepMessageID, Message: tooDeepMessage, Line: 1, Column: 47, EndLine: 1, EndColumn: 70}},
			},
			// 2 levels, inner-inner in the consequent position: innermost reports too-deep.
			{
				Code:   `const foo = i > 5 ? true : (i < 100 ? (i > 50 ? false : true) : false);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: tooDeepMessageID, Message: tooDeepMessage, Line: 1, Column: 40, EndLine: 1, EndColumn: 61}},
			},
			// 1 level unparenthesized → should-parenthesized on the inner.
			{
				Code:   `const foo = i > 5 ? i < 100 ? true : false : true;`,
				Output: []string{`const foo = i > 5 ? (i < 100 ? true : false) : true;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 43}},
			},
			// Two unparenthesized inners under one outer → 2 should-parenthesized errors.
			{
				Code:   `const foo = i > 5 ? i < 100 ? true : false : i < 100 ? true : false;`,
				Output: []string{`const foo = i > 5 ? (i < 100 ? true : false) : (i < 100 ? true : false);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 43},
					{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 46, EndLine: 1, EndColumn: 68},
				},
			},
			// 1 level unparenthesized on the alternate side.
			{
				Code:   `const foo = i > 5 ? true : i < 100 ? true : false;`,
				Output: []string{`const foo = i > 5 ? true : (i < 100 ? true : false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 28, EndLine: 1, EndColumn: 50}},
			},
			// Top-level statement, alternate side unparenthesized.
			{
				Code:   `foo ? bar : baz === qux ? quxx : foobar;`,
				Output: []string{`foo ? bar : (baz === qux ? quxx : foobar);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 13, EndLine: 1, EndColumn: 40}},
			},
			// Top-level statement, consequent side unparenthesized.
			{
				Code:   `foo ? baz === qux ? quxx : foobar : bar;`,
				Output: []string{`foo ? (baz === qux ? quxx : foobar) : bar;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID, Message: shouldParenMessage, Line: 1, Column: 7, EndLine: 1, EndColumn: 34}},
			},
		},
	)
}
