package no_magic_numbers

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoMagicNumbersRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Valid: basic non-magic numbers
		{
			Code: `const FOO = -1;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-1},
			},
		},
		{
			Code: `type Foo = 'bar';`,
		},
		{
			Code: `type Foo = true;`,
		},
		{
			Code: `type Foo = 1;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": true,
			},
		},
		{
			Code: `type Foo = -1;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": true,
			},
		},
		{
			Code: `type Foo = 1 | 2 | 3;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": true,
			},
		},
		{
			Code: `type Foo = 1 | -1;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": true,
			},
		},

		// Valid: ignoreEnums
		{
			Code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +1,
}`,
			Options: map[string]interface{}{
				"ignoreEnums": true,
			},
		},

		// Valid: ignoreReadonlyClassProperties
		{
			Code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 1;
  static readonly D = 1;
  readonly E = -1;
  readonly F = +1;
  private readonly G = 100n;
}`,
			Options: map[string]interface{}{
				"ignoreReadonlyClassProperties": true,
			},
		},

		// Valid: ignoreTypeIndexes
		{
			Code: `type Foo = Bar[0];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[-1];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[0xab];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[5.6e1];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[10n];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[1 | -2];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[1 & -2];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[1 & number];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar[((1 & -2) | 3) | 4];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Parameters<Bar>[2];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar['baz'];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": true,
			},
		},
		{
			Code: `type Foo = Bar['baz'];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": false,
			},
		},

		// Valid: ignore specific values
		{
			Code: `type Foo = 1;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{1},
			},
		},
		{
			Code: `type Foo = -2;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-2},
			},
		},
		{
			Code: `type Foo = 3n;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{"3n"},
			},
		},
		{
			Code: `type Foo = -4n;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{"-4n"},
			},
		},
		{
			Code: `type Foo = 5.6;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{5.6},
			},
		},
		{
			Code: `type Foo = -7.8;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-7.8},
			},
		},
		{
			Code: `type Foo = 0x0a;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{0x0a},
			},
		},
		{
			Code: `type Foo = -0xbc;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-0xbc},
			},
		},
		{
			Code: `type Foo = 1e2;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{1e2},
			},
		},
		{
			Code: `type Foo = -3e4;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-3e4},
			},
		},
		{
			Code: `type Foo = 5e-6;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{5e-6},
			},
		},
		{
			Code: `type Foo = -7e-8;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-7e-8},
			},
		},

	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// Invalid: ignoreNumericLiteralTypes = false
		{
			Code: `type Foo = 1;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `type Foo = -1;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `type Foo = 1 | 2 | 3;`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    16,
				},
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    20,
				},
			},
		},

		// Invalid: interface properties (not ignored by ignoreNumericLiteralTypes)
		{
			Code: `
interface Foo {
  bar: 1;
}`,
			Options: map[string]interface{}{
				"ignoreNumericLiteralTypes": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    8,
				},
			},
		},

		// Invalid: ignoreEnums = false
		{
			Code: `
enum foo {
  SECOND = 1000,
  NUM = '0123456789',
  NEG = -1,
  POS = +1,
}`,
			Options: map[string]interface{}{
				"ignoreEnums": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    12,
				},
				{
					MessageId: "noMagic",
					Line:      5,
					Column:    9,
				},
				{
					MessageId: "noMagic",
					Line:      6,
					Column:    10,
				},
			},
		},

		// Invalid: ignoreReadonlyClassProperties = false
		{
			Code: `
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 3;
  static readonly D = 4;
  readonly E = -5;
  readonly F = +6;
  private readonly G = 100n;
}`,
			Options: map[string]interface{}{
				"ignoreReadonlyClassProperties": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    16,
				},
				{
					MessageId: "noMagic",
					Line:      4,
					Column:    16,
				},
				{
					MessageId: "noMagic",
					Line:      5,
					Column:    30,
				},
				{
					MessageId: "noMagic",
					Line:      6,
					Column:    23,
				},
				{
					MessageId: "noMagic",
					Line:      7,
					Column:    16,
				},
				{
					MessageId: "noMagic",
					Line:      8,
					Column:    17,
				},
				{
					MessageId: "noMagic",
					Line:      9,
					Column:    24,
				},
			},
		},

		// Invalid: ignoreTypeIndexes = false
		{
			Code: `type Foo = Bar[0];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			Code: `type Foo = Bar[-1];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			Code: `type Foo = Bar[0xab];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			Code: `type Foo = Bar[10n];`,
			Options: map[string]interface{}{
				"ignoreTypeIndexes": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    16,
				},
			},
		},

		// Invalid: ignore value mismatches
		{
			Code: `type Foo = 1;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{-1},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `type Foo = -2;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{2},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `type Foo = 3n;`,
			Options: map[string]interface{}{
				"ignore": []interface{}{"-3n"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    12,
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, validTestCases, invalidTestCases)
}