package require_atomic_updates

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRequireAtomicUpdatesExtras locks in the resolution-sensitive behaviors
// of the Refs-API implementation that the upstream ESLint suite doesn't
// exercise. The rule keys all of its state by symbol, so every case here pins
// the identity contract between declaration-site symbols (taken from the
// declaration nodes) and reference resolution through ctx.Refs — a mismatch
// surfaces as a false positive (a local treated as outer-scope), which no
// upstream case would catch as such.
func TestRequireAtomicUpdatesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAtomicUpdatesRule,
		[]rule_tester.ValidTestCase{
			// Branch lock-in: destructured parameter names resolve to their
			// BindingElement symbols, so they register as locals.
			{Code: `async function x({ count }: { count: number }) { count; await p; count = count + 1; }`},
			// Branch lock-in: destructured local declarations register the same way.
			{Code: `async function x() { let { value } = obj; value; await p; value = value + 1; }`},
			// Branch lock-in: a hoisted function declaration's name is collected
			// from the declaration node's symbol; writes to it stay local.
			{Code: `async function x() { function helper() {} helper; await p; helper = () => {}; }`},
			// Branch lock-in: a local shadowing an outer name binds its own
			// symbol, so state on the outer variable never leaks onto it.
			{Code: `let value: number; async function x() { let value = 0; value; await p; value = value + 1; }`},
			// Branch lock-in: `declare global` symbols resolve through the
			// checker fallback but are filtered like unresolved references,
			// matching ESLint's scope analyzer.
			{Code: `declare global { var mutableFlag: number; }
export async function x() { mutableFlag; await p; mutableFlag = mutableFlag + 1; }`},
		},
		[]rule_tester.InvalidTestCase{
			// Real-user shape: a destructured local that escapes into a nested
			// arrow is race-prone again — escape marking and write resolution
			// must agree on the BindingElement symbol.
			{
				Code: `async function x() { let { value } = obj; const f = () => value; value += await p; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
			// Branch lock-in: shorthand destructuring assignment targets
			// resolve to the outer binding (the value symbol, not the property).
			{
				Code: `let value: number; async function x() { value; await p; ({ value } = obj); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicObjectUpdate"},
				},
			},
			// Branch lock-in: an ambient global resolved via the checker
			// fallback is an outer-scope variable like any other.
			{
				Code: `declare var ambientCounter: number;
async function x() { ambientCounter += await p; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonAtomicUpdate"},
				},
			},
		},
	)
}
