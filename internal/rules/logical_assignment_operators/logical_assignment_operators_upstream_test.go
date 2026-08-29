package logical_assignment_operators

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestLogicalAssignmentOperatorsUpstream migrates the full valid/invalid suite
// from upstream tests/lib/rules/logical-assignment-operators.js (eslint
// v10.9.1) 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// logical_assignment_operators_extras_test.go.
func TestLogicalAssignmentOperatorsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&LogicalAssignmentOperatorsRule,
		[]rule_tester.ValidTestCase{
			// ---- Unrelated ----
			{Code: `a || b`},
			{Code: `a && b`},
			{Code: `a ?? b`},
			{Code: `a || a || b`},
			{Code: `var a = a || b`},
			{Code: `a === undefined ? a : b`},
			{Code: `while (a) a = b`},
			// ---- Preferred ----
			{Code: `a ||= b`},
			{Code: `a &&= b`},
			{Code: `a ??= b`},
			// ---- > Operator ----
			{Code: `a += a || b`},
			{Code: `a *= a || b`},
			{Code: `a ||= a || b`},
			{Code: `a &&= a || b`},
			// ---- > Right ----
			{Code: `a = a`},
			{Code: `a = b`},
			{Code: `a = a === b`},
			{Code: `a = a + b`},
			{Code: `a = a / b`},
			{Code: `a = fn(a) || b`},
			// ---- > Reference ----
			{Code: `a = false || c`},
			{Code: `a = f() || g()`},
			{Code: `a = b || c`},
			{Code: `a = b || a`},
			{Code: `object.a = object.b || c`},
			{Code: `[a] = a || b`},
			{Code: `({ a } = a || b)`},
			// ---- Logical ----
			{Code: `(a = b) || a`},
			{Code: `a + (a = b)`},
			{Code: `a || (b ||= c)`},
			{Code: `a || (b &&= c)`},
			{Code: `a || b === 0`},
			{Code: `a || fn()`},
			{Code: `a || (b && c)`},
			{Code: `a || (b ?? c)`},
			// ---- > Reference ----
			{Code: `a || (b = c)`},
			{Code: `a || (a ||= b)`},
			{Code: `fn() || (a = b)`},
			{Code: `a.b || (a = b)`},
			{Code: `a?.b || (a.b = b)`},
			{Code: `class Class { #prop; constructor() { this.#prop || (this.prop = value) } }`},
			{Code: `class Class { #prop; constructor() { this.prop || (this.#prop = value) } }`},
			// ---- If ----
			{Code: `if (a) a = b`},
			{Code: `if (a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: false}}},
			{Code: `if (a) { a = b } else {}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) { a = b } else if (a) {}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (unrelated) {} else if (a) a = b; else {}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (unrelated) {} else if (a) a = b; else if (unrelated) {}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- > Body ----
			{Code: `if (a) {}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) { before; a = b }`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) { a = b; after }`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) throw new Error()`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) a`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) a ||= b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) b = a`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) { a() }`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a) { a += a || b }`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- > Test ----
			{Code: `if (true) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (predicate(a)) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a?.b) a.b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (!a?.b) a.b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === b) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a != null) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null && a === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === 0 || a === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === 1) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a == null || a == undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === !0) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === +0) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === null) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === undefined || a === void 0) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === void void 0) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === void 'string') a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === void fn()) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- > Test > Yoda ----
			{Code: `if (a == a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a == b) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null == null) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (undefined == undefined) undefined = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null == x) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null == fn()) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null === a || a === 0) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (0 === a || null === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (1 === a || a === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (undefined === a || 1 === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === b) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (b === undefined || a === null) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null === a || b === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null === null || undefined === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null === null || a === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (undefined === undefined || a === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (null === undefined || a === a) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- > Test > Undefined ----
			{Code: `{
   const undefined = 0;
   if (a == undefined) a = b
}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `(() => {
   const undefined = 0;
   if (condition) {
       if (a == undefined) a = b
   }
})()`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `{
   if (a == undefined) a = b
}
var undefined = 0;`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `{
   const undefined = 0;
   if (undefined == null) undefined = b
}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `{
   const undefined = 0;
   if (a === undefined || a === null) a = b
}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `{
   const undefined = 0;
   if (undefined === a || null === a) a = b
}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- > Reference ----
			{Code: `if (a) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (!a) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (!!a) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a == null) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || a === undefined) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || b === undefined) a = b`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (a === null || b === undefined) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `if (Boolean(a)) b = c`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			{Code: `function fn(Boolean) {
   if (Boolean(a)) a = b
}`, Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}},
			// ---- Never ----
			{Code: `a = a || b`, Options: []any{`never`}},
			{Code: `a = a && b`, Options: []any{`never`}},
			{Code: `a = a ?? b`, Options: []any{`never`}},
			{Code: `a = b`, Options: []any{`never`}},
			{Code: `a += b`, Options: []any{`never`}},
			{Code: `a -= b`, Options: []any{`never`}},
			{Code: `a.b = a.b || c`, Options: []any{`never`}},
			// ---- 3 or more operands ----
			{Code: `a = a && b || c`, Options: []any{`always`}},
			{Code: `a = a && b && c || d`, Options: []any{`always`}},
			{Code: `a = (a || b) || c`, Options: []any{`always`}},
			{Code: `a = (a && b) && c`, Options: []any{`always`}},
			{Code: `a = (a ?? b) ?? c`, Options: []any{`always`}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Assignment ----
			{
				Code:   `a = a || b`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `a = a && b`,
				Output: []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (&&=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `a = a ?? b`,
				Output: []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `foo = foo || bar`,
				Output: []string{`foo ||= bar`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- > Right ----
			{
				Code:   `a = a || fn()`,
				Output: []string{`a ||= fn()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:   `a = a || b && c`,
				Output: []string{`a ||= b && c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `a = a || (b || c)`,
				Output: []string{`a ||= (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   `a = a || (b ? c : d)`,
				Output: []string{`a ||= (b ? c : d)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- > Comments ----
			{
				Code:   `/* before */ a = a || b`,
				Output: []string{`/* before */ a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 14, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:   `a = a || b // after`,
				Output: []string{`a ||= b // after`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `a /* between */ = a || b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `a = /** @type */ a || b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `a = a || /* between */ b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			// ---- > Parenthesis ----
			{
				Code:   `(a) = a || b`,
				Output: []string{`(a) ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a = (a) || b`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a = a || (b)`,
				Output: []string{`a ||= (b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a = a || ((b))`,
				Output: []string{`a ||= ((b))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `(a = a || b)`,
				Output: []string{`(a ||= b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:   `a = a || (f(), b)`,
				Output: []string{`a ||= (f(), b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- > Suggestions ----
			{
				Code: `a.b = a.b ?? c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a.b ??= c`},
						},
					},
				},
			},
			{
				Code: `a.b.c = a.b.c ?? d`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a.b.c ??= d`},
						},
					},
				},
			},
			{
				Code: `a[b] = a[b] ?? c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a[b] ??= c`},
						},
					},
				},
			},
			{
				Code: `a['b'] = a['b'] ?? c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a['b'] ??= c`},
						},
					},
				},
			},
			{
				Code: `a.b = a['b'] ?? c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a.b ??= c`},
						},
					},
				},
			},
			{
				Code: `a['b'] = a.b ?? c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a['b'] ??= c`},
						},
					},
				},
			},
			{
				Code: `this.prop = this.prop ?? {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `this.prop ??= {}`},
						},
					},
				},
			},
			// ---- > With ----
			{
				Code: `with (object) a = a || b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `with (object) a ||= b`},
						},
					},
				},
			},
			{
				Code: `with (object) { a = a || b }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 17, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `with (object) { a ||= b }`},
						},
					},
				},
			},
			{
				Code: `with (object) { if (condition) a = a || b }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 32, EndLine: 1, EndColumn: 42,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `with (object) { if (condition) a ||= b }`},
						},
					},
				},
			},
			{
				Code:   `with (a = a || b) {}`,
				Output: []string{`with (a ||= b) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 7, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `with (object) {} a = a || b`,
				Output: []string{`with (object) {} a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:   `a = a || b; with (object) {}`,
				Output: []string{`a ||= b; with (object) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:   `if (condition) a = a || b`,
				Output: []string{`if (condition) a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 16, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `with (object) {
  "use strict";
   a = a || b
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 3, Column: 4, EndLine: 3, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `with (object) {
  "use strict";
   a ||= b
}`},
						},
					},
				},
			},
			// ---- > Context ----
			{
				Code:   `fn(a = a || b)`,
				Output: []string{`fn(a ||= b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 4, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:   `fn((a = a || b))`,
				Output: []string{`fn((a ||= b))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `(a = a || b) ? c : d`,
				Output: []string{`(a ||= b) ? c : d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 2, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:   `a = b = b || c`,
				Output: []string{`a = b ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Logical ----
			{
				Code:   `a || (a = b)`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a && (a = b)`,
				Output: []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (&&=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a ?? (a = b)`,
				Output: []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `foo ?? (foo = bar)`,
				Output: []string{`foo ??= bar`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// ---- > Right ----
			{
				Code:   `a || (a = 0)`,
				Output: []string{`a ||= 0`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `a || (a = fn())`,
				Output: []string{`a ||= fn()`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `a || (a = (b || c))`,
				Output: []string{`a ||= (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- > Parenthesis ----
			{
				Code:   `(a) || (a = b)`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `a || ((a) = b)`,
				Output: []string{`(a) ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `a || (a = (b))`,
				Output: []string{`a ||= (b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `a || ((a = b))`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `a || (((a = b)))`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `a || ( ( a = b ) )`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			// ---- > Comments ----
			{
				Code:   `/* before */ a || (a = b)`,
				Output: []string{`/* before */ a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 14, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:   `a || (a = b) // after`,
				Output: []string{`a ||= b // after`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code: `a /* between */ || (a = b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `a || /* between */ (a = b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- > Fix Condition ----
			{
				Code:   `a.b || (a.b = c)`,
				Output: []string{`a.b ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `class Class { #prop; constructor() { this.#prop || (this.#prop = value) } }`,
				Output: []string{`class Class { #prop; constructor() { this.#prop ||= value } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 38, EndLine: 1, EndColumn: 72},
				},
			},
			{
				Code:   `a['b'] || (a['b'] = c)`,
				Output: []string{`a['b'] ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:   `a[0] || (a[0] = b)`,
				Output: []string{`a[0] ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:   `a[this] || (a[this] = b)`,
				Output: []string{`a[this] ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:   `foo.bar || (foo.bar = baz)`,
				Output: []string{`foo.bar ||= baz`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `a.b.c || (a.b.c = d)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `a.b.c ||= d`},
						},
					},
				},
			},
			{
				Code: `a[b.c] || (a[b.c] = d)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `a[b.c] ||= d`},
						},
					},
				},
			},
			{
				Code: `a[b?.c] || (a[b?.c] = d)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `a[b?.c] ||= d`},
						},
					},
				},
			},
			{
				Code: `with (object) a.b || (a.b = c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 15, EndLine: 1, EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `with (object) a.b ||= c`},
						},
					},
				},
			},
			// ---- > Context ----
			{
				Code:   `a = a.b || (a.b = {})`,
				Output: []string{`a = a.b ||= {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 5, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:   `a || (a = 0) || b`,
				Output: []string{`(a ||= 0) || b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `(a || (a = 0)) || b`,
				Output: []string{`(a ||= 0) || b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 2, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:   `a || (b || (b = 0))`,
				Output: []string{`a || (b ||= 0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 7, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:   `a = b || (b = c)`,
				Output: []string{`a = b ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 5, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `a || (a = 0) ? b : c`,
				Output: []string{`(a ||= 0) ? b : c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:   `fn(a || (a = 0))`,
				Output: []string{`fn(a ||= 0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 4, EndLine: 1, EndColumn: 16},
				},
			},
			// ---- If ----
			{
				Code:    `if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `if (Boolean(a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `if (!!a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `if (!a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ||=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `if (!Boolean(a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ||=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `if (a == undefined) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `if (a == null) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `if (a === null || a === undefined) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (a === undefined || a === null) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (a === null || a === void 0) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    `if (a === void 0 || a === null) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    `if (a) { a = b; }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `{ const undefined = 0; }
if (a == undefined) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`{ const undefined = 0; }
a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code: `if (a == undefined) a = b
{ const undefined = 0; }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`a ??= b
{ const undefined = 0; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- > Yoda ----
			{
				Code:    `if (null == a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `if (undefined == a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `if (undefined === a || a === null) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (a === undefined || null === a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (undefined === a || null === a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (null === a || a === undefined) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (a === null || undefined === a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    `if (null === a || undefined === a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			// ---- > Parenthesis ----
			{
				Code:    `if ((a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`(a) &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `if (a) a = (b)`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= (b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `if (a) (a = b)`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`(a &&= b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- > Previous statement ----
			{
				Code:    `;if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`;(a) &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 2, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `{ if (a) (a) = b }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`{ (a) &&= b }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 3, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `fn();if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`fn();(a) &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 6, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: `fn()
if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`fn()
a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code: `id
if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
				},
			},
			{
				Code: `object.prop
if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
				},
			},
			{
				Code: `object[computed]
if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
				},
			},
			{
				Code: `fn()
if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
				},
			},
			// ---- > Adding semicolon ----
			{
				Code:    `if (a) a = b; fn();`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b; fn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `if (a) { a = b }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `if (a) { a = b; }
fn();`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`a &&= b;
fn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `if (a) { a = b }
fn();`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`a &&= b;
fn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `if (a) { a = b } fn();`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b; fn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code: `if (a) { a = b
} fn();`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b; fn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// ---- > Spacing ----
			{
				Code:    `if (a) a  =  b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a  &&=  b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `if (a)
 a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 2, EndColumn: 7},
				},
			},
			{
				Code: `if (a) {
 a = b; 
}`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			// ---- > Comments ----
			{
				Code:    `/* before */ if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`/* before */ a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 14, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `if (a) a = b /* after */`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b /* after */`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `if (a) /* between */ a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `if (a) a = /* between */ b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- > Members > Single Property Access ----
			{
				Code:    `if (a.b) a.b = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a.b &&= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `if (a[b]) a[b] = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a[b] &&= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `if (a['b']) a['b'] = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a['b'] &&= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `if (this.prop) this.prop = value`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`this.prop &&= value`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    `(class extends SuperClass { method() { if (super.prop) super.prop = value } })`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`(class extends SuperClass { method() { super.prop &&= value } })`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 40, EndLine: 1, EndColumn: 74},
				},
			},
			{
				Code:    `with (object) if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`with (object) a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 15, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- > Members > Possible Multiple Property Accesses ----
			{
				Code:    `if (a.b === undefined || a.b === null) a.b = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 47,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertIf`, Output: `a.b ??= c`},
						},
					},
				},
			},
			{
				Code:    `if (a.b.c) a.b.c = d`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertIf`, Output: `a.b.c &&= d`},
						},
					},
				},
			},
			{
				Code:    `if (a.b.c.d) a.b.c.d = e`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertIf`, Output: `a.b.c.d &&= e`},
						},
					},
				},
			},
			{
				Code:    `if (a[b].c) a[b].c = d`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertIf`, Output: `a[b].c &&= d`},
						},
					},
				},
			},
			{
				Code:    `with (object) if (a.b) a.b = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 15, EndLine: 1, EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertIf`, Output: `with (object) a.b &&= c`},
						},
					},
				},
			},
			// ---- > Else if ----
			{
				Code:    `if (unrelated) {} else if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) {} else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 24, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `if (a) {} else if (b) {} else if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (a) {} else if (b) {} else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 31, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code: `if (unrelated) {} else
if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`if (unrelated) {} else
a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},
			{
				Code: `if (unrelated) {
}
else if (a) {
a = b;
}`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`if (unrelated) {
}
else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 3, Column: 6, EndLine: 5, EndColumn: 2},
				},
			},
			{
				Code:    `if (unrelated) statement; else if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) statement; else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code: `if (unrelated) id
else if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 6, EndLine: 2, EndColumn: 20},
				},
			},
			{
				Code:    `if (unrelated) {} else if (a) a = b; else if (c) c = d`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) {} else if (a) a = b; else c &&= d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 43, EndLine: 1, EndColumn: 55},
				},
			},
			// ---- > Else if > Comments ----
			{
				Code:    `if (unrelated) { /* body */ } else if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) { /* body */ } else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:    `if (unrelated) {} /* before else */ else if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) {} /* before else */ else a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 42, EndLine: 1, EndColumn: 55},
				},
			},
			{
				Code: `if (unrelated) {} else // Line
if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`if (unrelated) {} else // Line
a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				},
			},
			{
				Code:    `if (unrelated) {} else /* Block */ if (a) a = b;`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (unrelated) {} else /* Block */ a &&= b;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
				},
			},
			// ---- > Patterns ----
			{
				Code:    `if (array) array = array.filter(predicate)`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`array &&= array.filter(predicate)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},
			// ---- Never ----
			{
				Code:    `a ||= b`,
				Options: []any{`never`},
				Output:  []string{`a = a || b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `a &&= b`,
				Options: []any{`never`},
				Output:  []string{`a = a && b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (&&=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `a ??= b`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `foo ||= bar`,
				Options: []any{`never`},
				Output:  []string{`foo = foo || bar`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// ---- > Suggestions ----
			{
				Code:    `a.b ||= c`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `a.b = a.b || c`},
						},
					},
				},
			},
			{
				Code:    `a[b] ||= c`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `a[b] = a[b] || c`},
						},
					},
				},
			},
			{
				Code:    `a['b'] ||= c`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `a['b'] = a['b'] || c`},
						},
					},
				},
			},
			{
				Code:    `this.prop ||= 0`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `this.prop = this.prop || 0`},
						},
					},
				},
			},
			{
				Code:    `with (object) a ||= b`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `with (object) a = a || b`},
						},
					},
				},
			},
			// ---- > Parenthesis ----
			{
				Code:    `(a) ||= b`,
				Options: []any{`never`},
				Output:  []string{`(a) = a || b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `a ||= (b)`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `(a ||= b)`,
				Options: []any{`never`},
				Output:  []string{`(a = a || b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 2, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- > Comments ----
			{
				Code:    `/* before */ a ||= b`,
				Options: []any{`never`},
				Output:  []string{`/* before */ a = a || b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a ||= b // after`,
				Options: []any{`never`},
				Output:  []string{`a = a || b // after`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:    `a /* before */ ||= b`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a ||= /* after */ b`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- > Precedence ----
			{
				Code:    `a ||= b && c`,
				Options: []any{`never`},
				Output:  []string{`a = a || b && c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a &&= b || c`,
				Options: []any{`never`},
				Output:  []string{`a = a && (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (&&=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a ||= b || c`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a &&= b && c`,
				Options: []any{`never`},
				Output:  []string{`a = a && (b && c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (&&=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			// ---- > Mixed ----
			{
				Code:    `a ??= b || c`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a ??= b && c`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? (b && c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a ??= b ?? c`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? (b ?? c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a ??= (b || c)`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `a ??= b + c`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? b + c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// ---- https://github.com/eslint/eslint/issues/17173 ----
			{
				Code:    `a ||= b as number;`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b as number);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `a.b.c || (a.b.c = d as number)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `a.b.c ||= d as number`},
						},
					},
				},
			},
			{
				Code: `a.b.c || (a.b.c = (d as number))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `a.b.c ||= (d as number)`},
						},
					},
				},
			},
			{
				Code: `(a.b.c || (a.b.c = d)) as number`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 2, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `(a.b.c ||= d) as number`},
						},
					},
				},
			},
			// ---- 3 or more operands ----
			{
				Code:    `a = a || b || c`,
				Options: []any{`always`},
				Output:  []string{`a ||= b || c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a = a && b && c`,
				Options: []any{`always`},
				Output:  []string{`a &&= b && c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (&&=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a = a ?? b ?? c`,
				Options: []any{`always`},
				Output:  []string{`a ??= b ?? c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a = a || b && c`,
				Options: []any{`always`},
				Output:  []string{`a ||= b && c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a = a || b || c || d`,
				Options: []any{`always`},
				Output:  []string{`a ||= b || c || d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a = a && b && c && d`,
				Options: []any{`always`},
				Output:  []string{`a &&= b && c && d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (&&=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a = a ?? b ?? c ?? d`,
				Options: []any{`always`},
				Output:  []string{`a ??= b ?? c ?? d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (??=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a = a || b || c && d`,
				Options: []any{`always`},
				Output:  []string{`a ||= b || c && d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a = a || b && c || d`,
				Options: []any{`always`},
				Output:  []string{`a ||= b && c || d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    `a = (a) || b || c`,
				Options: []any{`always`},
				Output:  []string{`a ||= b || c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `a = a || (b || c) || d`,
				Options: []any{`always`},
				Output:  []string{`a ||= (b || c) || d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `a = (a || b || c)`,
				Options: []any{`always`},
				Output:  []string{`a ||= (b || c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `a = ((a) || (b || c) || d)`,
				Options: []any{`always`},
				Output:  []string{`a ||= ((b || c) || d)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
		},
	)
}
