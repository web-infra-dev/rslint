package no_unnecessary_type_constraint

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeConstraintRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeConstraintRule, []rule_tester.ValidTestCase{
		{Code: `function data() {}`},
		{Code: `function data<T>() {}`},
		{Code: `function data<T, U>() {}`},
		{Code: `function data<T extends number>() {}`},
		{Code: `function data<T extends number | string>() {}`},
		{Code: `function data<T extends any | number>() {}`},
		{Code: `
type TODO = any;
function data<T extends TODO>() {}
    `},
		{Code: `const data = () => {};`},
		{Code: `const data = <T,>() => {};`},
		{Code: `const data = <T, U>() => {};`},
		{Code: `const data = <T extends number>() => {};`},
		{Code: `const data = <T extends number | string>() => {};`},
		// ---- Selector-alignment guards ----
		// typescript-eslint's selector is `TSTypeParameterDeclaration > TSTypeParameter[constraint]`.
		// `infer U extends any`, mapped-type `[P in ...]`, and JSDoc `@template` also surface as
		// KindTypeParameter in tsgo but have no TSTypeParameterDeclaration analog in ESTree, so
		// upstream never reports them. Lock the behavior in.
		{Code: `type First<T> = T extends [infer U extends any, ...unknown[]] ? U : never;`},
		{Code: `type Head<T> = T extends [infer U extends unknown, ...unknown[]] ? U : never;`},
		// mapped type: `[K in any]` has constraint `any` on its type parameter K, but it's
		// structurally a TSMappedType, not a TSTypeParameterDeclaration.
		{Code: `type M = { [K in any]: K };`},
		// conditional-type `T extends any ? ... : ...` — `any` is the extends-type of the
		// conditional, not a constraint on a type parameter.
		{Code: `type X<T> = T extends any ? T : never;`},
		{Code: `type IsAny<T> = 0 extends 1 & T ? true : false;`},
		// Default-only `any` / `unknown` (no constraint) — nothing to remove.
		{Code: `function data<T = any>() {}`},
		{Code: `function data<T = unknown>() {}`},
		{Code: `function data<T extends string = any>() {}`},
		// Parenthesized `any` / `unknown` — tsgo keeps a `KindParenthesizedType` wrapper, so
		// the constraint kind isn't `KindAnyKeyword`/`KindUnknownKeyword`. Matches upstream,
		// which sees `TSParenthesizedType` on ESTree.
		{Code: `function data<T extends (any)>() {}`},
		{Code: `function data<T extends (unknown)>() {}`},
		// `keyof any` is a `KindTypeOperator`, not `KindAnyKeyword` — upstream doesn't trigger
		// either.
		{Code: `function data<T extends keyof any>() {}`},
		{Code: `type Idx<T extends keyof any> = Record<T, unknown>;`},
		// ---- JSDoc `@template` guard ----
		// tsgo may or may not parse these into KindTypeParameter nodes depending on file
		// extension / allowJs. The guard ensures that if it does, the rule stays silent (matching
		// ESLint, which never sees JSDoc templates as TSTypeParameterDeclaration).
		{Code: `
/** @template T */
function data() {}
    `},
		{Code: `
/** @template {any} T */
function data() {}
    `},
		{Code: `
/** @template {unknown} T, U */
function data() {}
    `},
		// ---- Mapped-type guard under modifiers / name remap ----
		// Make sure the `KindMappedType` guard survives readonly / negative modifiers and
		// `as` name-remap clauses — those wrap the type parameter differently but parent is
		// still KindMappedType.
		{Code: `type R = { readonly [K in any]: K };`},
		{Code: `type O = { -readonly [K in any]-?: K };`},
		{Code: `type N = { [K in any as ` + "`prefix_${string & K}`" + `]: K };`},
		// ---- InferType guard across nested contexts ----
		// Template-literal type containing infer.
		{Code: "type Head<T> = T extends `${infer U extends any}${string}` ? U : never;"},
		// Multiple `infer` in a tuple extends-clause, each independently guarded.
		{Code: `type Pair<T> = T extends [infer A extends any, infer B extends unknown] ? [A, B] : never;`},
		// ---- Constraint shapes whose kind is NOT KindAnyKeyword/KindUnknownKeyword ----
		// ConditionalType that happens to yield any in one branch — still a KindConditionalType.
		{Code: `function data<T extends (1 extends 1 ? any : never)>() {}`},
		// IntersectionType with any inside — IntersectionType kind at the top.
		{Code: `function data<T extends any & string>() {}`},
		// Readonly-qualified array of any — TypeOperator wraps the ArrayType.
		{Code: `function data<T extends readonly any[]>() {}`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function data<T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `function data<T extends any, U>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T, U>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `function data<T, U extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T, U>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `function data<T extends any, U extends T>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T, U extends T>() {}`,
						},
					},
				},
			},
		},
		// ---- Arrow functions with tsx (requires trailing comma disambiguation) ----
		{
			Code: `const data = <T extends any>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T,>() => {};`,
						},
					},
				},
			},
		},
		{
			Code:     `const data = <T extends any>() => {};`,
			FileName: "file.mts",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T,>() => {};`,
						},
					},
				},
			},
		},
		{
			Code:     `const data = <T extends any>() => {};`,
			FileName: "file.cts",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T,>() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any,>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T,>() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any, >() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T, >() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any ,>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T ,>() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any , >() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T , >() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any = unknown>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T = unknown>() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends any, U extends any>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T, U extends any>() => {};`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 43,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T extends any, U>() => {};`,
						},
					},
				},
			},
		},
		// ---- Unknown constraint ----
		{
			Code: `function data<T extends unknown>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T>() {}`,
						},
					},
				},
			},
		},
		// ---- Non-tsx arrow functions (no trailing comma needed) ----
		{
			Code: `const data = <T extends any>() => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T>() => {};`,
						},
					},
				},
			},
		},
		{
			Code: `const data = <T extends unknown>() => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const data = <T>() => {};`,
						},
					},
				},
			},
		},
		// ---- Class / interface / type alias / member ----
		{
			Code: `class Data<T extends unknown> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Data<T> {}`,
						},
					},
				},
			},
		},
		{
			Code: `const Data = class<T extends unknown> {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const Data = class<T> {};`,
						},
					},
				},
			},
		},
		{
			Code: `
class Data {
  member<T extends unknown>() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `
class Data {
  member<T>() {}
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
const Data = class {
  member<T extends unknown>() {}
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `
const Data = class {
  member<T>() {}
};
      `,
						},
					},
				},
			},
		},
		{
			Code: `interface Data<T extends unknown> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `interface Data<T> {}`,
						},
					},
				},
			},
		},
		{
			Code: `type Data<T extends unknown> = {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type Data<T> = {};`,
						},
					},
				},
			},
		},
		// ---- Variance / const modifiers (TS 4.7 `in`/`out`, TS 5.0 `const`) ----
		{
			Code: `class Data<in T extends any> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Data<in T> {}`,
						},
					},
				},
			},
		},
		{
			Code: `class Data<out T extends unknown> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Data<out T> {}`,
						},
					},
				},
			},
		},
		{
			Code: `function data<const T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<const T>() {}`,
						},
					},
				},
			},
		},
		// ---- Non-arrow with default type (fix must preserve ` = default`) ----
		{
			Code: `function data<T extends any = string>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T = string>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `class Data<T extends unknown = string> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Data<T = string> {}`,
						},
					},
				},
			},
		},
		// ---- Multiple invalid in the same parameter list (non-arrow) ----
		{
			Code: `function data<T extends any, U extends unknown>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T, U extends unknown>() {}`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T extends any, U>() {}`,
						},
					},
				},
			},
		},
		// ---- Method / call / construct signatures ----
		{
			Code: `interface I { m<T extends any>(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `interface I { m<T>(): void; }`,
						},
					},
				},
			},
		},
		{
			Code: `interface I { <T extends any>(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `interface I { <T>(): void; }`,
						},
					},
				},
			},
		},
		{
			Code: `interface I { new <T extends any>(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `interface I { new <T>(): void; }`,
						},
					},
				},
			},
		},
		// ---- FunctionType / ConstructorType ----
		{
			Code: `type F = <T extends any>() => void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type F = <T>() => void;`,
						},
					},
				},
			},
		},
		{
			Code: `type F = new <T extends any>() => void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type F = new <T>() => void;`,
						},
					},
				},
			},
		},
		// ---- async / generator function ----
		{
			Code: `async function data<T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `async function data<T>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `function* data<T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function* data<T>() {}`,
						},
					},
				},
			},
		},
		// ---- Nested arrow functions (in .ts, no JSX conflict) ----
		{
			Code: `const f = <T extends any>() => <U extends any>() => 0;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const f = <T>() => <U extends any>() => 0;`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    33,
					EndLine:   1,
					EndColumn: 46,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const f = <T extends any>() => <U>() => 0;`,
						},
					},
				},
			},
		},
		// ---- tsx but not arrow: no trailing comma should be added ----
		{
			Code: `function data<T extends any>() {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `class Data<T extends any> {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Data<T> {}`,
						},
					},
				},
			},
		},
		// ---- Comments interleaved with `extends` ----
		{
			Code: `function data<T /* a */ extends /* b */ any /* c */>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T /* c */>() {}`,
						},
					},
				},
			},
		},
		// ---- Multi-line type parameter list ----
		{
			Code: `function data<
  T extends any,
  U
>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `function data<
  T,
  U
>() {}`,
						},
					},
				},
			},
		},
		// ---- TypeAlias with multiple invalid type parameters ----
		{
			Code: `type X<T extends any, U extends unknown> = T | U;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    8,
					EndLine:   1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type X<T, U extends unknown> = T | U;`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 40,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type X<T extends any, U> = T | U;`,
						},
					},
				},
			},
		},
		// ---- Exported / default-exported / ambient function ----
		{
			Code: `export function data<T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 35,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `export function data<T>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `export default function <T extends any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `export default function <T>() {}`,
						},
					},
				},
			},
		},
		{
			Code: `declare function data<T extends any>(): void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `declare function data<T>(): void;`,
						},
					},
				},
			},
		},
		{
			Code: `declare class C<T extends any> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `declare class C<T> {}`,
						},
					},
				},
			},
		},
		// ---- FunctionExpression (anonymous + named) ----
		{
			Code: `const f = function <T extends any>() {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const f = function <T>() {};`,
						},
					},
				},
			},
		},
		{
			Code: `const f = function named<T extends any>() {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    26,
					EndLine:   1,
					EndColumn: 39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const f = function named<T>() {};`,
						},
					},
				},
			},
		},
		// ---- ClassExpression with name ----
		{
			Code: `const C = class Named<T extends unknown> {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 40,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const C = class Named<T> {};`,
						},
					},
				},
			},
		},
		// ---- Class with extends / implements ----
		{
			Code: `class Sub<T extends any> extends Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class Sub<T> extends Base {}`,
						},
					},
				},
			},
		},
		// ---- Class method modifiers: static / private / abstract ----
		{
			Code: `class C { static m<T extends any>() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class C { static m<T>() {} }`,
						},
					},
				},
			},
		},
		{
			Code: `class C { private m<T extends any>() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class C { private m<T>() {} }`,
						},
					},
				},
			},
		},
		{
			Code: `abstract class C { abstract m<T extends any>(): void; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `abstract class C { abstract m<T>(): void; }`,
						},
					},
				},
			},
		},
		// ---- Namespace-scoped function ----
		{
			Code: `namespace N { function data<T extends any>() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `namespace N { function data<T>() {} }`,
						},
					},
				},
			},
		},
		// ---- Overload signatures: each declaration is reported independently ----
		{
			Code: `function data<T extends any>(x: T): T;
function data<T extends any>(x: T): T { return x; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `function data<T>(x: T): T;
function data<T extends any>(x: T): T { return x; }`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `function data<T extends any>(x: T): T;
function data<T>(x: T): T { return x; }`,
						},
					},
				},
			},
		},
		// ---- Constraint and default both `any` ----
		{
			Code: `function data<T extends any = any>() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `function data<T = any>() {}`,
						},
					},
				},
			},
		},
		// ---- tsx arrow with 2+ params: no disambiguation trailing comma needed ----
		{
			Code: `const f = <T extends any, U>() => {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const f = <T, U>() => {};`,
						},
					},
				},
			},
		},
		// ---- Interface with heritage clause ----
		{
			Code: `interface I<T extends any> extends J<T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `interface I<T> extends J<T> {}`,
						},
					},
				},
			},
		},
		// ---- Arrow function assigned to a class field ----
		{
			Code: `class C { m = <T extends any>() => {}; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `class C { m = <T>() => {}; }`,
						},
					},
				},
			},
		},
		// ---- Abstract class itself carrying the constraint ----
		{
			Code: `abstract class C<T extends any> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `abstract class C<T> {}`,
						},
					},
				},
			},
		},
		// ---- Overload signatures: only one signature has the unnecessary constraint ----
		{
			Code: `function data<T extends any>(x: T): T;
function data<T>(x: T): T { return x; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output: `function data<T>(x: T): T;
function data<T>(x: T): T { return x; }`,
						},
					},
				},
			},
		},
		// ---- Arrow function as an object-literal property value ----
		{
			Code: `const obj = { m: <T extends any>() => {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const obj = { m: <T>() => {} };`,
						},
					},
				},
			},
		},
		// ---- Object-literal shorthand method (distinct AST from arrow property value) ----
		{
			Code: `const obj = { m<T extends any>() {} };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `const obj = { m<T>() {} };`,
						},
					},
				},
			},
		},
		// ---- Outer TypeParameter is reported, inner mapped-type is guarded ----
		// Proves the guard is parent-scoped and doesn't swallow the outer-level constraint.
		{
			Code: `type X<T extends any> = { [K in keyof T]: T[K] };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    8,
					EndLine:   1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type X<T> = { [K in keyof T]: T[K] };`,
						},
					},
				},
			},
		},
		// ---- Outer TypeParameter is reported, inner FunctionType's TypeParameter also reported;
		// ---- the `infer R` inside the function-type's return is guarded (InferType) ----
		{
			Code: `type X<T extends any> = T extends <U extends any>(x: U) => infer R ? R : never;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    8,
					EndLine:   1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type X<T> = T extends <U extends any>(x: U) => infer R ? R : never;`,
						},
					},
				},
				{
					MessageId: "unnecessaryConstraint",
					Line:      1,
					Column:    36,
					EndLine:   1,
					EndColumn: 49,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeUnnecessaryConstraint",
							Output:    `type X<T extends any> = T extends <U>(x: U) => infer R ? R : never;`,
						},
					},
				},
			},
		},
	})
}
