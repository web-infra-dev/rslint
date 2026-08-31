package unified_signatures

// TestUnifiedSignaturesExtras covers tsgo-specific shapes, real-world regressions,
// and branches not isolated by the upstream suite. The 1:1 upstream migration
// lives in unified_signatures_upstream_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUnifiedSignaturesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: receiver/expression wrappers ----
		// N/A: the rule inspects declarations and type nodes, not runtime receivers.
		// ---- Dimension 4: access/key forms ----
		{Code: `interface I { f(x: string): void; "f"(x: number): void; }`},
		{Code: `interface I { 0(x: string): void; "0"(x: number): void; }`},
		{Code: `declare class C { #f(x: string): void; f(x: number): void; }`},
		{Code: `declare const a: unique symbol; declare const b: unique symbol; interface I { [a](x: string): void; [b](x: number): void; }`},
		{Code: `declare class C { constructor(x: string); "constructor"(x: number); }`},
		// ---- Dimension 4: declaration/container forms ----
		{Code: `const C = class { f(x: string): void; f(x: number): void; };`},
		{Code: `declare module "m" { export default function (x: string): void; function export_default(x: number): void; }`},
		// The raw fallback collision stays scoped to one declaration container.
		{Code: `declare function ExportDefaultDeclaration(x: string): void; declare module "m" { export default function (x: number): void; }`},
		// Grouping by the raw key does not bypass the ordinary compatibility checks.
		{Code: `declare function ExportDefaultDeclaration(x: string): string; export default function (x: number): number;`},
		{Code: `declare function ExportDefaultDeclaration(x: string): void; export default function (y: number): void;`, Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}}},
		{Code: `/** one */ declare function ExportDefaultDeclaration(x: string): void;
/** two */ export default function (x: number): void;`, Options: []any{map[string]any{"ignoreOverloadsWithDifferentJSDoc": true}}},
		// Function expressions and arrows cannot declare overload signatures.
		// Async/generator overload signatures are rejected by TypeScript's grammar.
		// ---- Dimension 4: nesting/traversal boundaries ----
		{Code: `declare function f(x: string): void; namespace N { function f(x: number): void; }`},
		{Code: `interface Outer { f(x: string): void; nested: { f(x: number): void }; }`},
		// ---- Dimension 4: graceful degradation ----
		{Code: `interface Empty {}`},
		{Code: `type Empty = {}; declare class EmptyClass {}`},
		// Spread/rest binding shapes cannot occur as signature containers; destructured
		// parameters are accepted and must not crash or acquire a static name.
		{Code: `interface I { f({x}: {x: string}): void; f({y}: {y: number}): number; }`, Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}}},

		// ---- Real-user: typescript-eslint#740 async return types remain distinct ----
		{Code: `function p(key: string): Promise<string | undefined>;
function p(key: string, defaultValue: string): Promise<string>;
function p(key: string, defaultValue?: string): Promise<string | undefined> { throw 0; }`},
		// Locks in upstream signaturesCanBeUnified return-type equality arm.
		{Code: `function f(x: number): number; function f(x: string): string;`},
		// Locks in upstream differently-named-parameter option arm.
		{Code: `function f(a: number): void; function f(b: string): void;`, Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}}},
		// Locks in upstream different-JSDoc option arm.
		{Code: `/** one */ declare function f(x: number): void;
/** two */ declare function f(x: string): void;`, Options: []any{map[string]any{"ignoreOverloadsWithDifferentJSDoc": true}}},
		// Locks in upstream explicit-this mismatch arm.
		{Code: `function f(): void; function f(this: {}): void;`},
		// Locks in upstream this:void exemption in equal-arity comparisons.
		{Code: `function f(this: void, x: number): void; function f(this: void, x: string): void;`},
		{Code: `function f(this: (void), x: number): void; function f(this: (void), x: string): void;`},
		// Locks in upstream this:void exemption in different-arity comparisons.
		{Code: `function f(this: void): void; function f(this: void, x?: string): void;`},
		// Locks in upstream 2+-extra-parameters required-middle arm.
		{Code: `interface I { f(): void; f(x: string, y: number): void; }`},
		// Locks in upstream common-parameter-type mismatch arm.
		{Code: `interface I { f(x: string): void; f(x: number, y?: boolean): void; }`},
		// Locks in upstream shorter-signature-rest arm.
		{Code: `interface I { f(...x: string[]): void; f(...x: string[], y?: boolean): void; }`},
		// Locks in upstream unequal-rest-sigils arm.
		{Code: `interface I { f(x: string): void; f(...x: number[]): void; }`},
		// Locks in upstream two-differing-parameters arm.
		{Code: `interface I { f(x: string, y: number): void; f(x: number, y: string): void; }`},
		// Locks in upstream outer-type-parameter usage parity arm.
		{Code: `interface I<T> { f(x: T[]): void; f(x: number): void; }`},
		{Code: `interface I<T> { f(x: (T)): void; f(x: string): void; }`},
		{Code: `interface I<T> { f(x: keyof T): void; f(x: string): void; }`},
		{Code: `interface I<T> { f(x: readonly T[]): void; f(x: string): void; }`},
		// Locks in upstream type-parameter-name equality arm.
		{Code: `function f<T>(x: T): void; function f<U>(x: U): void;`},
		// Locks in upstream constraint-kind equality arm.
		{Code: `function f<T extends number>(x: T): void; function f<T extends string>(x: T): void;`},
		// ESTree exposes a null constraint as TSNullKeyword rather than TSLiteralType.
		{Code: `function f<T extends null>(x: string): void; function f<T extends 'x'>(x: number): void;`},
	}

	invalid := []rule_tester.InvalidTestCase{
		// Upstream's anonymous default-export fallback is the raw ESTree node type,
		// so a declaration with that exact name belongs to the same overload group.
		{
			Code: `declare function ExportDefaultDeclaration(x: string): void;
export default function (x: number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Line:      2,
				Column:    26,
			}},
		},
		// Statement order does not change the raw fallback collision.
		{
			Code: `export default function (x: string): void;
declare function ExportDefaultDeclaration(x: number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 2, Column: 43}},
		},
		// The same keying applies inside module blocks.
		{
			Code: `declare module "m" {
  function ExportDefaultDeclaration(x: string): void;
  export default function (x: number): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 3, Column: 28}},
		},
		// Named and anonymous declarations join one larger overload group.
		{
			Code: `declare function ExportDefaultDeclaration(x: string): void;
declare function ExportDefaultDeclaration(x: number): void;
export default function (x: boolean): void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleParameterDifference", Line: 2, Column: 43},
				{MessageId: "singleParameterDifference", Line: 3, Column: 26},
				{MessageId: "singleParameterDifference", Line: 3, Column: 26},
			},
		},
		// Ordinary same-name function overloads retain their grouping after removing
		// the rslint-only key namespace.
		{
			Code: `declare function ExportDefaultDeclaration(x: string): void;
declare function ExportDefaultDeclaration(x: number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 2, Column: 43}},
		},
		// Parentheses and comments around a computed key are transparent in ESTree.
		{
			Code: `declare const key: unique symbol;
interface I {
  [key](x: string): void;
  [(key)](x: number): void;
  [/* comment */ key](x: boolean): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleParameterDifference", Line: 4, Column: 11},
				{MessageId: "singleParameterDifference", Line: 5, Column: 23},
				{MessageId: "singleParameterDifference", Line: 5, Column: 23},
			},
		},
		// The differently-named option compares names only for the same ESTree parameter shape.
		{
			Code:    `interface I { f({x}: {x: string}): void; f(y: number): void; }`,
			Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `{x: string} | number`.",
				Line:      1, Column: 44, EndLine: 1, EndColumn: 53,
			}},
		},
		{
			Code:    `interface I { f(...[value]: [string]): void; f(...values: [string]): void; }`,
			Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}},
		},
		{
			Code:    `declare function f(x: string = ""): void; declare function f(y: number): void;`,
			Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `number`.",
				Line:      1, Column: 62, EndLine: 1, EndColumn: 71,
			}},
		},
		{
			Code:    `declare function f(x: string = ""): void; declare function f(y: number = 0): void;`,
			Options: []any{map[string]any{"ignoreDifferentlyNamedParameters": true}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Message:   "These overloads can be combined into one signature with identical parameters.",
				Line:      1, Column: 43, EndLine: 1, EndColumn: 83,
			}},
		},
		// Parenthesized types are transparent in the unified type text.
		{
			Code: `declare function f(x: (string)): void;
declare function f(x: number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string | number`.",
				Line:      2, Column: 20, EndLine: 2, EndColumn: 29,
			}},
		},
		{
			Code: `declare function f(x: (string | number)): void;
declare function f(x: boolean): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string | number | boolean`.",
				Line:      2, Column: 20, EndLine: 2, EndColumn: 30,
			}},
		},
		// A quoted constructor method is not part of the real constructor group.
		{
			Code: `declare class C {
  constructor(x: string);
  constructor(x: number);
  "constructor"(x: boolean): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string | number`.",
				Line:      3, Column: 15, EndLine: 3, EndColumn: 24,
			}},
		},
		// ---- Real-user: typescript-eslint#12504 duplicate union members ----
		{
			Code: `function f(a: number | string): void;
function f(a: string | boolean): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `number | string | boolean`.",
				Line:      2, Column: 12, EndLine: 2, EndColumn: 31,
			}},
		},
		// Locks in upstream all-parameters-are-same arm and its full signature range.
		{
			Code: `declare function f(a: number): void;
declare function f(a: number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Message:   "These overloads can be combined into one signature with identical parameters.",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 37,
			}},
		},
		// Method signatures report their full ESTree node range, starting at the key.
		{
			Code: `interface I {
  f
  (x: string): void;
  f
  (x: string): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4,
				Column:    3,
			}},
		},
		// A class method value starts at its type-parameter-list opening token.
		{
			Code: `declare class C {
  f<T>
  (x: string): void;
  f<T>
  (x: string): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4,
				Column:    4,
			}},
		},
		// The export wrapper ends after default; the inner declaration begins at function.
		{
			Code: `declare function ExportDefaultDeclaration(x: string): void;
export default
function (x: string): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      3,
				Column:    1,
			}},
		},
		// In larger groups, the message line is the corresponding ESTree signature range.
		{
			Code: `declare class C {
  f
  (x: string): void;
  f
  (x: number): void;
  f
  (x: boolean): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 3 can be combined into one signature taking `string | number`.", Line: 5},
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 3 can be combined into one signature taking `string | boolean`.", Line: 7},
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 5 can be combined into one signature taking `number | boolean`.", Line: 7},
			},
		},
		// Locks in upstream optional-extra-parameter arm.
		{
			Code: `interface I {
  f(): void;
  f(x?: number): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "omittingSingleParameter",
				Message:   "These overloads can be combined into one signature with an optional parameter.",
				Line:      3, Column: 5, EndLine: 3, EndColumn: 15,
			}},
		},
		// Locks in upstream rest-extra-parameter arm.
		{
			Code: `interface I {
  f(): void;
  f(...x: number[]): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "omittingRestParameter",
				Message:   "These overloads can be combined into one signature with a rest parameter.",
				Line:      3, Column: 5, EndLine: 3, EndColumn: 19,
			}},
		},
		// Locks in upstream line-qualified message arm for groups of 3+ overloads.
		{
			Code: `interface I {
  f(x: string): void;
  f(x: number): void;
  f(x: boolean): string;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "This overload and the one on line 2 can be combined into one signature taking `string | number`.",
				Line:      3, Column: 5, EndLine: 3, EndColumn: 14,
			}},
		},
		// Locks in upstream computed-key grouping arm.
		{
			Code: `declare const key: unique symbol;
interface I {
  [key](x: string): void;
  [key](x: number): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 4, Column: 9}},
		},
		// Upstream groups non-literal computed keys by its absent Literal.raw value.
		{
			Code: `declare const key: unique symbol;
interface I {
  [+key](x: string): void;
  [-key](x: number): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 4, Column: 10, EndLine: 4, EndColumn: 19}},
		},
		{
			Code: `declare const E: any;
interface I {
  [E.key](x: string): void;
  [E["key"]](x: number): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 4, Column: 14, EndLine: 4, EndColumn: 23}},
		},
		// Locks in upstream call- and construct-signature keys as separate groups.
		{
			Code: `interface I {
  (x: string): void;
  (x: number): void;
  new (x: string): I;
  new (x: number): I;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleParameterDifference", Line: 3, Column: 4},
				{MessageId: "singleParameterDifference", Line: 5, Column: 8},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnifiedSignaturesRule, valid, invalid)
}
