package prefer_array_flat_map_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat_map"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const preferArrayFlatMapMessage = "Prefer `.flatMap(…)` over `.map(…).flat()`."

// TestPreferArrayFlatMapUpstream migrates the full valid/invalid suite from
// upstream test/prefer-array-flat-map.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the prefer_array_flat_map_extras_test.go file.
func TestPreferArrayFlatMapUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_flat_map.PreferArrayFlatMapRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream snapshot: valid ----
			{Code: `const bar = [1,2,3].map()`},
			{Code: `const bar = [1,2,3].map(i => i)`},
			{Code: `const bar = [1,2,3].map((i) => i)`},
			{Code: `const bar = [1,2,3].map((i) => { return i; })`},
			{Code: `const bar = foo.map(i => i)`},
			{Code: `const bar = foo.map?.(i => [i]).flat()`},
			{Code: `const bar = foo.map(i => [i])?.flat()`},
			{Code: `const bar = foo.map(i => [i]).flat?.()`},
			{Code: `const bar = [[1],[2],[3]].flat()`},
			{Code: `const bar = [1,2,3].map(i => [i]).sort().flat()`},
			{Code: `
let bar = [1,2,3].map(i => [i]);
bar = bar.flat();`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(2)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(1, null)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(Infinity)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(Number.POSITIVE_INFINITY)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(Number.MAX_VALUE)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(Number.MAX_SAFE_INTEGER)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(...[1])`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(0.4 +.6)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(+1)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(foo)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(foo.bar)`},
			{Code: `const bar = [[1],[2],[3]].map(i => [i]).flat(1.00)`},

			// Allowed
			{Code: `Children.map(children, fn).flat()`},
			{Code: `React.Children.map(children, fn).flat()`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Upstream snapshot: invalid ----
			{
				Code:   `const bar = [[1],[2],[3]].map(i => [i]).flat()`,
				Output: []string{`const bar = [[1],[2],[3]].flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID, Message: preferArrayFlatMapMessage,
					Line: 1, Column: 27, EndLine: 1, EndColumn: 47,
				}},
			},
			{
				Code:   `const bar = [[1],[2],[3]].map(i => [i]).flat(1,)`,
				Output: []string{`const bar = [[1],[2],[3]].flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 27, EndLine: 1, EndColumn: 49,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map(i => [i]).flat()`,
				Output: []string{`const bar = [1,2,3].flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 41,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map((i) => [i]).flat()`,
				Output: []string{`const bar = [1,2,3].flatMap((i) => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 43,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map((i) => { return [i]; }).flat()`,
				Output: []string{`const bar = [1,2,3].flatMap((i) => { return [i]; })`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 55,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map(foo).flat()`,
				Output: []string{`const bar = [1,2,3].flatMap(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 36,
				}},
			},
			{
				Code:   `const bar = foo.map(i => [i]).flat()`,
				Output: []string{`const bar = foo.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 17, EndLine: 1, EndColumn: 37,
				}},
			},
			{
				Code:   `const bar = foo?.map(i => [i]).flat()`,
				Output: []string{`const bar = foo?.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 18, EndLine: 1, EndColumn: 38,
				}},
			},
			{
				Code:   `const bar = { map: () => {} }.map(i => [i]).flat()`,
				Output: []string{`const bar = { map: () => {} }.flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 31, EndLine: 1, EndColumn: 51,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map(i => i).map(i => [i]).flat()`,
				Output: []string{`const bar = [1,2,3].map(i => i).flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 33, EndLine: 1, EndColumn: 53,
				}},
			},
			{
				Code:   `const bar = [1,2,3].sort().map(i => [i]).flat()`,
				Output: []string{`const bar = [1,2,3].sort().flatMap(i => [i])`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 28, EndLine: 1, EndColumn: 48,
				}},
			},
			{
				Code:   `const bar = (([1,2,3].map(i => [i]))).flat()`,
				Output: []string{`const bar = (([1,2,3].flatMap(i => [i])))`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 23, EndLine: 1, EndColumn: 45,
				}},
			},
			{
				Code: `
let bar = [1,2,3].map(i => {
	return [i];
}).flat();`,
				Output: []string{`
let bar = [1,2,3].flatMap(i => {
	return [i];
});`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      2, Column: 19, EndLine: 4, EndColumn: 10,
				}},
			},
			{
				Code: `
let bar = [1,2,3].map(i => {
	return [i];
})
.flat();`,
				Output: []string{`
let bar = [1,2,3].flatMap(i => {
	return [i];
});`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      2, Column: 19, EndLine: 5, EndColumn: 8,
				}},
			},
			{
				Code: `
let bar = [1,2,3].map(i => {
	return [i];
}) // comment
.flat();`,
				Output: []string{`
let bar = [1,2,3].flatMap(i => {
	return [i];
});`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      2, Column: 19, EndLine: 5, EndColumn: 8,
				}},
			},
			{
				Code: `
let bar = [1,2,3].map(i => {
	return [i];
}) // comment
.flat(); // other`,
				Output: []string{`
let bar = [1,2,3].flatMap(i => {
	return [i];
}); // other`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      2, Column: 19, EndLine: 5, EndColumn: 8,
				}},
			},
			{
				Code: `
let bar = [1,2,3]
	.map(i => { return [i]; })
	.flat();`,
				Output: []string{`
let bar = [1,2,3]
	.flatMap(i => { return [i]; });`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      3, Column: 3, EndLine: 4, EndColumn: 9,
				}},
			},
			{
				Code: `
let bar = [1,2,3].map(i => { return [i]; })
	.flat();`,
				Output: []string{`
let bar = [1,2,3].flatMap(i => { return [i]; });`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      2, Column: 19, EndLine: 3, EndColumn: 9,
				}},
			},
			{
				Code:   `let bar = [1,2,3] . map( x => y ) . flat () // 🤪`,
				Output: []string{`let bar = [1,2,3] . flatMap( x => y ) // 🤪`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 44,
				}},
			},
			{
				Code:   `const bar = [1,2,3].map(i => [i]).flat(1);`,
				Output: []string{`const bar = [1,2,3].flatMap(i => [i]);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      1, Column: 21, EndLine: 1, EndColumn: 42,
				}},
			},
			{
				Code: `
const foo = bars
	.filter(foo => !!foo.zaz)
	.map(foo => doFoo(foo))
	.flat();`,
				Output: []string{`
const foo = bars
	.filter(foo => !!foo.zaz)
	.flatMap(foo => doFoo(foo));`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Line:      4, Column: 3, EndLine: 5, EndColumn: 9,
				}},
			},
		},
	)
}
