package consistent_type_exports

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

func exportsFixtureRoot(t *testing.T, archive string) rule_tester.Root {
	t.Helper()
	return rule_tester.Root{
		Dir: tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/"+archive+".txtar").Materialize(t, "")),
		FS:  bundled.WrapFS(osvfs.FS()),
	}
}

// Ported from typescript-eslint v8.70.1: all 24 valid and 25 invalid cases.
func TestConsistentTypeExportsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(exportsFixtureRoot(t, "upstream"), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		{Code: `export { Foo } from 'foo';`},
		{Code: `export type { Type1 } from './consistent-type-exports';`},
		{Code: `export { value1 } from './consistent-type-exports';`},
		{Code: `export { value1 as "🍎" } from './consistent-type-exports';`},
		{Code: `export type { value1 } from './consistent-type-exports';`},
		{Code: `
const variable = 1;
class Class {}
enum Enum {}
function Func() {}
namespace ValueNS {
  export const x = 1;
}

export { variable, Class, Enum, Func, ValueNS };
    `},
		{Code: `
type Alias = 1;
interface IFace {}
namespace TypeNS {
  export type x = 1;
}

export type { Alias, IFace, TypeNS };
    `},
		{Code: `
const foo = 1;
export type { foo };
    `},
		{Code: `
namespace NonTypeNS {
  export const x = 1;
}

export { NonTypeNS };
    `},
		{Code: `export * from './unknown-module';`},
		{Code: `export * from './consistent-type-exports';`},
		{Code: `export type * from './consistent-type-exports/type-only-exports';`},
		{Code: `export type * from './consistent-type-exports/type-only-reexport';`},
		{Code: `export * from './consistent-type-exports/value-reexport';`},
		{Code: `export * as foo from './consistent-type-exports';`},
		{Code: `export type * as foo from './consistent-type-exports/type-only-exports';`},
		{Code: `export type * as foo from './consistent-type-exports/type-only-reexport';`},
		{Code: `export * as foo from './consistent-type-exports/value-reexport';`},
		{Code: `
import * as Foo from './consistent-type-exports';
type Foo = 1;
export { Foo }
    `},
		{Code: `
import { Type1 } from './consistent-type-exports';
const Type1 = 1;
export { Type1 };
    `},
		{Code: `
export { A } from './consistent-type-exports/reexport-2-named';
    `},
		{Code: `
import { A } from './consistent-type-exports/reexport-2-named';
export { A };
    `},
		{Code: `
export { A } from './consistent-type-exports/reexport-2-namespace';
    `},
		{Code: `
import { A } from './consistent-type-exports/reexport-2-namespace';
export { A };
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `export { Type1 } from './consistent-type-exports';`,
			Output: []string{`export type { Type1 } from './consistent-type-exports';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 51},
			},
		},
		{
			Code:   `export { Type1 as "🍎" } from './consistent-type-exports';`,
			Output: []string{`export type { Type1 as "🍎" } from './consistent-type-exports';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
			},
		},
		{
			Code: `export { Type1, value1 } from './consistent-type-exports';`,
			Output: []string{`export type { Type1 } from './consistent-type-exports';
export { value1 } from './consistent-type-exports';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export Type1 is not a value and should be exported using `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 59},
			},
		},
		{
			Code: `
export { Type1, value1, value2 } from './consistent-type-exports';
      `,
			Output: []string{`
export type { Type1 } from './consistent-type-exports';
export { value1, value2 } from './consistent-type-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export Type1 is not a value and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 67},
			},
		},
		{
			Code: `
export { Type1, value1, Type2, value2 } from './consistent-type-exports';
      `,
			Output: []string{`
export type { Type1, Type2 } from './consistent-type-exports';
export { value1, value2 } from './consistent-type-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports Type1 and Type2 are not values and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 74},
			},
		},
		{
			Code:   `export { Type2 as Foo } from './consistent-type-exports';`,
			Output: []string{`export type { Type2 as Foo } from './consistent-type-exports';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
			},
		},
		{
			Code: `
export { Type2 as Foo, value1 } from './consistent-type-exports';
      `,
			Output: []string{`
export type { Type2 as Foo } from './consistent-type-exports';
export { value1 } from './consistent-type-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export Type2 is not a value and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 66},
			},
		},
		{
			Code: `
export {
  Type2 as Foo,
  value1 as BScope,
  value2 as CScope,
} from './consistent-type-exports';
      `,
			Output: []string{`
export type { Type2 as Foo } from './consistent-type-exports';
export { value1 as BScope, value2 as CScope } from './consistent-type-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export Type2 is not a value and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 6, EndColumn: 36},
			},
		},
		{
			Code: `
import { Type2 } from './consistent-type-exports';
export { Type2 };
      `,
			Output: []string{`
import { Type2 } from './consistent-type-exports';
export type { Type2 };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 18},
			},
		},
		{
			Code: `
import { value2, Type2 } from './consistent-type-exports';
export { value2, Type2 };
      `,
			Output: []string{`
import { value2, Type2 } from './consistent-type-exports';
export type { Type2 };
export { value2 };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export Type2 is not a value and should be exported using `export type`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 26},
			},
		},
		{
			Code: `
type Alias = 1;
interface IFace {}
namespace TypeNS {
  export type x = 1;
  export const f = 1;
}

export { Alias, IFace, TypeNS };
      `,
			Output: []string{`
type Alias = 1;
interface IFace {}
namespace TypeNS {
  export type x = 1;
  export const f = 1;
}

export type { Alias, IFace };
export { TypeNS };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports Alias and IFace are not values and should be exported using `export type`.", Line: 9, Column: 1, EndLine: 9, EndColumn: 33},
			},
		},
		{
			Code: `
namespace TypeNS {
  export interface Foo {}
}

export { TypeNS };
      `,
			Output: []string{`
namespace TypeNS {
  export interface Foo {}
}

export type { TypeNS };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 6, Column: 1, EndLine: 6, EndColumn: 19},
			},
		},
		{
			Code: `
type T = 1;
export { type T, T };
      `,
			Output: []string{`
type T = 1;
export type { T, T };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 22},
			},
		},
		{
			Code: `
type T = 1;
export { type/* */T, type     /* */T, T };
      `,
			Output: []string{`
type T = 1;
export type { /* */T, /* */T, T };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 43},
			},
		},
		{
			Code: `
type T = 1;
const x = 1;
export { type T, T, x };
      `,
			Output: []string{`
type T = 1;
const x = 1;
export type { T, T };
export { x };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 25},
			},
		},
		{
			Code: `
type T = 1;
const x = 1;
export { T, x };
      `,
			Output: []string{`
type T = 1;
const x = 1;
export { type T, x };
      `},
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `singleExportIsType`, Message: "Type export T is not a value and should be exported using `export type`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 17},
			},
		},
		{
			Code: `
type T = 1;
export { type T, T };
      `,
			Output: []string{`
type T = 1;
export type { T, T };
      `},
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 22},
			},
		},
		{
			Code: `
export {
  Type1,
  Type2 as Foo,
  type value1 as BScope,
  value2 as CScope,
} from './consistent-type-exports';
      `,
			Output: []string{`
export type { Type1, Type2 as Foo, value1 as BScope } from './consistent-type-exports';
export { value2 as CScope } from './consistent-type-exports';
      `},
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": false}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports Type1 and Type2 are not values and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 7, EndColumn: 36},
			},
		},
		{
			Code: `
export {
  Type1,
  Type2 as Foo,
  type value1 as BScope,
  value2 as CScope,
} from './consistent-type-exports';
      `,
			Output: []string{`
export {
  type Type1,
  type Type2 as Foo,
  type value1 as BScope,
  value2 as CScope,
} from './consistent-type-exports';
      `},
			Options: []any{map[string]any{"fixMixedExportsWithInlineTypeSpecifier": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `multipleExportsAreTypes`, Message: "Type exports Type1 and Type2 are not values and should be exported using `export type`.", Line: 2, Column: 1, EndLine: 7, EndColumn: 36},
			},
		},
		{
			Code: `
        export * from './consistent-type-exports/type-only-exports';
      `,
			Output: []string{`
        export type * from './consistent-type-exports/type-only-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 69},
			},
		},
		{
			Code: `
        /* comment 1 */ export
          /* comment 2 */ *
            // comment 3
            from './consistent-type-exports/type-only-exports';
      `,
			Output: []string{`
        /* comment 1 */ export
          /* comment 2 */ type *
            // comment 3
            from './consistent-type-exports/type-only-exports';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 2, Column: 25, EndLine: 5, EndColumn: 64},
			},
		},
		{
			Code: `
        export * from './consistent-type-exports/type-only-reexport';
      `,
			Output: []string{`
        export type * from './consistent-type-exports/type-only-reexport';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 70},
			},
		},
		{
			Code: `
        export * as foo from './consistent-type-exports/type-only-reexport';
      `,
			Output: []string{`
        export type * as foo from './consistent-type-exports/type-only-reexport';
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 77},
			},
		},
		{
			Code: `
        import type * as Foo from './consistent-type-exports';
        type Foo = 1;
        export { Foo };
      `,
			Output: []string{`
        import type * as Foo from './consistent-type-exports';
        type Foo = 1;
        export type { Foo };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 4, Column: 9, EndLine: 4, EndColumn: 24},
			},
		},
		{
			Code: `
        import { type NAME as Foo } from './consistent-type-exports';
        export { Foo };
      `,
			Output: []string{`
        import { type NAME as Foo } from './consistent-type-exports';
        export type { Foo };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: `typeOverValue`, Message: "All exports in the declaration are only used as types. Use `export type`.", Line: 3, Column: 9, EndLine: 3, EndColumn: 24},
			},
		},
	})
}
