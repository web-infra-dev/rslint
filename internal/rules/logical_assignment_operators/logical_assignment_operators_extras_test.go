package logical_assignment_operators

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestLogicalAssignmentOperatorsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in. The 1:1 migration of upstream's own suite lives in
// logical_assignment_operators_upstream_test.go.
//
// N/A: the Dimension 4 declaration/container rows (class declaration vs class
// expression, function declaration vs expression vs arrow vs method, async and
// generator variants) — this rule matches assignments, logical expressions,
// and `if` statements, and never inspects a function or class head.
func TestLogicalAssignmentOperatorsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&LogicalAssignmentOperatorsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional chain (tsgo flags the link, ESTree wraps in ChainExpression) ----
			{Code: `a?.b || (a?.b = c)`},
			{Code: `a?.() || (a = b)`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (Boolean?.(a)) a = b`},
			// ---- Dimension 4: TypeScript assertion wrappers (node types upstream never matches) ----
			{Code: `a! || (a = b)`},
			{Code: `(a satisfies T) || (a = b)`},
			{Code: `a = a! || b`},
			{Code: `a!.b = a.b || c`},
			{Code: `a = (a as any) || b`},
			{Code: `a[b!] || (a[b!] = c)`},
			{Code: `a.b! || (a.b = c)`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a!) a = b`},
			// ---- Dimension 4: access and key forms ----
			{Code: `class C { #p; m() { this['#p'] || (this.#p = v) } }`},
			// ---- Dimension 4: graceful degradation ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (Boolean(...a)) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (Boolean()) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a) { }`},
			{Code: `[a] = a || b`},
			{Code: `({ a } = a || b)`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a) [a] = b`},
			// ---- Dimension 4: options coverage ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: false}}, Code: `if (a) a = b`},
			{Options: []any{`always`}, Code: `if (a) a = b`},
			{Options: []any{`never`}, Code: `a = a || b`},
			// ---- Locks in upstream getExistence() arms ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (!!!a) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (!(a == null)) a = b`},
			// ---- Locks in upstream isUndefined() arms ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a === null || a === void 0n) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a === null || a === void 1) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a === null || a === void "0") a = b`},
			// ---- Locks in upstream isBooleanCast() arms ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (Boolean(a, b)) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (obj.Boolean(a)) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `function f(Boolean) { if (Boolean(a)) a = b }`},
			// ---- Dimension 4: `undefined` and `Boolean` resolved against file declarations ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `type Boolean = 1; if (Boolean(a)) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `declare const undefined: any; if (a == undefined) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `namespace N { const undefined = 1; if (a == undefined) a = b }`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `var undefined; if (a == undefined) a = b`},
			// ---- Locks in upstream isImplicitNullishComparison() arms ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a != null) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a == 0) a = b`},
			// ---- Locks in upstream isExplicitNullishComparison() arms ----
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a === null && a === undefined) a = b`},
			{Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}}, Code: `if (a === null || a === null) a = b`},
			// ---- Locks in upstream getLeftmostOperand() arms ----
			{Code: `a = (a || b) || c`},
			{Code: `a = ((a || b)) || c`},
			{Code: `a = (a && b) || c`},
			// ---- Real-user: eslint#17597 (`a = b || a` stays untouched) ----
			{Code: `a = b || a`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver (tsgo keeps the node, ESTree drops it) ----
			{
				Code:   `(a).b || ((a).b = c)`,
				Output: []string{`(a).b ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:   `a = ((a)) || b`,
				Output: []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `((a)) = a || b`,
				Output: []string{`((a)) ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `a || (((a)) = b)`,
				Output: []string{`((a)) ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `if ((a).b) (a).b = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`(a).b &&= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Dimension 4: optional chain (tsgo flags the link, ESTree wraps in ChainExpression) ----
			{
				Code: `(a?.b).c || ((a?.b).c = d)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `(a?.b).c ||= d`},
						},
					},
				},
			},
			{
				Code: `a?.b = a?.b || c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `useLogicalOperator`, Output: `a?.b ||= c`},
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
			// ---- Dimension 4: TypeScript assertion wrappers (node types upstream never matches) ----
			{
				Code:    `a ||= b!`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b!)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `a ||= b as T`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b as T)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `a ||= b satisfies T`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b satisfies T)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- Dimension 4: access and key forms ----
			{
				Code:   `a.b || (a['b'] = c)`,
				Output: []string{`a['b'] ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `a['b'] || (a.b = c)`,
				Output: []string{`a.b ||= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `a[0] || (a['0'] = b)`,
				Output: []string{`a['0'] ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:   "a[`b`] || (a[`b`] = c)",
				Output: []string{"a[`b`] ||= c"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:   "a['b'] || (a[`b`] = c)",
				Output: []string{"a[`b`] ||= c"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:   `class C { #p; m() { this.#p || (this.#p = v) } }`,
				Output: []string{`class C { #p; m() { this.#p ||= v } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 21, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    `class C { #p; m() { if (this.#p) this.#p = v } }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`class C { #p; m() { this.#p &&= v } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 21, EndLine: 1, EndColumn: 45},
				},
			},
			// ---- Dimension 4: nesting and traversal boundaries ----
			{
				Code:   `a || (b || (b = 0))`,
				Output: []string{`a || (b ||= 0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 7, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `if (a) if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`if (a) a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `class C { static { a || (a = b) } }`,
				Output: []string{`class C { static { a ||= b } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 20, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:   `(() => a || (a = b))`,
				Output: []string{`(() => a ||= b)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `a = a || (b = b || c)`,
				Output: []string{`a ||= (b ||= c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 11, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Dimension 4: options coverage ----
			{
				Code:    `a = a || b`,
				Options: []any{`always`},
				Output:  []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- Locks in upstream getExistence() arms ----
			{
				Code:    `if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
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
				Code:    `if (!!a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
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
				Code:    `if (!Boolean(a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ||=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
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
				Code:    `if (!(a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ||= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ||=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			// ---- Locks in upstream isUndefined() arms ----
			{
				Code:    `if (a === null || a === void 0x0) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `if (a === null || a === void 0.0) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `if (a === null || a === void (0)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `if (a === (null) || a === (undefined)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
				},
			},
			// ---- Locks in upstream isBooleanCast() arms ----
			{
				Code:    `if ((Boolean)(a)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `if (Boolean((a))) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `if (Boolean(a.b)) a.b = c`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a.b &&= c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- Dimension 4: `undefined` and `Boolean` resolved against file declarations ----
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
			// ---- Locks in upstream isImplicitNullishComparison() arms ----
			{
				Code:    `if ((a) == null) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `if (null == a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Locks in upstream isExplicitNullishComparison() arms ----
			{
				Code:    `if ((a === null) || (a === undefined)) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ??= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ??=.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    `if (a.b === null || a.b === undefined) a.b = c`,
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
			// ---- Locks in upstream getLeftmostOperand() arms ----
			{
				Code:   `a = a || b || c || d`,
				Output: []string{`a ||= b || c || d`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:   `a = (a) || b || c`,
				Output: []string{`a ||= b || c`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Locks in the outer-parenthesis arms of the logical fixer ----
			{
				Code:   `[a || (a = 0)]`,
				Output: []string{`[(a ||= 0)]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 2, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:   `a, a || (a = 0)`,
				Output: []string{`a, a ||= 0`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 4, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `var x = a || (a = 0)`,
				Output: []string{`var x = (a ||= 0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 9, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:   `while (a || (a = 0)) {}`,
				Output: []string{`while (a ||= 0) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   "`${a || (a = 0)}`",
				Output: []string{"`${(a ||= 0)}`"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 4, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `(() => a || (a = 0))`,
				Output: []string{`(() => a ||= 0)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:   `x = a || (a = 0)`,
				Output: []string{`x = a ||= 0`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 5, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `for (;a || (a = 0);) {}`,
				Output: []string{`for (;(a ||= 0);) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 7, EndLine: 1, EndColumn: 19},
				},
			},
			// ---- Locks in the right-side parenthesis arms of the never-mode fixer ----
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
				Code:    `a ??= b && c`,
				Options: []any{`never`},
				Output:  []string{`a = a ?? (b && c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (??=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
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
			{
				Code:    `a||=b`,
				Options: []any{`never`},
				Output:  []string{`a= a ||b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:    `a||=b||c`,
				Options: []any{`never`},
				Output:  []string{`a= a ||(b||c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `fn(a ||= b || c)`,
				Options: []any{`never`},
				Output:  []string{`fn(a = a || (b || c))`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 4, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a ||= b ? c : d`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b ? c : d)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `a ||= b = c`,
				Options: []any{`never`},
				Output:  []string{`a = a || (b = c)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			// ---- Locks in the previous-token guard of the if fixer ----
			{
				Code: `id
if (a) (a) = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
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
				Code: `id
if (this.a) this.a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output: []string{`id
this.a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 2, Column: 1, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    `while(1) if (a) a = b`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`while(1) a &&= b`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 10, EndLine: 1, EndColumn: 22},
				},
			},
			// ---- Locks in the with-block arms of cannotBeGetter() and accessesSingleProperty() ----
			// A top-level "use strict" makes the whole file strict, so a bare name in
			// a `with` body can no longer resolve to a property of the object.
			// ESLint rejects this source outright, since `with` is a syntax error in
			// strict code; tsgo parses it, and the rule answers as upstream would.
			{
				Code:   "\"use strict\";\nwith (object) { a = a || b }",
				Output: []string{"\"use strict\";\nwith (object) { a ||= b }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 2, Column: 17, EndLine: 2, EndColumn: 27},
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
				Code:   `with (object) { a || (a = b) }`,
				Output: []string{`with (object) { a ||= b }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `with (object) { obj.a || (obj.a = b) }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `logical`, Message: `Logical expression can be replaced with an assignment (||=).`, Line: 1, Column: 17, EndLine: 1, EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `convertLogical`, Output: `with (object) { obj.a ||= b }`},
						},
					},
				},
			},
			{
				Code:    `with (object) { if (a) a = b }`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`with (object) { a &&= b }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator &&=.`, Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    `with (object) { a ||= b }`,
				Options: []any{`never`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: `unexpected`, Message: `Unexpected logical operator assignment (||=) shorthand.`, Line: 1, Column: 17, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: `separate`, Output: `with (object) { a = a || b }`},
						},
					},
				},
			},
			// ---- Real-user: eslint#19672 (if-statement autofix kept as an autofix) ----
			{
				Code: `if (!a) {
  a = 1
}`,
				Options: []any{`always`, map[string]any{`enforceForIfStatements`: true}},
				Output:  []string{`a ||= 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `if`, Message: `'if' statement can be replaced with a logical operator assignment with operator ||=.`, Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			// ---- Real-user: eslint#17096 (three or more operands) ----
			{
				Code:   `a = a || 1 || 2`,
				Output: []string{`a ||= 1 || 2`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `a = a || (1 || 2)`,
				Output: []string{`a ||= (1 || 2)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: `assignment`, Message: `Assignment (=) can be replaced with operator assignment (||=).`, Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
		},
	)
}

// TestLogicalAssignmentOperatorsEditDemand checks that the diagnostics this
// rule reports are identical whichever edit categories the consumer asks for,
// and that each artifact is materialized only under its own demand. The three
// diagnostics cover the rule's three outcomes: an autofix, a suggestion, and a
// diagnostic whose rewrite is blocked by an interior comment.
func TestLogicalAssignmentOperatorsEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`fixable = fixable || fallback;
suggested.deep.property = suggested.deep.property || fallback;
commented = commented /* keep */ || fallback;`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     LogicalAssignmentOperatorsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return LogicalAssignmentOperatorsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
		}
		for index, diagnostic := range diagnostics {
			if diagnostic.Message.Id != "assignment" {
				t.Fatalf(
					"demand %d diagnostic %d: message id = %q, want assignment",
					demand,
					index,
					diagnostic.Message.Id,
				)
			}
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index, allEditsDiagnostic := range allEdits {
		for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly[index],
			rule.EditDemandAutofix:    autofixOnly[index],
			rule.EditDemandSuggestion: suggestionOnly[index],
		} {
			if got, want := withoutEdits(diagnostic), withoutEdits(allEditsDiagnostic); !reflect.DeepEqual(got, want) {
				t.Errorf(
					"diagnostic %d demand %d changed identity:\ngot  %#v\nwant %#v",
					index,
					demand,
					got,
					want,
				)
			}
		}
		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Errorf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if diagnosticsOnly[index].Suggestions != nil || autofixOnly[index].Suggestions != nil {
			t.Errorf("diagnostic %d: non-suggestion demand materialized suggestions", index)
		}
		if !reflect.DeepEqual(autofixOnly[index].FixesPtr, allEditsDiagnostic.FixesPtr) {
			t.Errorf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if !reflect.DeepEqual(suggestionOnly[index].Suggestions, allEditsDiagnostic.Suggestions) {
			t.Errorf(
				"diagnostic %d: suggestion and all-edits demands produced different suggestions",
				index,
			)
		}
	}

	// The rewrite is two edits, as upstream's fixer generator yields: write
	// the logical operator in front of the `=`, then delete the repeated
	// operand.
	if allEdits[0].FixesPtr == nil || len(*allEdits[0].FixesPtr) != 2 {
		t.Error("identifier target did not produce exactly two fixes")
	}
	if allEdits[0].Suggestions != nil {
		t.Error("fixable diagnostic unexpectedly produced a suggestion")
	}
	if allEdits[1].FixesPtr != nil {
		t.Error("nested member target unexpectedly produced a fix")
	}
	if allEdits[1].Suggestions == nil || len(*allEdits[1].Suggestions) != 1 {
		t.Error("nested member target did not produce exactly one suggestion")
	}
	if allEdits[2].FixesPtr != nil || allEdits[2].Suggestions != nil {
		t.Error("commented assignment unexpectedly produced an edit")
	}
}
