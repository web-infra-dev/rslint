package prefer_literal_enum_member

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferLiteralEnumMemberRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferLiteralEnumMemberRule, []rule_tester.ValidTestCase{
		{Code: `
enum ValidRegex {
  A = /test/,
}
`},
		{Code: `
enum ValidString {
  A = 'test',
}
`},
		{Code: "enum ValidLiteral {\n  A = `test`,\n}"},
		{Code: `
enum ValidNumber {
  A = 42,
}
`},
		{Code: `
enum ValidNumber {
  A = -42,
}
`},
		{Code: `
enum ValidNumber {
  A = +42,
}
`},
		{Code: `
enum ValidNull {
  A = null,
}
`},
		{Code: `
enum ValidPlain {
  A,
}
`},
		{Code: `
enum ValidQuotedKey {
  'a',
}
`},
		{Code: `
enum ValidQuotedKeyWithAssignment {
  'a' = 1,
}
`},
		{Code: `
enum ValidKeyWithComputedSyntaxButNoComputedKey {
  ['a'],
}
`},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 >> 0,
  C = 1 >>> 0,
  D = 1 | 0,
  E = 1 & 0,
  F = 1 ^ 0,
  G = ~1,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 >> 0,
  C = A | B,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 >> 0,
  C = Foo.A | Foo.B,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 >> 0,
  C = Foo['A'] | B,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  ['A-1'] = 1 << 0,
  C = ~Foo['A-1'],
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 << 1,
  C = 1 << 2,
  D = A | B | C,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 << 1,
  C = 1 << 2,
  D = Foo.A | Foo.B | Foo.C,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 << 1,
  C = 1 << 2,
  D = Foo.A | (Foo.B & ~Foo.C),
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 << 1,
  C = 1 << 2,
  D = Foo.A | -Foo.B,
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
enum InvalidObject {
  A = {},
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum InvalidArray {
  A = [],
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: "enum InvalidTemplateLiteral {\n  A = `foo ${0}`,\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      2,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum InvalidConstructor {
  A = new Set(),
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum InvalidExpression {
  A = 2 + 2,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum InvalidExpression {
  A = delete 2,
  B = -a,
  C = void 2,
  D = ~2,
  E = !0,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      3,
					Column:    3,
				},
				{
					MessageId: "notLiteral",
					Line:      4,
					Column:    3,
				},
				{
					MessageId: "notLiteral",
					Line:      5,
					Column:    3,
				},
				{
					MessageId: "notLiteral",
					Line:      6,
					Column:    3,
				},
				{
					MessageId: "notLiteral",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
const variable = 'Test';
enum InvalidVariable {
  A = 'TestStr',
  B = 2,
  C,
  V = variable,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum InvalidEnumMember {
  A = 'TestStr',
  B = A,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
const Valid = { A: 2 };
enum InvalidObjectMember {
  A = 'TestStr',
  B = Valid.A,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum Valid {
  A,
}
enum InvalidEnumMember {
  A = 'TestStr',
  B = Valid.A,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      7,
					Column:    3,
				},
			},
		},
		{
			Code: `
const obj = { a: 1 };
enum InvalidSpread {
  A = 'TestStr',
  B = { ...a },
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notLiteral",
					Line:      5,
					Column:    3,
				},
			},
		},
		{
			Code: `
enum Foo {
  A = 1 << 0,
  B = 1 >> 0,
  C = 1 >>> 0,
  D = 1 | 0,
  E = 1 & 0,
  F = 1 ^ 0,
  G = ~1,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notLiteral", Line: 3, Column: 3},
				{MessageId: "notLiteral", Line: 4, Column: 3},
				{MessageId: "notLiteral", Line: 5, Column: 3},
				{MessageId: "notLiteral", Line: 6, Column: 3},
				{MessageId: "notLiteral", Line: 7, Column: 3},
				{MessageId: "notLiteral", Line: 8, Column: 3},
				{MessageId: "notLiteral", Line: 9, Column: 3},
			},
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": false,
				},
			},
		},
		{
			Code: `
const x = 1;
enum Foo {
  A = x << 0,
  B = x >> 0,
  C = x >>> 0,
  D = x | 0,
  E = x & 0,
  F = x ^ 0,
  G = ~x,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notLiteralOrBitwiseExpression", Line: 4, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 5, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 6, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 7, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 8, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 9, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 10, Column: 3},
			},
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
const x = 1;
enum Foo {
  A = 1 << 0,
  B = x >> Foo.A,
  C = x >> A,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notLiteralOrBitwiseExpression", Line: 5, Column: 3},
				{MessageId: "notLiteralOrBitwiseExpression", Line: 6, Column: 3},
			},
			Options: []interface{}{
				map[string]interface{}{
					"allowBitwiseExpressions": true,
				},
			},
		},
		{
			Code: `
enum Foo {
  A,
  B = +A,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notLiteral", Line: 4},
			},
		},
	})
}
