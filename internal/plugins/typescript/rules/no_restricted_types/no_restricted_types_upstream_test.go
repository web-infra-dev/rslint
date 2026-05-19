// TestNoRestrictedTypesUpstream migrates the full valid/invalid suite from
// upstream
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-restricted-types.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in no_restricted_types_extras_test.go.
package no_restricted_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// optionsArr wraps a single options object the way upstream's
// `options: [{...}]` ships it across the JS bridge — exercising the
// `GetOptionsMap` JSON path rather than a typed-struct shortcut.
func optionsArr(o map[string]interface{}) []interface{} {
	return []interface{}{o}
}

func TestNoRestrictedTypesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedTypesRule, []rule_tester.ValidTestCase{
		{Code: `let f = Object();`},
		{Code: `let f: { x: number; y: number } = { x: 1, y: 1 };`},
		{
			Code:    `let f = Object();`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Object": true}}),
		},
		{
			Code:    `let f = Object(false);`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Object": true}}),
		},
		{
			Code:    `let g = Object.create(null);`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Object": true}}),
		},
		{
			Code:    `let e: namespace.Object;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Object": true}}),
		},
		{
			Code:    `let value: _.NS.Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"NS.Banned": true}}),
		},
		{
			Code:    `let value: NS.Banned._;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"NS.Banned": true}}),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `let value: bigint;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"bigint": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `bigint` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: boolean;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"boolean": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `boolean` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: never;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"never": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `never` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: null;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"null": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `null` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: number;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"number": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `number` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: object;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"object": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `object` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: string;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"string": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `string` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: symbol;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"symbol": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `symbol` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: undefined;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"undefined": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `undefined` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: unknown;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"unknown": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `unknown` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: void;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"void": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `void` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: [];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"[]": "Use unknown[] instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `[]` as a type. Use unknown[] instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: [  ];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"[]": "Use unknown[] instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `[]` as a type. Use unknown[] instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: [[]];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"[]": "Use unknown[] instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `[]` as a type. Use unknown[] instead.", Line: 1, Column: 13},
			},
		},
		{
			Code:    `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use '{}' instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use '{}' instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: Banned[];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use '{}' instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use '{}' instead.", Line: 1, Column: 12},
			},
		},
		{
			Code:    `let value: [Banned];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use '{}' instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use '{}' instead.", Line: 1, Column: 13},
			},
		},
		{
			// Upstream's getCustomMessage short-circuits on !bannedType, so
			// `Banned: ''` collapses to a bare ban with no trailing space.
			Code:    `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": ""}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
			},
		},
		{
			Code: `let b: { c: Banned };`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`let b: { c: Ok };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 13},
			},
		},
		{
			Code: `1 as Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`1 as Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 6},
			},
		},
		{
			Code: `class Derived implements Banned {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`class Derived implements Ok {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 26},
			},
		},
		{
			Code: `class Derived implements Banned1, Banned2 {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned1": map[string]interface{}{"fixWith": "Ok1", "message": "Use Ok1 instead."},
				"Banned2": map[string]interface{}{"fixWith": "Ok2", "message": "Use Ok2 instead."},
			}}),
			Output: []string{`class Derived implements Ok1, Ok2 {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned1` as a type. Use Ok1 instead.", Line: 1, Column: 26},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned2` as a type. Use Ok2 instead.", Line: 1, Column: 35},
			},
		},
		{
			Code: `class Derived implements Omit<Foo, 'a'> {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Omit": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`class Derived implements Ok<Foo, 'a'> {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Omit` as a type. Use Ok instead.", Line: 1, Column: 26},
			},
		},
		{
			Code: `interface Derived extends Banned {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`interface Derived extends Ok {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 27},
			},
		},
		{
			Code: `interface Derived extends Omit<Foo, 'a'> {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Omit": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`interface Derived extends Ok<Foo, 'a'> {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Omit` as a type. Use Ok instead.", Line: 1, Column: 27},
			},
		},
		{
			Code: `type Intersection = Banned & {};`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`type Intersection = Ok & {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 21},
			},
		},
		{
			Code: `type Union = Banned | {};`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`type Union = Ok | {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 14},
			},
		},
		{
			Code: `let value: NS.Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"NS.Banned": map[string]interface{}{"fixWith": "NS.Ok", "message": "Use NS.Ok instead."},
			}}),
			Output: []string{`let value: NS.Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `NS.Banned` as a type. Use NS.Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: {} = {};`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"{}": map[string]interface{}{"fixWith": "object", "message": "Use object instead."},
			}}),
			Output: []string{`let value: object = {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `{}` as a type. Use object instead.", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: NS.Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"  NS.Banned  ": map[string]interface{}{"fixWith": "NS.Ok", "message": "Use NS.Ok instead."},
			}}),
			Output: []string{`let value: NS.Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `NS.Banned` as a type. Use NS.Ok instead.", Line: 1, Column: 12},
			},
		},
		{
			Code: `let value: Type<   Banned   >;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"       Banned      ": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok instead."},
			}}),
			Output: []string{`let value: Type<   Ok   >;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 20},
			},
		},
		{
			Code: `type Intersection = Banned<any>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned<any>": "Don't use `any` as a type parameter to `Banned`",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<any>` as a type. Don't use `any` as a type parameter to `Banned`", Line: 1, Column: 21},
			},
		},
		{
			Code: `type Intersection = Banned<A,B>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned<A, B>": "Don't pass `A, B` as parameters to `Banned`",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<A,B>` as a type. Don't pass `A, B` as parameters to `Banned`", Line: 1, Column: 21},
			},
		},
	})
}
