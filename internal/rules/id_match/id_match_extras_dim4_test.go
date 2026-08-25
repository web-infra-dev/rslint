package id_match

// TestIdMatchExtrasDim4 locks in the universal edge shapes and JSX forms that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// naming the Dimension 4 row or tsgo AST quirk it covers, so future refactors
// can't silently regress them without breaking a named lock-in. Its siblings
// are id_match_extras_branches_test.go,
// id_match_extras_realuser_test.go and id_match_extras_typescript_test.go.
//
// N/A: Dimension 3 (autofix boundaries) — the rule emits neither fixes nor
// suggestions.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdMatchExtrasDim4(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdMatchRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized receiver ----
			{
				Code:    `(a_1).b = 1`,
				Options: []any{`^[^_]+$`},
			},
			// ---- Dimension 4: parenthesized assignment source ----
			{
				Code:    `x = (a.b_1)`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: optional chain ----
			{
				Code:    `a?.b_1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: optional chain ----
			{
				Code:    `a_1?.();`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: optional chain ----
			{
				Code:    `x = a?.b_1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: body-absent members under onlyDeclarations ----
			{
				Code:    `function f_1(): void;`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
			},
			// ---- Dimension 4: key forms ----
			{
				Code:    `const o = { 'a_1': 1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: key forms ----
			{
				Code:    `const o = { 0: 1, 1e3: 2 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: element access ----
			{
				Code:    `a['b_1'] = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: element access ----
			{
				Code:    "a[`b_1`] = 1;",
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: element access ----
			{
				Code:    `a[0] = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: element access ----
			{
				Code:    `a[Symbol.iterator] = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			// ---- Dimension 4: JSX names are not identifiers ----
			{
				Code:    `<Foo_1 />;`,
				Options: []any{`^[^_]+$`},
				Tsx:     true,
			},
			// ---- Dimension 4: JSX names are not identifiers ----
			{
				Code:    `<div a_1="x" />;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Tsx:     true,
			},
			// ---- Dimension 4: JSX names are not identifiers ----
			{
				Code:    `<foo_1.bar_1 />;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Tsx:     true,
			},
			// ---- Dimension 4: ancestor walk crosses a class static block ----
			{
				Code:    `class C { static { const { a_1 } = o; } }`,
				Options: []any{`^[^_]+$`, map[string]any{"ignoreDestructuring": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver ----
			{
				Code:    `(a_1).b = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    2,
						EndLine:   1,
						EndColumn: 5,
					},
				},
			},
			// ---- Dimension 4: parenthesized receiver ----
			{
				Code:    `((a_1)).b = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    3,
						EndLine:   1,
						EndColumn: 6,
					},
				},
			},
			// ---- Dimension 4: parenthesized receiver ----
			{
				Code:    `(a).b_1 = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion receiver ----
			{
				Code:    `a_1!.b = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion receiver ----
			{
				Code:    `a!.b_1 = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    4,
						EndLine:   1,
						EndColumn: 7,
					},
				},
			},
			// ---- Dimension 4: TS type-expression wrappers ----
			{
				Code:    `(a_1 as any).b = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    2,
						EndLine:   1,
						EndColumn: 5,
					},
				},
			},
			// ---- Dimension 4: TS type-expression wrappers ----
			{
				Code:    `(a as any).b_1 = 1`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Dimension 4: TS type-expression wrappers ----
			{
				Code:    `let s = x satisfies T_1;`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'T_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			// ---- Dimension 4: optional chain ----
			{
				Code:    `a_1?.b;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
				},
			},
			// ---- Dimension 4: optional-chain assignment preserves its ChainExpression boundary ----
			{
				Code:    `a_1?.b_1 = a_2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_2' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Dimension 4: nested optional chain remains inside ChainExpression ----
			{
				Code:    `a_1?.b_1.c_1 = a_2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 4,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_2' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			// ---- Dimension 4: parentheses terminate ChainExpression ----
			{
				Code:    `(a_1?.b_1).c_1 = a_2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    2,
						EndLine:   1,
						EndColumn: 5,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'c_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_2' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			// ---- Dimension 4: key forms ----
			{
				Code:    `const o = { a_1: 1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// ---- Dimension 4: key forms ----
			{
				Code:    `const o = { [a_1]: 1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			// ---- Dimension 4: key forms ----
			{
				Code:    `class C { #a_1 = 1; }`,
				Options: []any{`^[^_]+$`, map[string]any{"classFields": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Dimension 4: element access ----
			{
				Code:    `a[b_1] = 1;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    3,
						EndLine:   1,
						EndColumn: 6,
					},
				},
			},
			// ---- Dimension 4: class declaration vs class expression ----
			{
				Code:    `class C_1 {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'C_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			// ---- Dimension 4: class declaration vs class expression ----
			{
				Code:    `const D = class E_1 {};`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'E_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    17,
						EndLine:   1,
						EndColumn: 20,
					},
				},
			},
			// ---- Dimension 4: function forms ----
			{
				Code:    `function f_1() {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'f_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 13,
					},
				},
			},
			// ---- Dimension 4: function forms ----
			{
				Code:    `const g = function h_1() {};`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'h_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			// ---- Dimension 4: function forms ----
			{
				Code:    `const i_1 = () => {};`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'i_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			// ---- Dimension 4: function forms ----
			{
				Code:    `class C { m_1() {} }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'm_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 14,
					},
				},
			},
			// ---- Dimension 4: function forms ----
			{
				Code:    `class C { p_1 = () => {}; }`,
				Options: []any{`^[^_]+$`, map[string]any{"classFields": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'p_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 14,
					},
				},
			},
			// ---- Dimension 4: async and generator forms ----
			{
				Code:    `async function a_1() {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			// ---- Dimension 4: async and generator forms ----
			{
				Code:    `function* g_1() {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'g_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 14,
					},
				},
			},
			// ---- Dimension 4: async and generator forms ----
			{
				Code:    `async function* ag_1() {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'ag_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    17,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			// ---- Dimension 4: async and generator forms ----
			{
				Code:    `const o = { async *m_1() {} };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'm_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			// ---- Dimension 4: same-kind nesting ----
			{
				Code:    `class A_1 { m() { class B_1 {} } }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'A_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'B_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    25,
						EndLine:   1,
						EndColumn: 28,
					},
				},
			},
			// ---- Dimension 4: same-kind nesting ----
			{
				Code:    `function f_1() { function g_1() {} }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'f_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 13,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'g_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},
			// ---- Dimension 4: same-kind nesting ----
			{
				Code:    `const { a_1: { b_1 } } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			// ---- Dimension 4: ancestor walk crosses a function body ----
			{
				Code:    `const { a_1 = () => { let b_1; } } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"ignoreDestructuring": true, "properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'b_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 30,
					},
				},
			},
			// ---- Dimension 4: spread and rest ----
			{
				Code:    `const o = { ...s_1 };`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 's_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			// ---- Dimension 4: spread and rest ----
			{
				Code:    `const { ...r_1 } = o;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'r_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Dimension 4: spread and rest ----
			{
				Code:    `function f(...rest_1) {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'rest_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			// ---- Dimension 4: empty forms ----
			{
				Code: `class C_1 {}
function f_1() {}
const {} = o;
f_1();`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'C_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 10,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'f_1' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    10,
						EndLine:   2,
						EndColumn: 13,
					},
				},
			},
			// ---- Dimension 4: body-absent members ----
			{
				Code:    `declare function d_1(): void;`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'd_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			// ---- Dimension 4: body-absent members ----
			{
				Code:    `abstract class A { abstract m_1(): void; }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'm_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    29,
						EndLine:   1,
						EndColumn: 32,
					},
				},
			},
			// ---- Dimension 4: body-absent members ----
			{
				Code: `function o_1(a: string): void;
function o_1(a: number): void;
function o_1(a: any): void {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'o_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 13,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'o_1' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    10,
						EndLine:   2,
						EndColumn: 13,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'o_1' does not match the pattern '^[^_]+$'.`,
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 13,
					},
				},
			},
			// ---- Dimension 4: body-absent overloads are TSDeclareFunction under onlyDeclarations ----
			{
				Code: `function o_1(a: string): void;
function o_1(a: number): void;
function o_1(a: any): void {}`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'o_1' does not match the pattern '^[^_]+$'.`,
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 13,
					},
				},
			},
			// ---- Dimension 4: JSX names are not identifiers ----
			{
				Code:    `<Foo bar={baz_1} />;`,
				Options: []any{`^[^_]+$`},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'baz_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// ---- Dimension 4: ancestor walk crosses a class static block ----
			{
				Code:    `class C { static { const { a_1 } = o; } }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    28,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			// ---- Dimension 4: TypeScript-only declaration forms ----
			{
				Code:    `type X_1 = number;`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'X_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 9,
					},
				},
			},
			// ---- Dimension 4: TypeScript-only declaration forms ----
			{
				Code:    `enum E_1 { A_1 }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'E_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    6,
						EndLine:   1,
						EndColumn: 9,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'A_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- Dimension 4: TypeScript-only declaration forms ----
			{
				Code:    `namespace N_1 { export const a = 1; }`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'N_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 14,
					},
				},
			},
			// ---- Dimension 4: TypeScript-only declaration forms ----
			{
				Code:    `function f<T_1>(): void {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'T_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
		},
	)
}
