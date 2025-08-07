package no_empty_interface

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmpty",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `interface Foo extends {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmpty",
					Line:      1,
					Column:    11,
					EndLine:   1,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
			Output: []string{`type Foo = Array<number>`},
		},
		{
			Code: "interface Foo extends Array<number | {}> {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      5,
					Column:    11,
					EndLine:   5,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      3,
					Column:    11,
					EndLine:   3,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      2,
					Column:    11,
					EndLine:   2,
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
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      4,
					Column:    20,
					EndLine:   4,
					EndColumn: 23,
				},
			},
			// Note: In Go tests, the filename is always "file.ts", so the ambient declaration
			// check based on .d.ts filename won't trigger. However, the rule will still
			// report the error and provide a fix (not a suggestion) in this test.
			// The TypeScript tests properly test the .d.ts behavior with suggestions.
			Output: []string{`
declare module FooBar {
  type Baz = typeof baz;
  export type Bar = Baz
}
`},
		},
	})
}
