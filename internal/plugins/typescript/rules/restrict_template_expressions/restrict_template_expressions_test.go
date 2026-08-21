package restrict_template_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRestrictTemplateExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RestrictTemplateExpressionsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "const msg = `arg = ${'foo'}`;\n",
			},
			{
				Code: "const arg = 'foo';\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: number;\nconst msg = `arg = ${String(arg)}`;\n",
			},
			{
				Code: "declare const arg: number;\nconst msg = `arg = ${arg.toString()}`;\n",
			},
			{
				Code: "declare const arg: number;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: bigint;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: boolean;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: null;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: undefined;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: any;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: RegExp;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "const msg = `arg = ${/regex/}`;\n",
			},
			{
				Code: "declare const arg: Error;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: URL;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: URLSearchParams;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "class CustomError extends Error {}\ndeclare const arg: CustomError;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "class CustomError extends Error {}\nclass Deeper extends CustomError {}\ndeclare const arg: Deeper;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "interface Extra { extra: true }\ndeclare const Mixed: new () => Error & Extra;\nclass Derived extends Mixed {}\ndeclare const arg: Derived;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: object;\ndeclare function tag(strings: TemplateStringsArray, ...values: unknown[]): string;\nconst msg = tag`arg = ${arg}`;\n",
			},
			{
				Code: "const msg = `arg`;\n",
			},
			{
				Code: "declare const arg: string | number;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: string & { __brand: 'x' };\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "function f<T extends string>(arg: T) {\n  return `arg = ${arg}`;\n}\n",
			},
			{
				Code: "enum E {\n  A = 'a',\n}\ndeclare const arg: E;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "enum E {\n  A,\n}\ndeclare const arg: E;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: `a${number}`;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code:    "declare const arg: string[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: string[][];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: [string, number];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: never;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowNever": true},
			},
			{
				Code:    "declare const arg: string;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowAny": false, "allowBoolean": false, "allowNever": false, "allowNullish": false, "allowNumber": false, "allowRegExp": false},
			},
			{
				Code:    "interface Foo {\n  a: string;\n}\ndeclare const arg: Foo;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{map[string]any{"from": "file", "name": "Foo"}}},
			},
			{
				Code:    "interface Foo {\n  a: string;\n}\ndeclare const arg: Foo;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{"Foo"}},
			},
			{
				Code:    "interface Foo {\n  a: string;\n}\ninterface Bar extends Foo {\n  b: string;\n}\ndeclare const arg: Bar;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{map[string]any{"from": "file", "name": "Foo"}}},
			},
			{
				Code:    "declare const arg: number;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{}, "allowNumber": true},
			},
			{
				Code:    "declare const arg: Error;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowNumber": false},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "declare const arg: object;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"object\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "const arg = {};\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"{}\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: string[];\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"string[]\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: [string, string];\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"[string, string]\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: never;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"never\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: unknown;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"unknown\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: symbol;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"symbol\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: () => void;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"() => void\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: Map<string, string>;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"Map<string, string>\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const a: object;\ndeclare const b: string;\ndeclare const c: symbol;\nconst msg = `${a} ${b} ${c}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"object\" of template literal expression.",
						Line:      4,
						Column:    16,
						EndLine:   4,
						EndColumn: 17,
					},
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"symbol\" of template literal expression.",
						Line:      4,
						Column:    26,
						EndLine:   4,
						EndColumn: 27,
					},
				},
			},
			{
				Code: "declare const arg: object;\nconst msg = `outer ${`inner ${arg}`}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"object\" of template literal expression.",
						Line:      2,
						Column:    31,
						EndLine:   2,
						EndColumn: 34,
					},
				},
			},
			{
				Code: "declare const arg: string | object;\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"string | object\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "declare const arg: { a: 1 } & { b: 2 };\nconst msg = `arg = ${arg}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"{ a: 1; } & { b: 2; }\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "function f<T>(arg: T) {\n  return `arg = ${arg}`;\n}\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"T\" of template literal expression.",
						Line:      2,
						Column:    19,
						EndLine:   2,
						EndColumn: 22,
					},
				},
			},
			{
				Code:    "declare const arg: number;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowNumber": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"number\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "interface Demo { value: string }\ndeclare const arg: Demo;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{map[string]any{"from": "file", "name": "Demo", "path": ""}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"Demo\" of template literal expression.",
						Line:      3,
						Column:    22,
						EndLine:   3,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: bigint;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowNumber": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"bigint\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: boolean;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowBoolean": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"boolean\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: null | undefined;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowNullish": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"null | undefined\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: any;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowAny": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"any\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: RegExp;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowRegExp": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"RegExp\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: number[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true, "allowNumber": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"number[]\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: object[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"object[]\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: [];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"[]\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: number;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowAny": false, "allowBoolean": false, "allowNever": false, "allowNullish": false, "allowNumber": false, "allowRegExp": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"number\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    "declare const arg: Error;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidType",
						Message:   "Invalid type \"Error\" of template literal expression.",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 25,
					},
				},
			},
		},
	)
}

func TestRestrictTemplateExpressionsEdgeCases(t *testing.T) {
	invalidType := func(message string) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{
			MessageId: "invalidType",
			Message:   message,
		}}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RestrictTemplateExpressionsRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "declare const arg: readonly string[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "const arg = ['foo', 1] as const;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: [string, number?];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: [string, ...number[]];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "const msg = `arg = ${[, 2]}`;\n",
				Options: map[string]any{"allowArray": true},
			},
			{
				Code:    "declare const arg: never[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true, "allowNever": true},
			},
			{
				Code: "type Recursive = Recursive[];\ndeclare const arg: Recursive;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{
					"allowArray": true,
					"allow":      []any{"Recursive"},
				},
			},
			{
				Code: "class Base {}\nclass Middle extends Base {}\nclass Leaf extends Middle {}\ndeclare const arg: Leaf;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{
					map[string]any{"from": "file", "name": "Base"},
				}},
			},
			{
				Code: "interface A { a: 1 }\ninterface B { b: 1 }\ninterface C extends A, B { c: 1 }\ndeclare const arg: C;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{
					map[string]any{"from": "file", "name": "B"},
				}},
			},
			{
				Code: "class Base<T> { value!: T }\nclass Leaf extends Base<string> {}\ndeclare const arg: Leaf;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{
					map[string]any{"from": "file", "name": "Base"},
				}},
			},
			{
				Code: "type Custom = { value: string };\ndeclare const arg: Custom;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allow": []any{
					map[string]any{"from": "file", "name": "Custom"},
				}},
			},
			{
				Code: "class Child extends URL {}\ndeclare const arg: Child;\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: number & { readonly brand: unique symbol };\nconst msg = `arg = ${arg}`;\n",
			},
			{
				Code: "declare const arg: object;\ndeclare function tag(strings: TemplateStringsArray, ...values: unknown[]): string;\nconst msg = tag`arg = ${arg}`;\n",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "declare const arg: [string, number?];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true, "allowNullish": false},
				Errors:  invalidType("Invalid type \"[string, (number | undefined)?]\" of template literal expression."),
			},
			{
				Code:    "declare const arg: (string | object)[][];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors:  invalidType("Invalid type \"(string | object)[][]\" of template literal expression."),
			},
			{
				Code:    "const msg = `arg = ${[, 2]}`;\n",
				Options: map[string]any{"allowArray": true, "allowNullish": false},
				Errors:  invalidType("Invalid type \"(number | undefined)[]\" of template literal expression."),
			},
			{
				Code:    "declare const arg: never[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors:  invalidType("Invalid type \"never[]\" of template literal expression."),
			},
			{
				Code:    "declare const arg: any[];\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true, "allowAny": false},
				Errors:  invalidType("Invalid type \"any[]\" of template literal expression."),
			},
			{
				Code:    "type Recursive = Recursive[];\ndeclare const arg: Recursive;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors:  invalidType("Invalid type \"Recursive\" of template literal expression."),
			},
			{
				Code:    "type Left = Right[];\ntype Right = Left[];\ndeclare const arg: Left;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors:  invalidType("Invalid type \"Left\" of template literal expression."),
			},
			{
				Code:    "class CustomArray extends Array<string> {}\ndeclare const arg: CustomArray;\nconst msg = `arg = ${arg}`;\n",
				Options: map[string]any{"allowArray": true},
				Errors:  invalidType("Invalid type \"CustomArray\" of template literal expression."),
			},
			{
				Code:   "class Child extends RegExp {}\ndeclare const arg: Child;\nconst msg = `arg = ${arg}`;\n",
				Errors: invalidType("Invalid type \"Child\" of template literal expression."),
			},
			{
				Code:   "const arg = new String('value');\nconst msg = `arg = ${arg}`;\n",
				Errors: invalidType("Invalid type \"String\" of template literal expression."),
			},
			{
				Code:   "declare const arg: unknown;\nconst msg = `arg = ${arg as object}`;\n",
				Errors: invalidType("Invalid type \"object\" of template literal expression."),
			},
			{
				Code:   "declare const arg: { a: 1 };\nconst msg = `arg = ${arg satisfies object}`;\n",
				Errors: invalidType("Invalid type \"{ a: 1; }\" of template literal expression."),
			},
			{
				Code:   "declare const arg: object;\nconst msg = `arg = ${(arg)}`;\n",
				Errors: invalidType("Invalid type \"object\" of template literal expression."),
			},
			{
				Code: "declare const a: object;\ndeclare const b: string;\ndeclare const c: symbol;\nconst msg = `${a} ${b} ${c}`;\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidType", Message: "Invalid type \"object\" of template literal expression."},
					{MessageId: "invalidType", Message: "Invalid type \"symbol\" of template literal expression."},
				},
			},
		},
	)
}
