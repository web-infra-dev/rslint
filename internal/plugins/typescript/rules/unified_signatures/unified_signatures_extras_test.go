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
		// A generic constructor keeps its constructor key before scanning <T>.
		// The corresponding overloads are covered in the invalid cases below.
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
		// The outer type parameter can occur in the value annotation of a mapped type.
		{Code: `interface I<T> { f(x: { [K in keyof T]: T }): void; f(x: string): void; }`},
		{Code: `interface I<T> { f(x: { [K in keyof T]: T[] }): void; f(x: string): void; }`},
		// Locks in upstream type-parameter-name equality arm.
		{Code: `function f<T>(x: T): void; function f<U>(x: U): void;`},
		// Locks in upstream constraint-text equality arm.
		{Code: `function f<T extends number>(x: T): void; function f<T extends string>(x: T): void;`},
		// A missing constraint is distinct from a present constraint.
		{Code: `function f<T>(x: T, y: string): void; function f<T extends number>(x: T): void;`},
		// Constraint equality is textual after transparent outer parentheses.
		{Code: `function f<T extends 1 | 2>(x: T, y: string): void; function f<T extends 1|2>(x: T): void;`},
		// Distinct literal constraint texts remain separate.
		{Code: `function f<T extends null>(x: string): void; function f<T extends 'x'>(x: number): void;`},
		// Method-signature diagnostics are disabled on the method-key line, not its parameter line.
		{Code: `interface I {
  f
  (x: string): void;
  // eslint-disable-next-line test
  f
  (x: string): void;
}`},
		// Generic class method values start at <, which remains on the declaration line.
		{Code: `declare class C {
  f<T>
  (x: string): void;
  // eslint-disable-next-line test
  f<T>
  (x: string): void;
}`},
		// Generic constructors use the same function-value range as class methods.
		{Code: `declare class C {
  constructor<T>
  (x: string);
  // eslint-disable-next-line test
  constructor<T>
  (x: string);
}`},
		{Code: `declare class C {
  constructor<T>
  (x: string);
  constructor<T> // eslint-disable-line test
  (x: string);
}`},
	}

	invalid := []rule_tester.InvalidTestCase{
		// JSDoc @overload signatures reach the same type-text comparison through
		// tsgo's reparsed declarations and retain its JSDoc trivia semantics.
		{
			Code: `/**
 * @overload
 * @param {string} value
 * @returns {void}
 *
 * @overload
 * @param {number} value
 * @returns {void}
 */
function f(value) {}`,
			FileName: "file.mjs",
			TSConfig: "tsconfig.allow-js.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string | number`.",
			}},
		},
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
		// Generic constructors must remain in the constructor overload group.
		{
			Code: `declare class C {
  constructor<T>(x: string);
  constructor<T>(x: number);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleParameterDifference", Line: 3}},
		},
		// A generic constructor's function-value range starts at <T>, even when
		// the parameter list begins on the next line.
		{
			Code: `declare class C {
  constructor<T>
  (x: string);
  constructor<T>
  (x: string);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4, Column: 14, EndLine: 5, EndColumn: 15,
			}},
		},
		// A quoted generic constructor-shaped member is a MethodDeclaration in
		// tsgo and retains the upstream function-value range.
		{
			Code: `declare class C {
  "constructor"<T>
  (x: string): void;
  "constructor"<T>
  (x: string): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4, Column: 16, EndLine: 5, EndColumn: 21,
			}},
		},
		// Without type parameters, the quoted form is a Constructor and falls
		// back to the parameter list's opening token.
		{
			Code: `declare class C {
  "constructor"(x: string): void;
  "constructor"(x: string): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      3, Column: 16, EndLine: 3, EndColumn: 34,
			}},
		},
		// The static generic form is a Constructor in tsgo.
		{
			Code: `declare class C {
  static constructor<T>
  (x: string): void;
  static constructor<T>
  (x: string): void;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4, Column: 21, EndLine: 5, EndColumn: 21,
			}},
		},
		// Type-parameter trivia does not move the opening-delimiter range.
		{
			Code: `declare class C {
  private constructor /* gap */ < /* inner */ T >
  (x: string);
  private constructor /* gap */ < /* inner */ T >
  (x: string);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4, Column: 33, EndLine: 5, EndColumn: 15,
			}},
		},
		// Non-generic constructors continue to start at the opening parenthesis.
		{
			Code: `declare class C {
  constructor
  (x: string);
  constructor
  (x: string);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      5, Column: 3, EndLine: 5, EndColumn: 15,
			}},
		},
		// Larger overload groups cite the line where each function value starts.
		{
			Code: `declare class C {
  constructor
  <T>
  (x: string);
  constructor
  <T>
  (x: string);
  constructor
  <T>
  (x: string);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allParametersAreSame", Message: "This overload and the one on line 3 can be combined into one signature with identical parameters.", Line: 6, Column: 3, EndLine: 7, EndColumn: 15},
				{MessageId: "allParametersAreSame", Message: "This overload and the one on line 3 can be combined into one signature with identical parameters.", Line: 9, Column: 3, EndLine: 10, EndColumn: 15},
				{MessageId: "allParametersAreSame", Message: "This overload and the one on line 6 can be combined into one signature with identical parameters.", Line: 9, Column: 3, EndLine: 10, EndColumn: 15},
			},
		},
		// Parenthesized constraints are transparent in typescript-estree and
		// therefore still compare equal.
		{
			Code: `function f<T extends (number)>(x: T, y: string): void;
function f<T extends number>(x: T): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "omittingSingleParameter",
				Line:      1, Column: 38, EndLine: 1, EndColumn: 47,
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
		// Named exports retain the inner declare-function range after the wrapper.
		{
			Code: `export declare function f(x: string): void;
export
declare function f(x: string): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      3,
				Column:    1,
			}},
		},
		// A directive before export does not suppress an anonymous default declaration on its function line.
		{
			Code: `declare function ExportDefaultDeclaration(x: string): void;
// eslint-disable-next-line test
export /* wrapper */ default
function (x: string): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "allParametersAreSame",
				Line:      4,
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
		// In larger groups, a single-parameter difference cites the first
		// differing parameter rather than the signature's first token.
		{
			Code: `declare function f(
  x: string): void;
declare function f(x: number): void;
declare function f(x: boolean): void;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 2 can be combined into one signature taking `string | number`.", Line: 3},
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 2 can be combined into one signature taking `string | boolean`.", Line: 4},
				{MessageId: "singleParameterDifference", Message: "This overload and the one on line 3 can be combined into one signature taking `number | boolean`.", Line: 4},
			},
		},
		// A missing peer annotation preserves the annotated type's exact text.
		{
			Code: `declare function plain(value): void;
declare function plain(value: string|number): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string|number`.",
				Line:      2,
			}},
		},
		// A sole annotation uses the same union-member formatting as paired annotations.
		{
			Code: `declare function functionType(value): void;
declare function functionType(value: () => void): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `(() => void)`.",
				Line:      2,
			}},
		},
		{
			Code: `declare function parenthesized(value): void;
declare function parenthesized(value: (string)): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "singleParameterDifference",
				Message:   "These overloads can be combined into one signature taking `string`.",
				Line:      2,
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
