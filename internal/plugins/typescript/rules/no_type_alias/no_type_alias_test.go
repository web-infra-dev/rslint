package no_type_alias

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTypeAliasRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoTypeAliasRule, []rule_tester.ValidTestCase{
		// Baseline: interface declarations are always fine
		{Code: `interface Foo { bar: string }`},

		// allowAliases: always (top-level alias allowed)
		{Code: `type Foo = string`, Options: map[string]interface{}{"allowAliases": "always"}},
		{Code: `type Foo = 'a'`, Options: map[string]interface{}{"allowAliases": "always"}},
		{Code: `type Foo = typeof bar`, Options: map[string]interface{}{"allowAliases": "always"}},
		{Code: `type Foo = typeof import('foo')`, Options: map[string]interface{}{"allowAliases": "always"}},

		// allowAliases: in-unions
		{Code: `type Foo = 'a' | 'b'`, Options: map[string]interface{}{"allowAliases": "in-unions"}},
		{Code: `type Foo = 'a' | 'b' | 'c'`, Options: map[string]interface{}{"allowAliases": "in-unions"}},
		{Code: `type Foo = typeof bar | typeof baz`, Options: map[string]interface{}{"allowAliases": "in-unions"}},

		// allowAliases: in-intersections
		{Code: `type Foo = 'a' & 'b'`, Options: map[string]interface{}{"allowAliases": "in-intersections"}},
		{Code: `type Foo = 'a' & 'b' & 'c'`, Options: map[string]interface{}{"allowAliases": "in-intersections"}},

		// allowAliases: in-unions-and-intersections
		{Code: `type Foo = 'a' | 'b'`, Options: map[string]interface{}{"allowAliases": "in-unions-and-intersections"}},
		{Code: `type Foo = 'a' & 'b'`, Options: map[string]interface{}{"allowAliases": "in-unions-and-intersections"}},
		{Code: `type Foo = 'a' | ('b' & 'c')`, Options: map[string]interface{}{"allowAliases": "in-unions-and-intersections"}},

		// allowCallbacks
		{Code: `type Foo = () => void`, Options: map[string]interface{}{"allowCallbacks": "always"}},
		{Code: `type Foo = () => void | string`, Options: map[string]interface{}{"allowCallbacks": "always"}},

		// allowLiterals
		{Code: `type Foo = {}`, Options: map[string]interface{}{"allowLiterals": "always"}},
		{Code: `type Foo = {} | {}`, Options: map[string]interface{}{"allowLiterals": "in-unions"}},
		{Code: `type Foo = {} | {}`, Options: map[string]interface{}{"allowLiterals": "in-unions-and-intersections"}},
		{Code: `type Foo = {} & {}`, Options: map[string]interface{}{"allowLiterals": "in-intersections"}},
		{Code: `type Foo = {} & {}`, Options: map[string]interface{}{"allowLiterals": "in-unions-and-intersections"}},

		// allowMappedTypes
		{Code: `type Foo<T> = { readonly [P in keyof T]: T[P] }`, Options: map[string]interface{}{"allowMappedTypes": "always"}},
		{Code: `type Foo<T> = { readonly [P in keyof T]: T[P] } | { readonly [P in keyof T]: T[P] }`, Options: map[string]interface{}{"allowMappedTypes": "in-unions"}},
		{Code: `type Foo<T> = { readonly [P in keyof T]: T[P] } & { readonly [P in keyof T]: T[P] }`, Options: map[string]interface{}{"allowMappedTypes": "in-intersections"}},
		{Code: `type Foo<T> = { readonly [P in keyof T]: T[P] } | { readonly [P in keyof T]: T[P] }`, Options: map[string]interface{}{"allowMappedTypes": "in-unions-and-intersections"}},

		// allowTupleTypes
		{Code: `type Foo = [string]`, Options: map[string]interface{}{"allowTupleTypes": "always"}},
		{Code: `type Foo = [string] | [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "in-unions"}},
		{Code: `type Foo = [string] & [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "in-intersections"}},
		{Code: `type Foo = ([string] & [number, number]) | [number, number, number]`, Options: map[string]interface{}{"allowTupleTypes": "in-unions-and-intersections"}},
		{Code: `type Foo = readonly [string] | [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "always"}},
		{Code: `type Foo = readonly [string] | readonly [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "always"}},
		{Code: `type Foo = keyof [string]`, Options: map[string]interface{}{"allowTupleTypes": "always"}},
		{Code: `type Foo = keyof [string] | [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "always"}},
		{Code: `type Foo = keyof [string] | [number, number]`, Options: map[string]interface{}{"allowTupleTypes": "in-unions"}},

		// allowConditionalTypes
		{Code: `type Foo<T> = T extends number ? number : null`, Options: map[string]interface{}{"allowConditionalTypes": "always"}},

		// allowConstructors
		{Code: `type Foo = new (bar: number) => string | null`, Options: map[string]interface{}{"allowConstructors": "always"}},

		// allowGenerics
		{Code: `type Foo = Record<string, number>`, Options: map[string]interface{}{"allowGenerics": "always"}},

		// Mixed options (real-world style)
		{
			Code: `type ClassValue = string | number | undefined | null | false`,
			Options: map[string]interface{}{
				"allowAliases":       "in-unions-and-intersections",
				"allowCallbacks":       "always",
				"allowLiterals":        "in-unions-and-intersections",
				"allowMappedTypes":     "in-unions-and-intersections",
				"allowTupleTypes":      "in-unions-and-intersections",
				"allowGenerics":        "always",
				"allowConditionalTypes": "always",
				"allowConstructors":    "always",
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Default: no aliases allowed
		{
			Code:   `type Foo = string`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:   `type Foo = 'a'`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:   `type Foo = typeof import('foo')`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},

		// Union aliases with default / never
		{
			Code: `type Foo = 'a' | 'b'`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code:    `type Foo = 'a' | 'b'`,
			Options: map[string]interface{}{"allowAliases": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code: `type Foo = 'a' | 'b' | 'c'`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
				{MessageId: "noCompositionAlias", Line: 1, Column: 24},
			},
		},
		{
			Code:    `type Foo = 'a' | 'b' | 'c'`,
			Options: map[string]interface{}{"allowAliases": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
				{MessageId: "noCompositionAlias", Line: 1, Column: 24},
			},
		},

		// Intersection aliases with default / never
		{
			Code: `type Foo = 'a' & 'b'`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code:    `type Foo = 'a' & 'b'`,
			Options: map[string]interface{}{"allowAliases": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code: `type Foo = 'a' & 'b' & 'c'`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
				{MessageId: "noCompositionAlias", Line: 1, Column: 24},
			},
		},

		// Composition mismatches for allowAliases
		{
			Code:    `type Foo = 'a' | 'b'`,
			Options: map[string]interface{}{"allowAliases": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code:    `type Foo = 'a' & 'b'`,
			Options: map[string]interface{}{"allowAliases": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},
		{
			Code:    `type Foo = 'a' | 'b' | 'c'`,
			Options: map[string]interface{}{"allowAliases": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
				{MessageId: "noCompositionAlias", Line: 1, Column: 24},
			},
		},
		{
			Code:    `type Foo = 'a' | 'b'`,
			Options: map[string]interface{}{"allowAliases": "in-intersections", "allowLiterals": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 18},
			},
		},

		// allowCallbacks: never (default)
		{
			Code:    `type Foo = () => void`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = () => void`,
			Options: map[string]interface{}{"allowCallbacks": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},

		// allowLiterals: never (default)
		{
			Code:   `type Foo = {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = {}`,
			Options: map[string]interface{}{"allowLiterals": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = {}`,
			Options: map[string]interface{}{"allowLiterals": "in-unions"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = {}`,
			Options: map[string]interface{}{"allowLiterals": "in-intersections"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = {}`,
			Options: map[string]interface{}{"allowLiterals": "in-unions-and-intersections"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},

		// Literal unions / intersections
		{
			Code: `type Foo = {} | {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},
		{
			Code:    `type Foo = {} | {}`,
			Options: map[string]interface{}{"allowLiterals": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},
		{
			Code:    `type Foo = {} | {}`,
			Options: map[string]interface{}{"allowLiterals": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},
		{
			Code: `type Foo = {} & {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},
		{
			Code:    `type Foo = {} & {}`,
			Options: map[string]interface{}{"allowLiterals": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},
		{
			Code:    `type Foo = {} & {}`,
			Options: map[string]interface{}{"allowLiterals": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 17},
			},
		},

		// allowMappedTypes: never (default)
		{
			Code:    `type Foo<T> = { readonly [P in keyof T]: T[P] }`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 15}},
		},
		{
			Code:    `type Foo<T> = { readonly [P in keyof T]: T[P] }`,
			Options: map[string]interface{}{"allowMappedTypes": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 15}},
		},
		{
			Code:    `type Foo<T> = { readonly [P in keyof T]: T[P] } | { readonly [P in keyof T]: T[P] }`,
			Options: map[string]interface{}{"allowMappedTypes": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 15},
				{MessageId: "noCompositionAlias", Line: 1, Column: 51},
			},
		},
		{
			Code:    `type Foo<T> = { readonly [P in keyof T]: T[P] } & { readonly [P in keyof T]: T[P] }`,
			Options: map[string]interface{}{"allowMappedTypes": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 15},
				{MessageId: "noCompositionAlias", Line: 1, Column: 51},
			},
		},

		// allowTupleTypes: never (default)
		{
			Code:    `type Foo = [string]`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = [string]`,
			Options: map[string]interface{}{"allowTupleTypes": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = [string] | [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 23},
			},
		},
		{
			Code:    `type Foo = [string] & [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 23},
			},
		},
		{
			Code:    `type Foo = readonly [string] | [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 32},
			},
		},
		{
			Code:    `type Foo = readonly [string] & [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 32},
			},
		},
		{
			Code:    `type Foo = [string] & readonly [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-unions"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 23},
			},
		},
		{
			Code:    `type Foo = keyof [string] | [number, number]`,
			Options: map[string]interface{}{"allowTupleTypes": "in-intersections"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noCompositionAlias", Line: 1, Column: 12},
				{MessageId: "noCompositionAlias", Line: 1, Column: 29},
			},
		},

		// allowConditionalTypes: never (default)
		{
			Code:    `type Foo<T> = T extends number ? number : null`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 15}},
		},
		{
			Code:    `type Foo<T> = T extends number ? number : null`,
			Options: map[string]interface{}{"allowConditionalTypes": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 15}},
		},

		// allowConstructors: never (default)
		{
			Code:    `type Foo = new (bar: number) => string | null`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = new (bar: number) => string | null`,
			Options: map[string]interface{}{"allowConstructors": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},

		// allowGenerics: never (default)
		{
			Code:    `type Foo = Record<string, number>`,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
		{
			Code:    `type Foo = Record<string, number>`,
			Options: map[string]interface{}{"allowGenerics": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noTypeAlias", Line: 1, Column: 12}},
		},
	})
}
