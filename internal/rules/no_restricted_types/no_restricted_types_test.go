package no_restricted_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoRestrictedTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedTypesRule, []rule_tester.ValidTestCase{
		// Valid cases
		{
			Code: `let f = Object();`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Object": true,
				},
			},
		},
		{
			Code: `let f: { x: number; y: number } = { x: 1, y: 1 };`,
		},
		{
			Code: `let f = Object(false);`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Object": true,
				},
			},
		},
		{
			Code: `let g = Object.create(null);`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Object": true,
				},
			},
		},
		{
			Code: `let e: namespace.Object;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Object": true,
				},
			},
		},
		{
			Code: `let value: _.NS.Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"NS.Banned": true,
				},
			},
		},
		{
			Code: `let value: NS.Banned._;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"NS.Banned": true,
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - keyword types
		{
			Code: `let value: bigint;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"bigint": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: boolean;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"boolean": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: never;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"never": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: null;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"null": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: number;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"number": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: object;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"object": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: string;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"string": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: symbol;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"symbol": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: undefined;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"undefined": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: unknown;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"unknown": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: void;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"void": "Use Ok instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},

		// Invalid cases - special types
		{
			Code: `let value: [];`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"[]": "Use unknown[] instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: [  ];`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"[]": "Use unknown[] instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: [[]];`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"[]": "Use unknown[] instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Invalid cases - type references
		{
			Code: `let value: Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": true,
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": "Use '{}' instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: Banned[];`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": "Use '{}' instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: [Banned];`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": "Use '{}' instead.",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `let value: Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": "",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},

		// Invalid cases - with fix
		{
			Code: `let b: { c: Banned };`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`let b: { c: Ok };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `1 as Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`1 as Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    6,
				},
			},
		},
		{
			Code: `class Derived implements Banned {}`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`class Derived implements Ok {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    26,
				},
			},
		},
		{
			Code: `class Derived implements Banned1, Banned2 {}`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned1": map[string]interface{}{
						"fixWith": "Ok1",
						"message": "Use Ok1 instead.",
					},
					"Banned2": map[string]interface{}{
						"fixWith": "Ok2",
						"message": "Use Ok2 instead.",
					},
				},
			},
			Output: []string{`class Derived implements Ok1, Ok2 {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    26,
				},
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    35,
				},
			},
		},
		{
			Code: `interface Derived extends Banned {}`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`interface Derived extends Ok {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    27,
				},
			},
		},
		{
			Code: `type Intersection = Banned & {};`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`type Intersection = Ok & {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			Code: `type Union = Banned | {};`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`type Union = Ok | {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code: `let value: NS.Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"NS.Banned": map[string]interface{}{
						"fixWith": "NS.Ok",
						"message": "Use NS.Ok instead.",
					},
				},
			},
			Output: []string{`let value: NS.Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: {} = {};`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"{}": map[string]interface{}{
						"fixWith": "object",
						"message": "Use object instead.",
					},
				},
			},
			Output: []string{`let value: object = {};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: NS.Banned;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"  NS.Banned  ": map[string]interface{}{
						"fixWith": "NS.Ok",
						"message": "Use NS.Ok instead.",
					},
				},
			},
			Output: []string{`let value: NS.Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: `let value: Type<   Banned   >;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"       Banned      ": map[string]interface{}{
						"fixWith": "Ok",
						"message": "Use Ok instead.",
					},
				},
			},
			Output: []string{`let value: Type<   Ok   >;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    20,
				},
			},
		},
		{
			Code: `type Intersection = Banned<any>;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned<any>": "Don't use `any` as a type parameter to `Banned`",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			Code: `type Intersection = Banned<A,B>;`,
			Options: map[string]interface{}{
				"types": map[string]interface{}{
					"Banned<A, B>": "Don't pass `A, B` as parameters to `Banned`",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Line:      1,
					Column:    21,
				},
			},
		},
	})
}
