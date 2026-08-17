// TestSortKeysExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise: tsgo-specific AST shapes (parenthesized and
// TS-wrapped computed keys, shared class/object-literal node kinds, TS
// interface/type-literal exclusion), real-user shapes pulled from the
// upstream rule's GitHub issue tracker, and upstream branches (numKeys /
// blank-line-across-skips accumulation) that its own suite doesn't reach.
// Each case carries an inline comment pointing at the specific thing it
// covers, so future refactors can't silently regress them without breaking
// a named lock-in. Upstream's migrated valid/invalid suite lives in
// sort_keys_upstream_test.go.
//
// Dimension 3 (autofix boundaries) is N/A: this rule has no autofix or
// suggestion — it is a suggestion-type rule with no `fix`.
package sort_keys

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortKeysExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortKeysRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: empty object literal ----
			{Code: "var obj = {};"},
			// ---- Dimension 1: single-property object (below minKeys) ----
			{Code: "var obj = { b: 1 };"},
			// ---- Dimension 2: class methods are not object-literal members ----
			{Code: "class C { b() {} a() {} }"},
			// ---- Dimension 2: class setters are not object-literal members ----
			{Code: "class C { set b(v) {} set a(v) {} }"},
			// ---- Dimension 4: optional-chain computed key is non-simple ----
			{Code: "var obj = { c: 1, [a?.b]: 2, x: 3 };"},
			// ---- Dimension 4: TS interface members are not object-literal properties ----
			{Code: "interface Foo { b: number; a: string; }"},
			// ---- Dimension 4: TS type-literal members are not object-literal properties ----
			{Code: "type Foo = { b: number; a: string; };"},
			// ---- Dimension 4: destructuring pattern, unsorted shorthand names ----
			{Code: "var { b, a } = obj;"},
			// ---- Dimension 4: destructuring pattern with renamed bindings ----
			{Code: "var { b: x, a: y } = obj;"},
			// ---- Real-user: eslint/eslint#18000 — computed member-access key is non-simple ----
			{Code: "var obj = { z: 1, [Foo.identifier]: 2, [Bar.identifier]: 3 };"},
			// ---- Locks in upstream Property(): a blank line before two prevNode-preserving skips (spread, then an ignored computed key) still forms a group boundary for the next real comparison ----
			{
				Code:    "var obj = {\n    c: 1,\n\n    ...z,\n    [a+b]: 1,\n    b: 1\n};",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true, "ignoreComputedKeys": true}},
			},
			// ---- Locks in upstream numKeys computation: below minKeys without the spread stays unenforced ----
			{
				Code:    "var obj = { b: 1, a: 2 };",
				Options: []any{"asc", map[string]any{"minKeys": 3}},
			},
			// ---- Blank line built from carriage returns alone is a group boundary ----
			{
				Code:    "var obj = {\r    b: 1,\r\r    a: 2\r};",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
			},
			// ---- Blank line built from U+2028 line separators is a group boundary ----
			{
				Code:    "var obj = {\u2028    b: 1,\u2028\u2028    a: 2\u2028};",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
			},
			// ---- Blank line built from U+2029 paragraph separators is a group boundary ----
			{
				Code:    "var obj = {\u2029    b: 1,\u2029\u2029    a: 2\u2029};",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: async method ----
			{
				Code: "var obj = { async b() {}, async a() {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},
			// ---- Dimension 1: generator method ----
			{
				Code: "var obj = { *b() {}, *a() {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 1: async generator method ----
			{
				Code: "var obj = { async *b() {}, async *a() {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},
			// ---- Dimension 1: object literal wrapped in `as const` ----
			{
				Code: "var obj = { b: 1, a: 2 } as const;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- Dimension 1: object literal wrapped in `satisfies` ----
			{
				Code: "var obj: Record<string, number> = { b: 1, a: 2 } satisfies Record<string, number>;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- Dimension 1: getter ----
			{
				Code: "var obj = { get b() {}, get a() {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 1: setter ----
			{
				Code: "var obj = { set b(v) {}, set a(v) {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			// ---- Dimension 2: class field initializer is still an object literal ----
			{
				Code: "class C { x = { b: 1, a: 2 }; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 2: 3 levels of nested object literals, independent tracking ----
			{
				Code: "var obj = { c: 1, b: { y: 1, x: { z: 1, w: 2 } }, a: 3 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'x' should be before 'y'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'w' should be before 'z'.", Line: 1, Column: 41, EndLine: 1, EndColumn: 42},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},
			// ---- Dimension 2: object literal as arrow function default parameter ----
			{
				Code: "var f = (x = { b: 1, a: 2 }) => x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Dimension 2: object literal as arrow function expression body ----
			{
				Code: "var f = () => ({ b: 1, a: 2 });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			// ---- Dimension 4: single-parenthesized computed key ----
			{
				Code: "var obj = { c: 1, [(b)]: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},
			// ---- Dimension 4: multi-parenthesized computed key ----
			{
				Code: "var obj = { c: 1, [((b))]: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Dimension 4: TS non-null assertion computed key is non-simple ----
			{
				Code: "var obj = { c: 1, [b!]: 2, a: 3 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: TS `as` computed key is non-simple ----
			{
				Code: "var obj = { c: 1, [(b as any)]: 2, a: 3 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			// ---- Dimension 4: TS `satisfies` computed key is non-simple ----
			{
				Code: "var obj = { c: 1, [(b satisfies string)]: 2, a: 3 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 1, Column: 46, EndLine: 1, EndColumn: 47},
				},
			},
			// ---- Real-user: eslint/eslint#19153 — computed identifier key sorts by its own name ----
			{
				Code: "var obj = { eName: 63, [aName]: 34 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'aName' should be before 'eName'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Locks in upstream numKeys computation: a spread element counts toward minKeys ----
			{
				Code:    "var obj = { b: 1, a: 2, ...z };",
				Options: []any{"asc", map[string]any{"minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- Schema defaults: an explicit empty options object behaves identically to no options ----
			{
				Code:    "var obj = { c: 1, a: 2 };",
				Options: []any{"asc", map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- A carriage return and the line feed behind it end one line, so these properties are one group ----
			{
				Code:    "var obj = {\r\n    b: 1,\r\n    a: 2\r\n};",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 6},
				},
			},
			// ---- Case-insensitive order lowercases the way JavaScript does: `İ` carries a combining dot above, which sorts after a bare `i` ----
			{
				Code:    "var obj = { 'İ': 1, i: 2 };",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. 'i' should be before 'İ'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},
			// ---- Case-insensitive order lowercases a sigma that ends a word to its final form, which sorts before a plain sigma ----
			{
				Code:    "var obj = { 'ασ': 1, 'ΑΣ': 2 };",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. 'ΑΣ' should be before 'ασ'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- A numeric key is named the way JavaScript writes the number, which stays exponential below 1e-6 ----
			{
				Code: "var obj = { 1e-7: 1, '1a': 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '1a' should be before '1e-7'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- A numeric key past the twenty-first place is named in exponential notation too ----
			{
				Code: "var obj = { 1e21: 1, '1a': 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '1a' should be before '1e+21'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
				},
			},
		},
	)
}
