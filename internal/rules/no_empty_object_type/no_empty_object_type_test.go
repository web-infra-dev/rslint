package no_empty_object_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoEmptyObjectTypeRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Valid cases
		{
			Code: `
interface Base {
  name: string;
}
`,
		},
		{
			Code: `
interface Base {
  name: string;
}

interface Derived {
  age: number;
}

// valid because extending multiple interfaces can be used instead of a union type
interface Both extends Base, Derived {}
`,
		},
		{
			Code: `interface Base {}`,
			Options: map[string]interface{}{
				"allowInterfaces": "always",
			},
		},
		{
			Code: `
interface Base {
  name: string;
}

interface Derived extends Base {}
`,
			Options: map[string]interface{}{
				"allowInterfaces": "with-single-extends",
			},
		},
		{
			Code: `let value: object;`,
		},
		{
			Code: `let value: Object;`,
		},
		{
			Code: `let value: { inner: true };`,
		},
		{
			Code: `type MyNonNullable<T> = T & {};`,
		},
		{
			Code: `type Base = {};`,
			Options: map[string]interface{}{
				"allowObjectTypes": "always",
			},
		},
		{
			Code: `type Base = {};`,
			Options: map[string]interface{}{
				"allowWithName": "Base",
			},
		},
		{
			Code: `type BaseProps = {};`,
			Options: map[string]interface{}{
				"allowWithName": "Props$",
			},
		},
		{
			Code: `interface Base {}`,
			Options: map[string]interface{}{
				"allowWithName": "Base",
			},
		},
		{
			Code: `interface BaseProps {}`,
			Options: map[string]interface{}{
				"allowWithName": "Props$",
			},
		},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `interface Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = object",
						},
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = unknown",
						},
					},
				},
			},
		},
		{
			Code: `interface Base {}`,
			Options: map[string]interface{}{
				"allowInterfaces": "never",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = object",
						},
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = unknown",
						},
					},
				},
			},
		},
		{
			Code: `
interface Base {
  props: string;
}

interface Derived extends Base {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      6,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
interface Base {
  props: string;
}

type Derived = Base
`,
						},
					},
				},
			},
		},
		{
			Code: `interface Base extends Array<number> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output:    "type Base = Array<number>",
						},
					},
				},
			},
		},
		{
			Code: `interface Base<T> extends Derived<T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output:    "type Base<T> = Derived<T>",
						},
					},
				},
			},
		},
		{
			Code: `type Base = {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = object;",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = unknown;",
						},
					},
				},
			},
		},
		{
			Code: `type Base = {};`,
			Options: map[string]interface{}{
				"allowObjectTypes": "never",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = object;",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = unknown;",
						},
					},
				},
			},
		},
		{
			Code: `let value: {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "let value: object;",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "let value: unknown;",
						},
					},
				},
			},
		},
		{
			Code: `type MyUnion<T> = T | {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type MyUnion<T> = T | object;",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type MyUnion<T> = T | unknown;",
						},
					},
				},
			},
		},
		{
			Code: `type Base = {};`,
			Options: map[string]interface{}{
				"allowWithName": "Mismatch",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = object;",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "type Base = unknown;",
						},
					},
				},
			},
		},
		{
			Code: `interface Base {}`,
			Options: map[string]interface{}{
				"allowWithName": ".*Props$",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = object",
						},
						{
							MessageId: "replaceEmptyInterface",
							Output:    "type Base = unknown",
						},
					},
				},
			},
		},
		{
			Code: `interface Base extends Array<number | {}> {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      1,
					Column:    11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output:    "type Base = Array<number | {}>",
						},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "interface Base extends Array<number | object> {}",
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output:    "interface Base extends Array<number | unknown> {}",
						},
					},
				},
			},
		},
		{
			Code: `
let value: {
  /* ... */
};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      2,
					Column:    12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyObjectType",
							Output: `
let value: object;
`,
						},
						{
							MessageId: "replaceEmptyObjectType",
							Output: `
let value: unknown;
`,
						},
					},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyObjectTypeRule, validTestCases, invalidTestCases)
}
