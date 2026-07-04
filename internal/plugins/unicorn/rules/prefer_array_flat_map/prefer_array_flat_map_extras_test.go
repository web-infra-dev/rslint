package prefer_array_flat_map_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat_map"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "prefer-array-flat-map"

// TestPreferArrayFlatMapExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / upstream issue it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestPreferArrayFlatMapExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_flat_map.PreferArrayFlatMapRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: computed access does not match dotted method calls ----
			{Code: `const bar = foo["map"](i => [i]).flat()`},
			{Code: `const bar = foo.map(i => [i])["flat"]()`},
			{Code: "const bar = foo[`map`](i => [i]).flat()"},

			// ---- Dimension 4: raw flat depth must be exactly `1` ----
			{Code: `const bar = foo.map(i => [i]).flat(0x1)`},
			{Code: `const bar = foo.map(i => [i]).flat(1_0)`},
			{Code: `const bar = foo.map(i => [i]).flat(+1)`},
			{Code: `const bar = foo.map(i => [i]).flat(1n)`},
			{Code: `const bar = foo.map(i => [i]).flat(1 as number)`},
			{Code: `const bar = foo.map(callback).flat(1e0)`},
			{Code: `const bar = foo.map(callback).flat(0b1)`},
			{Code: `const bar = foo.map(callback).flat(0o1)`},
			{Code: `const bar = foo.map(callback).flat(1.)`},
			{Code: "const bar = foo.map(callback).flat(`1`)"},

			// ---- Dimension 4: optional call/member exclusions ----
			{Code: `const bar = foo.map?.(i => [i]).flat()`},
			{Code: `const bar = foo.map(i => [i])?.flat()`},
			{Code: `const bar = foo.map(i => [i]).flat?.()`},
			{Code: `const bar = React.Children?.map(children, fn).flat()`},
			{Code: `const bar = Children?.map(children, fn).flat()`},
			{Code: `const bar = (foo?.map)(callback).flat()`},
			{Code: `const bar = (foo?.bar.map)(callback).flat()`},
			{Code: `const bar = (foo.bar?.map)(callback).flat()`},
			{Code: `const bar = (foo.map)?.(callback).flat()`},
			{Code: `const bar = foo?.map?.(callback).flat()`},

			// ---- Dimension 4: TS wrappers between map() and flat() are not transparent upstream ----
			{Code: `const bar = (foo.map(i => [i]) as any).flat()`},
			{Code: `const bar = foo.map(i => [i])!.flat()`},
			{Code: `const bar = (foo.map(i => [i]) satisfies unknown[]).flat()`},

			// ---- Dimension 4: computed access remains unmatched inside ignored React.Children shapes ----
			{Code: `const bar = React.Children["map"](children, fn).flat()`},
			{Code: `const bar = React.Children.map(children, fn)["flat"]()`},

			// ---- Real-user: #751 React.Children.map has no flatMap equivalent ----
			{Code: `const bar = (React.Children).map(children, fn).flat()`},
			{Code: `const bar = ((Children)).map(children, fn).flat()`},

			// ---- Real-user: #848 depth Infinity should not be rewritten to flatMap ----
			{Code: `const bar = a.map(v => v).flat(Infinity)`},

			// ---- Real-user: #315 depth greater than 1 should not be rewritten to flatMap ----
			{Code: `const foo = [[1]].map(i => [i]).flat(2)`},

			// ---- Real-user: #1298 concat-spread case was removed from this rule ----
			{Code: `const foo = [].concat(...array.map(item => [item]))`},

			// N/A: declaration/container forms do not apply; this rule only inspects CallExpression chains.
			// N/A: object/class property key forms do not apply; method names are member-access properties.
			// N/A: cross-scope traversal boundaries do not apply; every CallExpression is checked independently.
			// N/A: empty class/function/destructuring forms do not apply to method-call chain detection.
			// N/A: overload/abstract/declare members do not apply to expression-call chains.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver before map ----
			{
				Code:   `const bar = ((foo)).map(i => [i]).flat()`,
				Output: []string{`const bar = ((foo)).flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 41,
				}},
			},

			// ---- Dimension 4: TS wrappers before map are ordinary receivers ----
			{
				Code:   `const bar = (foo as any).map(i => [i]).flat()`,
				Output: []string{`const bar = (foo as any).flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 26}},
			},
			{
				Code:   `const bar = foo!.map(i => [i]).flat()`,
				Output: []string{`const bar = foo!.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 18}},
			},
			{
				Code:   `const bar = (foo satisfies any[]).map(i => [i]).flat()`,
				Output: []string{`const bar = (foo satisfies any[]).flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 35}},
			},

			// ---- Dimension 4: optional member on map is allowed by upstream ----
			{
				Code:   `const bar = foo?.map(i => [i]).flat()`,
				Output: []string{`const bar = foo?.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 18}},
			},
			{
				Code:   `const bar = foo?.bar.map(callback).flat()`,
				Output: []string{`const bar = foo?.bar.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 22}},
			},
			{
				Code:   `const bar = foo?.bar?.map(callback).flat()`,
				Output: []string{`const bar = foo?.bar?.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 23}},
			},

			// ---- Dimension 4: parenthesized raw depth 1 matches ESTree's paren-flattened literal ----
			{
				Code:   `const bar = foo.map(i => [i]).flat((1))`,
				Output: []string{`const bar = foo.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},
			{
				Code:   `const bar = foo.map(callback).flat(1 /* c */)`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// ---- Dimension 4: comments between map() and flat() are removed with the flat call ----
			{
				Code:   "const bar = foo.map(i => [i]) /* keep? */ .flat()",
				Output: []string{"const bar = foo.flatMap(i => [i])"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},
			{
				Code:   `const bar = foo.map /* map */ (callback).flat()`,
				Output: []string{`const bar = foo.flatMap /* map */ (callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},
			{
				Code:   `const bar = foo.map(callback) /* a */ .flat /* b */ (1)`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// ---- Dimension 4: TS call syntax is preserved or removed like upstream ----
			{
				Code:   `const bar = foo.map<string>(callback).flat()`,
				Output: []string{`const bar = foo.flatMap<string>(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},
			{
				Code:   `const bar = foo.map(callback).flat<number>()`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},
			{
				Code:   `const bar = (<any>foo).map(callback).flat()`,
				Output: []string{`const bar = (<any>foo).flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 24}},
			},

			// ---- Real-user: #316 Apollo/TypeScript chain used to expose overlapping fixes upstream ----
			{
				Code: `
const cartItemsWithVariants: CartItem[] = filteredProducts
	.filter(product => !!product.variants)
	.map(product => transformProduct(product))
	.flat();`,
				Output: []string{`
const cartItemsWithVariants: CartItem[] = filteredProducts
	.filter(product => !!product.variants)
	.flatMap(product => transformProduct(product));`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      4, Column: 3, EndLine: 5, EndColumn: 9,
				}},
			},

			// Locks in upstream create() arm 1: .flat() with no arguments.
			{
				Code:   `const bar = foo.map(callback).flat()`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// Locks in upstream create() arm 2: .flat(1) with raw literal `1`.
			{
				Code:   `const bar = foo.map(callback).flat(1)`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// Locks in upstream create() arm 3: map call may itself be parenthesized.
			{
				Code:   `const bar = (foo.map(callback)).flat()`,
				Output: []string{`const bar = (foo.flatMap(callback))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 18}},
			},

			// Locks in upstream removeMethodCall(): parenthesized flat callee preserves the callee wrapper.
			{
				Code:   `const bar = (foo.map(callback).flat)()`,
				Output: []string{`const bar = (foo.flatMap(callback))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 18}},
			},

			// Locks in upstream isMethodCall(): map's callee may be parenthesized.
			{
				Code:   `const bar = (foo.map)(callback).flat()`,
				Output: []string{`const bar = (foo.flatMap)(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 18}},
			},
			{
				Code:   `const bar = ((foo.map))(callback).flat()`,
				Output: []string{`const bar = ((foo.flatMap))(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 19}},
			},

			// Locks in upstream removeMethodCall(): comments between flat and the call parens are removed.
			{
				Code:   `const bar = foo.map(callback).flat /* c */ ()`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// Locks in upstream create() arm 2: comments around raw depth 1 do not block the rewrite.
			{
				Code:   `const bar = foo.map(callback).flat(/* c */ 1)`,
				Output: []string{`const bar = foo.flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 17}},
			},

			// ---- Dimension 4: ignored React.Children path stays dotted-only and non-optional ----
			{
				Code:   `const bar = React?.Children.map(children, fn).flat()`,
				Output: []string{`const bar = React?.Children.flatMap(children, fn)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 29}},
			},
			{
				Code:   `const bar = React["Children"].map(children, fn).flat()`,
				Output: []string{`const bar = React["Children"].flatMap(children, fn)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 31}},
			},
			{
				Code:   `const bar = globalThis.React.Children.map(children, fn).flat()`,
				Output: []string{`const bar = globalThis.React.Children.flatMap(children, fn)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 39}},
			},
			{
				Code:   `class C { m() { return this.Children.map(children, fn).flat(); } }`,
				Output: []string{`class C { m() { return this.Children.flatMap(children, fn); } }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 38}},
			},

			// ---- Real-user: factory and control-flow receivers are syntactic candidates ----
			{
				Code:   `const bar = getItems().map(callback).flat()`,
				Output: []string{`const bar = getItems().flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 24}},
			},
			{
				Code:   `const bar = (condition ? first : second).map(callback).flat()`,
				Output: []string{`const bar = (condition ? first : second).flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 42}},
			},
			{
				Code:   `const bar = (first, second).map(callback).flat()`,
				Output: []string{`const bar = (first, second).flatMap(callback)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 29}},
			},
			{
				Code:   `function f(items) { return items.map(callback).flat(); }`,
				Output: []string{`function f(items) { return items.flatMap(callback); }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1, Column: 34}},
			},

			// ---- Dimension 4: nested chains report independently and converge after repeated fixes ----
			{
				Code: `const bar = foo.map(x => x.map(y => [y]).flat()).flat()`,
				Output: []string{
					`const bar = foo.flatMap(x => x.map(y => [y]).flat())`,
					`const bar = foo.flatMap(x => x.flatMap(y => [y]))`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 1, Column: 17},
					{MessageId: messageID, Line: 1, Column: 28},
				},
			},
			{
				Code:   `const bar = foo.map(x => [x]).flat().map(y => [y]).flat()`,
				Output: []string{`const bar = foo.flatMap(x => [x]).flatMap(y => [y])`},
				// The outer final flat() is visited before its receiver, so the
				// diagnostics are not source-sorted even though the fixes are.
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 1, Column: 38},
					{MessageId: messageID, Line: 1, Column: 17},
				},
			},
		},
	)
}
