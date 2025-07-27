package no_empty_interface

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoEmptyInterfaceRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyInterfaceRule, []rule_tester.ValidTestCase{
		// Valid cases
		{Code: `
interface Foo {
  name: string;
}
`},
		{Code: `
interface Foo {
  name: string;
}

interface Bar {
  age: number;
}

// valid because extending multiple interfaces can be used instead of a union type
interface Baz extends Foo, Bar {}
`},
		{
			Code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
`,
			Options: map[string]interface{}{"allowSingleExtends": true},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
`,
			Options: map[string]interface{}{"allowSingleExtends": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases
		{
			Code: "interface Foo {}",
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An empty interface is equivalent to `{}`.",
					Line:     1,
					Column:   11,
					EndLine:  1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `interface Foo extends {}`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An empty interface is equivalent to `{}`.",
					Line:     1,
					Column:   11,
					EndLine:  1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Baz {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     6,
					Column:   11,
					EndLine:  6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  props: string;
}

type Bar = Foo

class Baz {}
`},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     6,
					Column:   11,
					EndLine:  6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			// No output when merged with class
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

const bar = class Bar {};
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     6,
					Column:   11,
					EndLine:  6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  props: string;
}

type Bar = Foo

const bar = class Bar {};
`},
		},
		{
			Code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     6,
					Column:   11,
					EndLine:  6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  name: string;
}

type Bar = Foo
`},
		},
		{
			Code: "interface Foo extends Array<number> {}",
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     1,
					Column:   11,
					EndLine:  1,
					EndColumn: 14,
				},
			},
			Output: []string{`type Foo = Array<number>`},
		},
		{
			Code: "interface Foo extends Array<number | {}> {}",
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     1,
					Column:   11,
					EndLine:  1,
					EndColumn: 14,
				},
			},
			Output: []string{`type Foo = Array<number | {}>`},
		},
		{
			Code: `
interface Bar {
  bar: string;
}
interface Foo extends Array<Bar> {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     5,
					Column:   11,
					EndLine:  5,
					EndColumn: 14,
				},
			},
			Output: []string{`
interface Bar {
  bar: string;
}
type Foo = Array<Bar>
`},
		},
		{
			Code: `
type R = Record<string, unknown>;
interface Foo extends R {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     3,
					Column:   11,
					EndLine:  3,
					EndColumn: 14,
				},
			},
			Output: []string{`
type R = Record<string, unknown>;
type Foo = R
`},
		},
		{
			Code: `
interface Foo<T> extends Bar<T> {}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     2,
					Column:   11,
					EndLine:  2,
					EndColumn: 14,
				},
			},
			Output: []string{`
type Foo<T> = Bar<T>
`},
		},
	})
}

// Test case for ambient declarations in .d.ts files
func TestNoEmptyInterfaceRuleAmbient(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyInterfaceRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare module FooBar {
  type Baz = typeof baz;
  export interface Bar extends Baz {}
}
`,
			Errors: []rule_tester.ExpectedDiagnostic{
				{
					Message:  "An interface declaring no members is equivalent to its supertype.",
					Line:     4,
					Column:   20,
					EndLine:  4,
					EndColumn: 23,
				},
			},
			Filename: "test.d.ts",
			// No output for ambient declarations, should use suggestions instead
		},
	})
}