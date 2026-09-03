// TestNoArrayFrontMutationExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case identifies the upstream
// branch, Dimension 4 row, or real-user scenario it covers.
package no_array_front_mutation_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_front_mutation"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayFrontMutationExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_front_mutation.NoArrayFrontMutationRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream create() arm 1: only CallExpression nodes match.
			valid(`array.shift`, "file.mjs"),
			valid(`array.unshift`, "file.mjs"),

			// Locks in upstream isMethodCall() method-name rejection branch.
			valid(`array.pop()`, "file.mjs"),
			valid(`array.push(value)`, "file.mjs"),

			// ---- Dimension 4: element-access key forms are excluded ----
			valid(`array['shift']()`, "file.mjs"),
			valid("array[`unshift`](value)", "file.mjs"),
			valid(`array[0]()`, "file.mjs"),
			valid(`array[Symbol.iterator]()`, "file.mjs"),

			// ---- Dimension 4: optional calls are excluded ----
			valid(`(array.shift)?.()`, "file.mjs"),
			valid(`array.unshift?.(value)`, "file.mjs"),

			// ---- Dimension 4: authored TypeScript callee wrappers stay visible ----
			valid(`(array.shift as () => void)()`, "file.ts"),
			valid(`(array.unshift satisfies (value: unknown) => number)(value)`, "file.ts"),

			// Locks in upstream ignoredUnshiftCallees branch for stream APIs.
			valid(`stream.unshift(chunk)`, "file.mjs"),
			valid(`process.stdin.unshift(chunk)`, "file.mjs"),

			// Locks in upstream shouldSkipKnownNonArrayReceiver() branch.
			valid(`function consume(queue: Set<number>) { queue.shift(); }`, "file.ts"),
			valid(`function consume(queue: Map<string, number>) { queue.unshift(1); }`, "file.ts"),

			// N/A: declaration/container and property-key forms; the rule only inspects calls.
			// N/A: ancestor walks and body-absent declarations; the rule performs no ancestor traversal.
			// N/A: object spread and binding rest; neither can be a matched dot-method callee.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver wrappers are transparent ----
			invalid(`(array).shift()`, "file.mjs", "shift"),
			invalid(`((array)).unshift(value)`, "file.mjs", "unshift"),

			// ---- Dimension 4: TypeScript receiver wrappers are transparent ----
			invalid(`items!.shift()`, "file.ts", "shift"),
			invalid(`(items as unknown[]).unshift(value)`, "file.ts", "unshift"),
			invalid(`(items satisfies unknown[]).shift()`, "file.ts", "shift"),

			// ---- Dimension 4: optional member access remains reportable ----
			invalid(`items?.shift()`, "file.mjs", "shift"),
			invalid(`items?.unshift(value)`, "file.mjs", "unshift"),

			// ---- Dimension 4: same-kind nesting reports both calls ----
			invalid(`items.unshift(other.shift())`, "file.mjs", "unshift", "shift"),

			// Locks in upstream ignored-callee condition's method guard: shift is not exempt.
			invalid(`process.stdin.shift()`, "file.mjs", "shift"),
			// Locks in the ignored-path rejection branch: other unshift receivers still report.
			invalid(`socket.unshift(chunk)`, "file.mjs", "unshift"),

			// ---- Real-user: issue #1282 proposal uses whitespace before call parens ----
			invalid(
				"const items = [];\nitems.shift ();\nitems.unshift ();",
				"file.mjs",
				"shift",
				"unshift",
			),

			// ---- Real-user: issue #1282 highlights repeated front mutation in hot loops ----
			invalid(
				"for (let index = 0; index < 10_000; index++) {\n\titems.unshift(value);\n}",
				"file.mjs",
				"unshift",
			),
		},
	)
}
