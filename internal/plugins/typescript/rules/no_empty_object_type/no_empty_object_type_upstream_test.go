// TestNoEmptyObjectTypeUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/tests/rules/no-empty-object-type.test.ts 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the no_empty_object_type_extras_test.go file(s).
package no_empty_object_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyObjectTypeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyObjectTypeRule, []rule_tester.ValidTestCase{
		{Code: `
interface Base {
  name: string;
}
`},
		{Code: `
interface Base {
  name: string;
}

interface Derived {
  age: number;
}

// valid because extending multiple interfaces can be used instead of a union type
interface Both extends Base, Derived {}
`},
		{
			Code:    `interface Base {}`,
			Options: map[string]interface{}{"allowInterfaces": "always"},
		},
		{
			Code: `
interface Base {
  name: string;
}

interface Derived extends Base {}
`,
			Options: map[string]interface{}{"allowInterfaces": "with-single-extends"},
		},
		{
			Code: `
interface Base {
  props: string;
}

interface Derived extends Base {}

class Derived {}
`,
			Options: map[string]interface{}{"allowInterfaces": "with-single-extends"},
		},
		{Code: `let value: object;`},
		{Code: `let value: Object;`},
		{Code: `let value: { inner: true };`},
		{Code: `type MyNonNullable<T> = T & {};`},
		{
			Code:    `type Base = {};`,
			Options: map[string]interface{}{"allowObjectTypes": "always"},
		},
		{
			Code:    `type Base = {};`,
			Options: map[string]interface{}{"allowWithName": "Base"},
		},
		{
			Code:    `type BaseProps = {};`,
			Options: map[string]interface{}{"allowWithName": "Props$"},
		},
		{
			Code:    `interface Base {}`,
			Options: map[string]interface{}{"allowWithName": "Base"},
		},
		{
			Code:    `interface BaseProps {}`,
			Options: map[string]interface{}{"allowWithName": "Props$"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `interface Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Base = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Base = unknown`},
					},
				},
			},
		},
		{
			Code:    `interface Base {}`,
			Options: map[string]interface{}{"allowInterfaces": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Base = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Base = unknown`},
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

class Other {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
interface Base {
  props: string;
}

type Derived = Base

class Other {}
`,
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

class Derived {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
interface Base {
  props: string;
}

interface Derived extends Base {}

const derived = class Derived {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
interface Base {
  props: string;
}

type Derived = Base

const derived = class Derived {};
`,
						},
					},
				},
			},
		},
		{
			Code: `
interface Base {
  name: string;
}

interface Derived extends Base {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
interface Base {
  name: string;
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
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Base = Array<number>`},
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
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Base = Array<number | {}>`},
					},
				},
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    39,
					EndLine:   1,
					EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `interface Base extends Array<number | object> {}`},
						{MessageId: "replaceEmptyObjectType", Output: `interface Base extends Array<number | unknown> {}`},
					},
				},
			},
		},
		{
			Code: `
interface Derived {
  property: string;
}
interface Base extends Array<Derived> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      5,
					Column:    11,
					EndLine:   5,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
interface Derived {
  property: string;
}
type Base = Array<Derived>
`,
						},
					},
				},
			},
		},
		{
			Code: `
type R = Record<string, unknown>;
interface Base extends R {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
type R = Record<string, unknown>;
type Base = R
`,
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
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterfaceWithSuper", Output: `type Base<T> = Derived<T>`},
					},
				},
			},
		},
		{
			Code: `
declare namespace BaseAndDerived {
  type Base = typeof base;
  export interface Derived extends Base {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterfaceWithSuper",
					Line:      4,
					Column:    20,
					EndLine:   4,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "replaceEmptyInterfaceWithSuper",
							Output: `
declare namespace BaseAndDerived {
  type Base = typeof base;
  export type Derived = Base
}
`,
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
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Base = object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Base = unknown;`},
					},
				},
			},
		},
		{
			Code:    `type Base = {};`,
			Options: map[string]interface{}{"allowObjectTypes": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Base = object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Base = unknown;`},
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
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `let value: object;`},
						{MessageId: "replaceEmptyObjectType", Output: `let value: unknown;`},
					},
				},
			},
		},
		{
			Code:    `let value: {};`,
			Options: map[string]interface{}{"allowObjectTypes": "never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `let value: object;`},
						{MessageId: "replaceEmptyObjectType", Output: `let value: unknown;`},
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
					EndLine:   4,
					EndColumn: 2,
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
		{
			Code: `type MyUnion<T> = T | {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type MyUnion<T> = T | object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type MyUnion<T> = T | unknown;`},
					},
				},
			},
		},
		{
			Code:    `type Base = {} | null;`,
			Options: map[string]interface{}{"allowWithName": "Base"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Base = object | null;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Base = unknown | null;`},
					},
				},
			},
		},
		{
			Code:    `type Base = {};`,
			Options: map[string]interface{}{"allowWithName": "Mismatch"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyObject",
					Line:      1,
					Column:    13,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyObjectType", Output: `type Base = object;`},
						{MessageId: "replaceEmptyObjectType", Output: `type Base = unknown;`},
					},
				},
			},
		},
		{
			Code:    `interface Base {}`,
			Options: map[string]interface{}{"allowWithName": ".*Props$"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyInterface",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceEmptyInterface", Output: `type Base = object`},
						{MessageId: "replaceEmptyInterface", Output: `type Base = unknown`},
					},
				},
			},
		},
	})
}
