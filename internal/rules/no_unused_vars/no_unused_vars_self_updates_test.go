// TestNoUnusedVarsSelfUpdates exercises discarded self-updates through varied
// expression trees and nested execution boundaries. The rule must recognize
// the enclosing assignment generically while preserving uses that can escape
// through another scope, a loop iteration, a callback, or a consumed result.
package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsSelfUpdates(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// The write executes in another variable scope and can be observed later.
			{Code: `let x = 0; function update() { x = x + 1; } update();`},

			// A later loop iteration observes the value written by the prior one.
			{Code: `let x = 0; for (let i = 0; i < 2; i++) { x = x + 1; }`},

			// A callback passed to another call is storable and may execute later.
			{Code: `let x; x = consume(() => x);`},
			{Code: `let x; x = consume({ value: () => x });`},
			{Code: `let x; let stored; x = (stored = { value: () => x }); consume(stored);`},

			// Consuming the assignment result makes the read meaningful.
			{Code: `let x = 0; consume(x = x + 1);`},

			// Logical assignments are conditional reads, not discarded self-updates.
			{Code: `let x; x ||= 1;`},

			// TS assertion wrappers are intentionally not erased when identifying
			// a recursive function initializer, matching the parser's ESTree shape.
			{Code: `const f = ((function () { return f(); }) as () => unknown);`},
			{Code: `let x: any = 0; (x as any) = x + 1;`},
		},
		[]rule_tester.InvalidTestCase{
			extraUnusedCase(`let x = []; x = x["concat"](x);`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = []; x = x?.["concat"](x);`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = 0; x = true ? x : 1;`, "x", true, 1, 12, 13, ""),
			extraUnusedCase(`let x = 0; x = [x][0];`, "x", true, 1, 12, 13, ""),
			extraUnusedCase(`let x = 0; x = ({ value: x }).value;`, "x", true, 1, 12, 13, ""),
			extraUnusedCase("let x = ''; x = `${x}`;", "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x; x = new Box(x);`, "x", true, 1, 8, 9, ""),
			extraUnusedCase("let x = ''; x = tag`${x}`;", "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x; x = (() => x)();`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x; x = { value: () => x };`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x; x = [() => x];`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x = 0; (x) += 1;`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = 0; (x)++;`, "x", true, 1, 13, 14, ""),

			// TypeScript expression wrappers inside the RHS must not hide the read.
			extraUnusedCase(`let x: any = []; x = (x as any)["concat"](x);`, "x", true, 1, 18, 19, ""),
			extraUnusedCase(`let x: any = []; x = x!["concat"](x);`, "x", true, 1, 18, 19, ""),
			extraUnusedCase(`let x: any = []; x = (x satisfies any)["concat"](x);`, "x", true, 1, 18, 19, ""),
		},
	)
}
