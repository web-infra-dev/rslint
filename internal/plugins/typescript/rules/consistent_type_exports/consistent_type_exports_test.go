package consistent_type_exports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeExportsRule(t *testing.T) {
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
			Code: `type T = string; export { T };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `interface T {} export { T };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `type T = string; export { T as U };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Multiple type exports without type keyword
		{
			Code: `type T1 = string; type T2 = number; export { T1, T2 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `interface T1 {} interface T2 {} export { T1, T2 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `type T1 = string; interface T2 {} export { T1, T2 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Re-exports of type-only modules
		{
			Code: `export { Type1 } from './consistent-type-exports-types-only';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `export { Type1, Type2 } from './consistent-type-exports-types-only';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Export * from type-only module
		{
			Code: `export * from './consistent-type-exports-types-only';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Mixed exports: types and values together
		{
			Code: `type T = string; const value = 1; export { T, value };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `type T1 = string; type T2 = number; const value = 1; export { T1, T2, value };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
		{
			Code: `const value = 1; type T = string; export { value, T };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `const value1 = 1; const value2 = 2; type T1 = string; type T2 = number; export { value1, T1, value2, T2 };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},

		// Mixed re-exports
		{
			Code: `export { Type1, value1 } from './consistent-type-exports-types';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `export { Type1, Type2, value1 } from './consistent-type-exports-types';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
		{
			Code: `export { Type1, value1, Type2, value2 } from './consistent-type-exports-types';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},

		// With aliases
		{
			Code: `type T = string; export { T as U };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},
		{
			Code: `type T1 = string; const value = 1; export { T1 as Type, value };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},

		// Generic types
		{
			Code: `type Generic<T> = T; export { Generic };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "typeOverValue"},
			},
		},

		// Type re-exports with values from same module
		{
			Code: `export { Type1, value1, Type2 } from './consistent-type-exports-types';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
	})
}

func TestConsistentTypeExportsRuleWithInlineTypeSpecifier(t *testing.T) {
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
			Code: `type T = string; const value = 1; export { T, value };`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "singleExportIsType"},
			},
		},
		{
			Code: `type T1 = string; type T2 = number; const value = 1; export { T1, T2, value };`,
			Options: []interface{}{map[string]interface{}{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "multipleExportsAreTypes"},
			},
		},
	})
}
