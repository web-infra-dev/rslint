// TestPreferCatchExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment explaining the
// specific branch it covers, so future refactors can't silently regress them
// without breaking a named lock-in.
package prefer_catch_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/prefer_catch"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCatchExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_catch.PreferCatchRule,
		[]rule_tester.ValidTestCase{
			// ---- element-access form is not a PropertyAccessExpression ----
			// N/A: ElementAccessExpression is never matched (callee must be PAE)
			{Code: `prom['then'](fn1, fn2)`},

			// ---- wrong method name does not report ----
			// Locks in upstream name === 'then' check: .catch/.finally with 2 args must not fire
			{Code: `prom.catch(fn1, fn2)`},
			{Code: `prom.finally(fn1, fn2)`},

			// ---- single argument is valid ----
			// Locks in upstream arguments.length >= 2 check
			{Code: `prom.then(fn)`},
			{Code: `prom.then(null)`},
			{Code: `prom.then(undefined)`},

			// ---- Real-user: computed-then is not dotted access ----
			{Code: `promise[methodName](fn1, fn2)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- parenthesized callee is unwrapped ----
			// (prom.then) is a ParenthesizedExpression wrapping the PAE; after
			// SkipOuterExpressions the callee becomes the PAE and matches.
			{
				Code:   `(prom.then)(fn1, fn2)`,
				Output: []string{`(prom.catch(fn2).then)(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 7}},
			},

			// ---- parenthesized receiver object ----
			// (prom).then(fn1, fn2): callee is PAE (name=then) on a paren-wrapped object.
			// The rule listens on the PAE itself (not its Expression), so this matches.
			{
				Code:   `(prom).then(fn1, fn2)`,
				Output: []string{`(prom).catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 8}},
			},

			// ---- optional-chain .then ----
			// prom?.then(fn1, fn2): upstream doesn't exclude optional chains.
			{
				Code:   `prom?.then(fn1, fn2)`,
				Output: []string{`prom?.catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 7}},
			},

			// ---- TS non-null assertion on receiver ----
			// prom!.then(fn1, fn2): callee is PAE on NonNullExpression — still matches.
			{
				Code:   `prom!.then(fn1, fn2)`,
				Output: []string{`prom!.catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 7}},
			},

			// ---- TS type-cast wrapper on receiver ----
			{
				Code:   `(prom as any).then(fn1, fn2)`,
				Output: []string{`(prom as any).catch(fn2).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 15}},
			},

			// ---- parenthesized first arg (null) ----
			// Locks in upstream isNullOrUndef check with paren-wrapped null.
			{
				Code:   `hey.then((null), fn2)`,
				Output: []string{`hey.catch(fn2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 5}},
			},

			// ---- parenthesized first arg (undefined) ----
			{
				Code:   `hey.then((undefined), fn2)`,
				Output: []string{`hey.catch(fn2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 5}},
			},

			// ---- Locks in upstream arguments.length >= 2: exactly 2 args triggers, 3 also triggers ----
			// 3 args: arg[1] removed first → hey.catch(fn2).then(fn1, fn3); then rule fires again
			// on the resulting .then(fn1, fn3) which still has 2 args.
			{
				Code:   `hey.then(fn1, fn2, fn3)`,
				Output: []string{`hey.catch(fn2).then(fn1, fn3)`, `hey.catch(fn2).catch(fn3).then(fn1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 5}},
			},

			// ---- Real-user: typical async/fetch error handler pattern ----
			{
				Code:   `fetch('/api').then(res => res.json(), err => console.error(err))`,
				Output: []string{`fetch('/api').catch(err => console.error(err)).then(res => res.json())`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 15}},
			},

			// ---- Real-user: null error handler in promise chain ----
			{
				Code:   `loadData().then(null, onError)`,
				Output: []string{`loadData().catch(onError)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCatchToThen", Line: 1, Column: 12}},
			},
		},
	)
}
