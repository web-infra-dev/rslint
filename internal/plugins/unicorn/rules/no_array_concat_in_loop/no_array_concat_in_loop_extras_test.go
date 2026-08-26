// TestNoArrayConcatInLoopExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case points at the specific
// branch, Dimension 4 row, or upstream issue it covers.
package no_array_concat_in_loop_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_concat_in_loop"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayConcatInLoopExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_concat_in_loop.NoArrayConcatInLoopRule,
		[]rule_tester.ValidTestCase{
			// Locks in isGlobalScopeVariable: top-level let/var declarations
			// belong to ESLint's global scope in script files and are excluded.
			{
				Code:            `let result = []; for (const chunk of chunks) { result = result.concat(chunk); }`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `var result = []; for (const chunk of chunks) { result = result.concat(chunk); }`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `{ var result = []; for (const chunk of chunks) { result = result.concat(chunk); } }`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `let result = []; for (const chunk of chunks) { result = result.concat(chunk); }`,
				FileName:        "file.mjs",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `module.exports = {}; let result = []; for (const chunk of chunks) { result = result.concat(chunk); }`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},

			// Locks in getArrayConcatInLoop arm 1: only simple equals assignments match.
			jsValid(`let result = []; for (const chunk of chunks) { result += result.concat(chunk); }`),

			// Locks in getArrayConcatInLoop arm 2: the left-hand side must be an identifier.
			jsValid(`let result = []; for (const chunk of chunks) { box.result = result.concat(chunk); }`),

			// Locks in getArrayConcatInLoop arm 3: require a direct, non-optional dot concat call with an argument.
			jsValid(`let result = []; for (const chunk of chunks) { result = result.notConcat(chunk); }`),
			jsValid(`let result = []; for (const chunk of chunks) { result = result.concat; }`),
			jsValid(`let result = []; for (const chunk of chunks) { result = result.concat?.(chunk); }`),
			jsValid(`let result = []; for (const chunk of chunks) { result = (result?.concat)(chunk); }`),
			jsValid("let result = []; for (const chunk of chunks) { result = result[`concat`](chunk); }"),

			// Locks in getArrayConcatInLoop arm 4: the receiver must resolve from an identifier.
			jsValid(`let result = []; for (const chunk of chunks) { result = getResult().concat(chunk); }`),

			// Locks in getArrayConcatInLoop arm 5: both identifiers must resolve to the same variable.
			jsValid(`let result = []; let other = []; for (const chunk of chunks) { result = other.concat(chunk); }`),
			jsValid(`for (const chunk of chunks) { missing = missing.concat(chunk); }`),

			// Locks in the upstream mutable empty-array variable gate: exactly one let/var identifier initialized to an empty array.
			jsValid(`var result = []; var result = []; for (const chunk of chunks) { result = result.concat(chunk); }`),
			jsValid(`function append(result) { for (const chunk of chunks) { result = result.concat(chunk); } }`),
			jsValid(`let {result} = source; for (const chunk of chunks) { result = result.concat(chunk); }`),
			jsValid(`using result = []; for (const chunk of chunks) { result = result.concat(chunk); }`),
			jsValid(`let result = [,]; for (const chunk of chunks) { result = result.concat(chunk); }`),

			// Locks in getNearestLoop: a function boundary wins over an outer loop.
			jsValid(`let result = []; for (const chunk of chunks) { (() => { result = result.concat(chunk); })(); }`),
			jsValid(`let result = []; for (const chunk of chunks) { class Box { method() { result = result.concat(chunk); } } }`),
			jsValid(`let result = []; for (const chunk of chunks) { const append = async function* () { result = result.concat(chunk); }; }`),

			// Locks in the destructuring-default gate: TypeScript parses these
			// as equals binary expressions, but ESTree makes them
			// `AssignmentPattern` nodes that upstream never visits.
			jsValid(`let result = []; for (const chunk of chunks) { [result = result.concat(chunk)] = source; }`),
			jsValid(`let result = []; for (const chunk of chunks) { [[result = result.concat(chunk)]] = source; }`),
			jsValid(`let result = []; for (const chunk of chunks) { ({value: result = result.concat(chunk)} = source); }`),
			jsValid(`let result = []; for (const chunk of chunks) { [...[result = result.concat(chunk)]] = source; }`),
			jsValid(`let result = []; for (const chunk of chunks) { for ([result = result.concat(chunk)] of sources) {} }`),

			// Locks in the loop-body declaration guard.
			jsValid(`for (const chunk of chunks) { let result = []; result = result.concat(chunk); }`),

			// ---- Dimension 4: optional receiver wrappers remain non-matches ----
			tsValid(`let result = []; for (const chunk of chunks) { result = (result as any)?.concat(chunk); }`),

			// ---- Dimension 4: non-dot access/key forms do not match ----
			jsValid(`let result = []; for (const chunk of chunks) { result = result[concat](chunk); }`),
			jsValid(`let result = []; for (const chunk of chunks) { result = result[0](chunk); }`),
			jsValid(`let result = []; for (const chunk of chunks) { result = result[Symbol.concat](chunk); }`),

			// ---- Dimension 4: same-name shadowing stays within the resolved declaration ----
			jsValid(`let result = []; for (const chunk of chunks) { { let result = [chunk]; result = result.concat(chunk); } }`),

			// ---- Dimension 4: graceful degradation for empty/spread containers and absent bodies ----
			jsValid(`const values = [...chunks]; for (;;) {}`),
			tsValid(`declare function append(): void; let result = [];`),

			// N/A: private, numeric, and computed declaration keys are unrelated; the rule targets variable identifiers.
			// N/A: class declarations and expressions are not targets; methods only form function boundaries.
			// N/A: autofix boundaries and edit-demand invariance do not apply; this rule has no edits.
		},
		[]rule_tester.InvalidTestCase{
			// A top-level declaration is module-scoped, not global-scoped.
			func() rule_tester.InvalidTestCase {
				testCase := invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`)
				testCase.LanguageOptions = rule.LanguageOptions{SourceType: "module"}
				return testCase
			}(),

			// CommonJS wraps top-level declarations in a function scope.
			func() rule_tester.InvalidTestCase {
				testCase := invalidConcat(`let result = [];
for (const chunk of chunks) {
	result = result.concat(chunk);
}`)
				testCase.LanguageOptions = rule.LanguageOptions{SourceType: "commonjs"}
				return testCase
			}(),

			// A top-level block's let declaration has its own block scope.
			func() rule_tester.InvalidTestCase {
				testCase := invalidConcat(`{
	let result = [];
	for (const chunk of chunks) {
		result = result.concat(chunk);
	}
}`)
				testCase.LanguageOptions = rule.LanguageOptions{SourceType: "script"}
				return testCase
			}(),

			// Nested declarations remain local even when their file is a script.
			func() rule_tester.InvalidTestCase {
				testCase := invalidConcat(`function append(chunks) {
	let result = [];
	for (const chunk of chunks) {
		result = result.concat(chunk);
	}
}`)
				testCase.FileName = "file.js"
				testCase.LanguageOptions = rule.LanguageOptions{SourceType: "script"}
				return testCase
			}(),

			// ---- Dimension 4: nested parentheses and transparent TypeScript wrappers ----
			invalidConcatTS(`let result = ((([]))) as string[];
for (const chunk of chunks) {
	((result)) = (((result as string[])!)).concat(chunk);
}`),
			invalidConcatTS(`let result = [] satisfies string[];
for (const chunk of chunks) {
	result = ((result satisfies string[])).concat(chunk);
}`),

			// An array literal that is not a destructuring target still holds a
			// real assignment expression, which upstream reports.
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	sink = [result = result.concat(chunk)];
}`),

			// ---- Dimension 4: nearest nested loop owns the assignment ----
			invalidConcat(`let result = [];
for (const group of groups) {
	for (const chunk of group) {
		result = result.concat(chunk);
	}
}`),

			// ---- Dimension 4: class static blocks are not function boundaries ----
			invalidConcat(`let result = [];
for (const chunk of chunks) {
	class Box { static { result = result.concat(chunk); } }
}`),

			// Locks in getNearestLoop for an unbraced loop body.
			invalidConcat(`let result = [];
while (condition) result = result.concat(chunk);`),

			// Locks in isNodeInside: a declaration in the loop header is outside the body.
			invalidConcat(`for (let result = []; condition; result = result.concat(chunk)) {}`),

			// ---- Real-user: unicorn#1250 array accumulation in a for-of loop ----
			invalidConcat(`let files = [];
for (const directory of directories) {
	files = files.concat(await readDirectory(directory));
}`),

			// ---- Real-user: unicorn#1250 accumulation under a production loop branch ----
			invalidConcat(`let entries = [];
for (const group of groups) {
	if (group.enabled) {
		entries = entries.concat(group.entries);
	}
}`),
		},
	)
}
