// TestAvoidNewExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.

package avoid_new_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/avoid_new"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestAvoidNewExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&avoid_new.AvoidNewRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS type assertion (as) on callee ----
			// ESLint: TSAsExpression callee → callee.name undefined → no report.
			// tsgo: AsExpression not unwrapped (only parens are) → no report. Matches ESLint.
			{Code: `new (Promise as any)()`},

			// ---- Dimension 4: element-access callee ----
			// Locks in the "callee is not an Identifier" branch: ElementAccessExpression
			// is not an Identifier → no report.
			{Code: `new Promise['resolve']()`},

			// ---- Dimension 4: non-Promise class names ----
			// Locks in the name-inequality branch: callee.Text !== "Promise" → no report.
			{Code: `new MyPromise()`},
			{Code: `new PromiseFactory()`},

			// ---- Real-user: namespace-scoped constructor ----
			// `new Bluebird.Promise()` — callee is PropertyAccessExpression, not Identifier.
			{Code: `new Bluebird.Promise(() => {})`},

			// ---- Real-user: util.promisify alternative ----
			// Static Promise methods are the encouraged replacement; make sure they
			// are not flagged.
			{Code: `Promise.resolve(someCallbackFn())`},
			{Code: `Promise.reject(new Error('oops'))`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callee ----
			// ESLint's ESTree parser drops parens, so `(Promise)` is seen as Identifier.
			// tsgo preserves ParenthesizedExpression; SkipOuterExpressions unwraps it
			// so behavior matches ESLint.
			{
				Code:   `new (Promise)()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 1}},
			},
			// Double-parenthesized callee — SkipOuterExpressions peels all layers.
			{
				Code:   `new ((Promise))()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 1}},
			},

			// ---- Dimension 4: new Promise with type arguments ----
			// TypeArguments live on the NewExpression node, not the callee, so the
			// callee remains a plain Identifier "Promise" → should report.
			{
				Code:   `new Promise<string>((resolve) => { resolve('x') })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 1}},
			},

			// ---- Dimension 4: new Promise deeply nested ----
			// The listener fires on every NewExpression regardless of nesting depth.
			{
				Code:   `foo(bar(new Promise(() => {})))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 9}},
			},

			// Locks in upstream NewExpression arm: callee is Identifier "Promise" → report.
			// Arm: non-Promise name (e.g. 'Horse') → no report, covered by valid cases.
			{
				Code:   `new Promise(function(resolve) {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 1}},
			},

			// ---- Real-user: new Promise inside arrow function ----
			{
				Code:   `const fn = () => new Promise((resolve) => setTimeout(resolve, 0))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 18}},
			},

			// ---- Real-user: new Promise exported from a module ----
			{
				Code:   `export const p = new Promise((resolve, reject) => { reject(new Error('fail')) })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 18}},
			},

			// Locks in: shadowing Promise does not suppress the report — the rule
			// only checks the callee's text, not whether the binding is the global Promise.
			{
				Code:   `const Promise = class {}; const p = new Promise()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNew", Message: avoidNewMessage, Line: 1, Column: 37}},
			},
		},
	)
}
