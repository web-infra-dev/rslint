package consistent_type_exports

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeExportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		// Value exports
		{Code: `export { Button } from 'some-library';`},
		{Code: `export { Button as ButtonAlias } from 'some-library';`},
		{Code: `export { ButtonAlias as Button } from 'some-library';`},
		{Code: `const btn = 1; export { btn };`},

		// Already type-only exports
		{Code: `export type { Type1 } from 'some-library';`},
		{Code: `export type { Type1, Type2 } from 'some-library';`},
		{Code: `type T = string; export type { T };`},
		{Code: `export type * from 'some-library';`},
		{Code: `export type * as ns from 'some-library';`},

		// Mixed exports with inline type specifier
		{Code: `export { Value, type Type } from 'some-library';`},
		{Code: `export { type Type, Value } from 'some-library';`},
		{Code: `export { Value, type Type1, type Type2 } from 'some-library';`},

		// Default exports (not affected by this rule)
		{Code: `export default 1;`},
		{Code: `export default function foo() {}`},

		// Export declarations (not affected by this rule)
		{Code: `export const value = 1;`},
		{Code: `export function foo() {}`},
		{Code: `export class Foo {}`},
		{Code: `export type Foo = string;`},
		{Code: `export interface Foo {}`},

		// Re-exports of value
		{Code: `export * from 'some-library';`},
		{Code: `export * as ns from 'some-library';`},

		// Empty exports
		{Code: `export {};`},
		{Code: `export {} from 'some-library';`},

		// Namespace exports with values
		{Code: `namespace Foo { export const x = 1; } export { Foo };`},

		// Enum exports (values)
		{Code: `enum Foo { A, B } export { Foo };`},

		// Mixed local and re-exports
		{Code: `const value = 1; export { value };`},
		{Code: `type Type = string; export type { Type };`},
	}, []rule_tester.InvalidTestCase{
		// Single type export without type keyword
		{
			Code:   `type T = string; export { T };`,
			Output: []string{`type T = string; export type { T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code:   `interface T {} export { T };`,
			Output: []string{`interface T {} export type { T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code:   `type T = string; export { T as U };`,
			Output: []string{`type T = string; export type { T as U };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Multiple type exports without type keyword
		{
			Code:   `type T1 = string; type T2 = number; export { T1, T2 };`,
			Output: []string{`type T1 = string; type T2 = number; export type { T1, T2 };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code:   `interface T1 {} interface T2 {} export { T1, T2 };`,
			Output: []string{`interface T1 {} interface T2 {} export type { T1, T2 };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code:   `type T1 = string; interface T2 {} export { T1, T2 };`,
			Output: []string{`type T1 = string; interface T2 {} export type { T1, T2 };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Re-exports of type-only modules
		{
			Code:   `export { Type1 } from './consistent-type-exports-types-only';`,
			Output: []string{`export type { Type1 } from './consistent-type-exports-types-only';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code:   `export { Type1, Type2 } from './consistent-type-exports-types-only';`,
			Output: []string{`export type { Type1, Type2 } from './consistent-type-exports-types-only';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Export * from type-only module
		{
			Code:   `export * from './consistent-type-exports-types-only';`,
			Output: []string{`export type * from './consistent-type-exports-types-only';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Mixed exports: types and values together
		{
			Code: `type T = string; const value = 1; export { T, value };`,
			Output: []string{`type T = string; const value = 1; export type { T };
export { value };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `type T1 = string; type T2 = number; const value = 1; export { T1, T2, value };`,
			Output: []string{`type T1 = string; type T2 = number; const value = 1; export type { T1, T2 };
export { value };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
		{
			Code: `const value = 1; type T = string; export { value, T };`,
			Output: []string{`const value = 1; type T = string; export type { T };
export { value };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `const value1 = 1; const value2 = 2; type T1 = string; type T2 = number; export { value1, T1, value2, T2 };`,
			Output: []string{`const value1 = 1; const value2 = 2; type T1 = string; type T2 = number; export type { T1, T2 };
export { value1, value2 };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},

		// Mixed re-exports
		{
			Code: `export { Type1, value1 } from './consistent-type-exports-types';`,
			Output: []string{`export type { Type1 } from './consistent-type-exports-types';
export { value1 } from './consistent-type-exports-types';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `export { Type1, Type2, value1 } from './consistent-type-exports-types';`,
			Output: []string{`export type { Type1, Type2 } from './consistent-type-exports-types';
export { value1 } from './consistent-type-exports-types';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
		{
			Code: `export { Type1, value1, Type2, value2 } from './consistent-type-exports-types';`,
			Output: []string{`export type { Type1, Type2 } from './consistent-type-exports-types';
export { value1, value2 } from './consistent-type-exports-types';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},

		// With aliases
		{
			Code:   `type T = string; export { T as U };`,
			Output: []string{`type T = string; export type { T as U };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `type T1 = string; const value = 1; export { T1 as Type, value };`,
			Output: []string{`type T1 = string; const value = 1; export type { T1 as Type };
export { value };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},

		// Generic types
		{
			Code:   `type Generic<T> = T; export { Generic };`,
			Output: []string{`type Generic<T> = T; export type { Generic };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Type re-exports with values from same module
		{
			Code: `export { Type1, value1, Type2 } from './consistent-type-exports-types';`,
			Output: []string{`export type { Type1, Type2 } from './consistent-type-exports-types';
export { value1 } from './consistent-type-exports-types';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
	})
}

func TestConsistentTypeExportsExtrasWithInlineTypeSpecifier(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		// With inline type specifier option enabled
		{
			Code:    `export { Value, type Type } from 'some-library';`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{
			Code:    `export { type Type, Value } from 'some-library';`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{
			Code:    `export { Value, type Type1, type Type2 } from 'some-library';`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{
			Code:    `export type { Type } from 'some-library';`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
	}, []rule_tester.InvalidTestCase{
		// Still report errors for non-inline type exports
		{
			Code:    `type T = string; const value = 1; export { T, value };`,
			Output:  []string{`type T = string; const value = 1; export { type T, value };`},
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code:    `type T1 = string; type T2 = number; const value = 1; export { T1, T2, value };`,
			Output:  []string{`type T1 = string; type T2 = number; const value = 1; export { type T1, type T2, value };`},
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
	})
}

// Checked against typescript-eslint v8.70.1 with TypeScript 6.
func TestConsistentTypeExportsAdversarial(t *testing.T) {
	rule_tester.RunRuleTester(exportsFixtureRoot(t, "extras"), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		{Code: `const value = 1; export { type Missing, value };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `const value = 1; export { type Missing, value };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `const value = 1; export { Missing, value };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `const value = 1; export { Missing, value };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `class C {} export { type C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `class C {} export { type C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `interface T {} namespace T { export const value = 1; } export { T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `interface T {} namespace T { export const value = 1; } export { T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `export { C } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export { C } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `export {} from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export {} from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `export type { value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export type { value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
		},
		{Code: `export * from "./script";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * as ns from "./script";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * from "./ordinary";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * as ns from "./ordinary";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * from "./default-value";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * as ns from "./default-value";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * from "./unknown";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * as ns from "./unknown";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export * from "virtual-types";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
		{Code: `export type * as ns from "./types";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
		},
	}, []rule_tester.InvalidTestCase{
		{Code: `type T = string; type U = number; const v = 1; export { type U, T, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; type U = number; const v = 1; export type { T, U };
export { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 72},
			},
		},
		{Code: `type T = string; type U = number; const v = 1; export { type U, T, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; type U = number; const v = 1; export { type U, type T, v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 72},
			},
		},
		{Code: `type T = string; const v = 1; export { type v, T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`type T = string; const v = 1; export type { v, T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 52},
			},
		},
		{Code: `type T = string; const v = 1; export { type v, T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; const v = 1; export type { v, T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 52},
			},
		},
		{Code: `type T = string; type U = number; type W = boolean; const v = 1; export { T, U, W, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; type U = number; type W = boolean; const v = 1; export type { T, U, W };
export { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports T, U and W are not values and should be exported using `export type`.", Line: 1, Column: 66, EndLine: 1, EndColumn: 88},
			},
		},
		{Code: `type T = string; type U = number; type W = boolean; const v = 1; export { T, U, W, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; type U = number; type W = boolean; const v = 1; export { type T, type U, type W, v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports T, U and W are not values and should be exported using `export type`.", Line: 1, Column: 66, EndLine: 1, EndColumn: 88},
			},
		},
		{Code: `type T = string; const v = 1; export { T as "🍎", v as "value" };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; const v = 1; export type { T as "🍎" };
export { v as "value" };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 66},
			},
		},
		{Code: `type T = string; const v = 1; export { T as "🍎", v as "value" };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; const v = 1; export { type T as "🍎", v as "value" };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 66},
			},
		},
		{Code: `type T = string; const v = 1; export { T as T, v as v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; const v = 1; export type { T };
export { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 57},
			},
		},
		{Code: `type T = string; const v = 1; export { T as T, v as v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; const v = 1; export { type T as T, v as v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 57},
			},
		},
		{Code: `type \u0054 = string; const v = 1; export { \u0054, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type \u0054 = string; const v = 1; export type { T };
export { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 57},
			},
		},
		{Code: `type \u0054 = string; const v = 1; export { \u0054, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type \u0054 = string; const v = 1; export { type \u0054, v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 57},
			},
		},
		{Code: `type type = string; const v = 1; export { type as renamed, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type type = string; const v = 1; export type { type as renamed };
export { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export type is not a value and should be exported using `export type`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 64},
			},
		},
		{Code: `type type = string; const v = 1; export { type as renamed, v };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type type = string; const v = 1; export { type type as renamed, v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export type is not a value and should be exported using `export type`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 64},
			},
		},
		{Code: `type T = string; const v = 1; export /* before */ { /* lead */ T /* after */, v, };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; const v = 1; export type { T };
export /* before */ { v };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 84},
			},
		},
		{Code: `type T = string; const v = 1; export /* before */ { /* lead */ T /* after */, v, };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; const v = 1; export /* before */ { /* lead */ type T /* after */, v, };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 84},
			},
		},
		{Code: `type T = string; type U = number; export { type /* keep */ U, T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`type T = string; type U = number; export type { /* keep */ U, T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 67},
			},
		},
		{Code: `type T = string; type U = number; export { type /* keep */ U, T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; type U = number; export type { /* keep */ U, T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 67},
			},
		},
		{Code: "type T = string; type U = number; export { type\uFEFF// keep\r\nU, T };",
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{"type T = string; type U = number; export type { // keep\r\nU, T };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 35, EndLine: 2, EndColumn: 8},
			},
		},
		{Code: "type T = string; type U = number; export { type\uFEFF// keep\r\nU, T };",
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{"type T = string; type U = number; export type { // keep\r\nU, T };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 35, EndLine: 2, EndColumn: 8},
			},
		},
		{Code: `type T = string; export
/* keep */ { T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; export type
/* keep */ { T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 18, EndLine: 2, EndColumn: 18},
			},
		},
		{Code: `type T = string; export
/* keep */ { T };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output: []string{`type T = string; export type
/* keep */ { T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 18, EndLine: 2, EndColumn: 18},
			},
		},
		{Code: `type T = string; export { T, Missing };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`type T = string; export type { T, Missing };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 40},
			},
		},
		{Code: `type T = string; export { T, Missing };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`type T = string; export type { T, Missing };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 40},
			},
		},
		{Code: `import type { C } from "./values"; export { C as Alias };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import type { C } from "./values"; export type { C as Alias };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 58},
			},
		},
		{Code: `import type { C } from "./values"; export { C as Alias };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import type { C } from "./values"; export type { C as Alias };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 58},
			},
		},
		{Code: `import { type C } from "./values"; export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import { type C } from "./values"; export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
			},
		},
		{Code: `import { type C } from "./values"; export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import { type C } from "./values"; export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
			},
		},
		{Code: `import type C from "./values"; export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import type C from "./values"; export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
			},
		},
		{Code: `import type C from "./values"; export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import type C from "./values"; export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
			},
		},
		{Code: `import type * as NS from "./values"; export { NS };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import type * as NS from "./values"; export type { NS };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 52},
			},
		},
		{Code: `import type * as NS from "./values"; export { NS };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import type * as NS from "./values"; export type { NS };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 52},
			},
		},
		{Code: `import type C = require("./values"); export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import type C = require("./values"); export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 51},
			},
		},
		{Code: `import type C = require("./values"); export { C };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import type C = require("./values"); export type { C };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 51},
			},
		},
		{Code: `import type Missing from "./unknown"; export { Missing };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`import type Missing from "./unknown"; export type { Missing };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 58},
			},
		},
		{Code: `import type Missing from "./unknown"; export { Missing };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`import type Missing from "./unknown"; export type { Missing };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 58},
			},
		},
		{Code: `export { C } from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type { C } from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		{Code: `export { C } from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export type { C } from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		{Code: `export { ns } from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type { ns } from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		{Code: `export { ns } from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export type { ns } from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		{Code: `export { "🍎" as Fruit, value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`export type { "🍎" as Fruit } from './values';
export { value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export 🍎 is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
			},
		},
		{Code: `export { "🍎" as Fruit, value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type "🍎" as Fruit, value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export 🍎 is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
			},
		},
		{Code: `export { "🍎" as "T", value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`export type { "🍎" as "T" } from './values';
export { value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export 🍎 is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
			},
		},
		{Code: `export { "🍎" as "T", value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type "🍎" as "T", value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export 🍎 is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
			},
		},
		{Code: `export { T, value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`export type { T } from './values';
export { value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
			},
		},
		{Code: `export { T, value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type T, value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
			},
		},
		{Code: `export { T as "T", value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`export type { T as "T" } from './values';
export { value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
			},
		},
		{Code: `export { T as "T", value } from "./values";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type T as "T", value } from "./values";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
			},
		},
		{Code: `export * from "./types";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./types";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		{Code: `export * as ns from "./types";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./types";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		{Code: `export * from "./empty";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./empty";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			},
		},
		{Code: `export * as ns from "./empty";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./empty";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			},
		},
		{Code: `export * from "./aliases";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./aliases";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		{Code: `export * as ns from "./aliases";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./aliases";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		{Code: `export * from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		{Code: `export * as ns from "./forward";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./forward";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		{Code: `export * from "./default-type";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./default-type";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			},
		},
		{Code: `export * as ns from "./default-type";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./default-type";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			},
		},
		{Code: `export * from "./cycle-a";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * from "./cycle-a";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			},
		},
		{Code: `export * as ns from "./cycle-a";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`export type * as ns from "./cycle-a";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			},
		},
		{Code: `/* star * */ export /* star * */ * /* keep */ from "./types";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output:  []string{`/* star * */ export /* star * */ type * /* keep */ from "./types";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 62},
			},
		},
		{Code: `type T = string; const v = 1; export { T, v };
export { T as U };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Output: []string{`type T = string; const v = 1; export type { T };
export { v };
export type { T as U };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 47},
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 19},
			},
		},
		{Code: `export { T, value } from "./values" with { type: "json" };`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type T, value } from "./values" with { type: "json" };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
			},
		},
		{Code: `export { T, value } from "./quote'module";`,
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Output:  []string{`export { type T, value } from "./quote'module";`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
			},
		},
	})
}

func TestConsistentTypeExportsFixSafety(t *testing.T) {
	rule_tester.RunRuleTester(exportsFixtureRoot(t, "extras"), "tsconfig.json", t, &ConsistentTypeExportsRule, nil, []rule_tester.InvalidTestCase{
		{
			Code: `export { T, value } from "./values" with { type: "json" };`,
			Output: []string{`export type { T } from './values';
export { value } from "./values" with { type: "json" };`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleExportIsType"}},
		},
		{
			Code: `export { T, value } from "./quote'module";`,
			Output: []string{`export type { T } from "./quote'module";
export { value } from "./quote'module";`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "singleExportIsType"}},
		},
	})
}

func TestConsistentTypeExportsEditDemandAndDirectives(t *testing.T) {
	for _, tc := range []struct {
		name, code, output, exportNames string
		inline                          bool
		count                           int
	}{
		{name: "all types with comments", code: "type T = 1; type U = 2; export { T, type /* keep */ U };", output: "type T = 1; type U = 2; export type { T, /* keep */ U };", count: 1},
		{name: "separate", code: "type T = 1; const v = 1; export { T, v };", output: "type T = 1; const v = 1; export type { T };\nexport { v };", exportNames: "T", count: 1},
		{name: "inline", code: "type T = 1; const v = 1; export { T, v };", output: "type T = 1; const v = 1; export { type T, v };", exportNames: "T", inline: true, count: 1},
		{name: "star", code: "export /* keep */ * from './consistent-type-exports-types-only';", output: "export /* keep */ type * from './consistent-type-exports-types-only';", count: 1},
		{name: "namespace star", code: "export * as ns from './consistent-type-exports-types-only';", output: "export type * as ns from './consistent-type-exports-types-only';", count: 1},
		{name: "disabled line", code: "type T = 1;\n// eslint-disable-next-line @typescript-eslint/consistent-type-exports\nexport { T };", count: 0},
		{name: "disabled block", code: "/* eslint-disable @typescript-eslint/consistent-type-exports */\ntype T = 1; export { T };", count: 0},
		{name: "enabled after block", code: "type T = 1;\n/* eslint-disable @typescript-eslint/consistent-type-exports */\nexport { T };\n/* eslint-enable @typescript-eslint/consistent-type-exports */\nexport { T as U };", output: "type T = 1;\n/* eslint-disable @typescript-eslint/consistent-type-exports */\nexport { T };\n/* eslint-enable @typescript-eslint/consistent-type-exports */\nexport type { T as U };", count: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(tc.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			program := lintprogram.NewFromCompiler(raw)
			options := rule_tester.ResolveTestCaseOptions(t, &ConsistentTypeExportsRule, []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": tc.inline}})
			var identity []rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandSuggestion, rule.EditDemandAutofix, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: program, File: file.FileName(), HasTypeInfo: true,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: "@typescript-eslint/consistent-type-exports", Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return ConsistentTypeExportsRule.Run(ctx, options) }}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != tc.count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), tc.count)
				}
				for _, diagnostic := range diagnostics {
					if diagnostic.Suggestions != nil {
						t.Fatalf("demand %d: unexpected suggestions", demand)
					}
					if diagnostic.Message.Data["exportNames"] != tc.exportNames {
						t.Fatalf("unexpected message data: %v", diagnostic.Message.Data)
					}
					if (diagnostic.FixesPtr != nil) != (demand&rule.EditDemandAutofix != 0) {
						t.Fatalf("demand %d: incorrect fix materialization", demand)
					}
				}
				if demand&rule.EditDemandAutofix != 0 && tc.count != 0 {
					output, _, fixed := linter.ApplyRuleFixes(tc.code, diagnostics)
					if !fixed || output != tc.output {
						t.Fatalf("demand %d: unexpected output %q", demand, output)
					}
				}
				for i := range diagnostics {
					diagnostics[i].FixesPtr = nil
				}
				if demand == rule.EditDemandNone {
					identity = diagnostics
				} else if !reflect.DeepEqual(diagnostics, identity) {
					t.Fatalf("demand %d changed diagnostics", demand)
				}
			}
		})
	}
}

func TestConsistentTypeExportsSchema(t *testing.T) {
	for _, options := range [][]any{nil, {}, {map[string]any{}}, {map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}}, {map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}}} {
		if err := ConsistentTypeExportsRule.Schema.Validate(options); err != nil {
			t.Errorf("rejected valid options %v: %v", options, err)
		}
	}
	for _, options := range [][]any{{nil}, {"inline"}, {map[string]any{"unknown": true}}, {map[string]any{"fixMixedExportsWithInlineTypeSpecifier": "true"}}, {map[string]any{"fixMixedExportsWithInlineTypeSpecifier": nil}}, {map[string]any{}, map[string]any{}}} {
		if err := ConsistentTypeExportsRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options %v", options)
		}
	}
}

func TestConsistentTypeExportsNodeNext(t *testing.T) {
	rule_tester.RunRuleTester(exportsFixtureRoot(t, "extras"), "tsconfig.nodenext.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		// Like upstream, star exports resolve without an import/require usage mode.
		{Code: `export * from 'dual-value';`, FileName: "export.mts"},
		{Code: `export * as ns from 'dual-value';`, FileName: "export.mts"},
	}, []rule_tester.InvalidTestCase{
		{Code: `export * from 'dual-types';`, FileName: "export.mts", Output: []string{`export type * from 'dual-types';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue"}}},
		{Code: `export * as ns from 'dual-types';`, FileName: "export.mts", Output: []string{`export type * as ns from 'dual-types';`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue"}}},
	})
}

func TestConsistentTypeExportsCyclicAliases(t *testing.T) {
	rule_tester.RunRuleTester(exportsFixtureRoot(t, "extras"), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		{Code: `export { Missing } from './cyclic-named-a';`},
	}, nil)
}
