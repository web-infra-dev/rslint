package no_empty_function

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoEmptyFunctionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyFunctionRule, []rule_tester.ValidTestCase{
		// Valid cases - exactly mirroring TypeScript tests
		{
			Code: `
class Person {
  private name: string;
  constructor(name: string) {
    this.name = name;
  }
}
      `,
		},
		{
			Code: `
class Person {
  constructor(private name: string) {}
}
      `,
		},
		{
			Code: `
class Person {
  constructor(name: string) {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"constructors"}},
		},
		{
			Code: `
class Person {
  otherMethod(name: string) {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"methods"}},
		},
		{
			Code: `
class Foo {
  private constructor() {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"private-constructors"}},
		},
		{
			Code: `
class Foo {
  protected constructor() {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"protected-constructors"}},
		},
		{
			Code: `
function foo() {
  const a = null;
}
      `,
		},
		{
			Code: `
class Foo {
  @decorator()
  foo() {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"decoratedFunctions"}},
		},
		{
			Code: `
class Foo extends Base {
  override foo() {}
}
      `,
			Options: map[string]interface{}{"allow": []interface{}{"overrideMethods"}},
		},

		// Additional comprehensive test cases
		{
			Code:    `const foo = () => {};`,
			Options: map[string]interface{}{"allow": []interface{}{"arrowFunctions"}},
		},
		{
			Code:    `function foo() {}`,
			Options: map[string]interface{}{"allow": []interface{}{"functions"}},
		},
		{
			Code:    `async function foo() {}`,
			Options: map[string]interface{}{"allow": []interface{}{"asyncFunctions"}},
		},
		{
			Code:    `function* foo() {}`,
			Options: map[string]interface{}{"allow": []interface{}{"generatorFunctions"}},
		},
		{
			Code: `
class Foo {
  get bar() {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"getters"}},
		},
		{
			Code: `
class Foo {
  set bar(value: string) {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"setters"}},
		},
		{
			Code: `
class Foo {
  async bar() {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"asyncMethods"}},
		},
		{
			Code: `
class Foo {
  *bar() {}
}`,
			Options: map[string]interface{}{"allow": []interface{}{"generatorMethods"}},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - exactly mirroring TypeScript tests
		{
			Code: `
class Person {
  constructor(name: string) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    28,
				},
			},
		},
		{
			Code: `
class Person {
  otherMethod(name: string) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    28,
				},
			},
		},
		{
			Code: `
class Foo {
  private constructor() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    24,
				},
			},
		},
		{
			Code: `
class Foo {
  protected constructor() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    26,
				},
			},
		},
		{
			Code: `
function foo() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      2,
					Column:    15,
				},
			},
		},
		{
			Code: `
class Foo {
  @decorator()
  foo() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      4,
					Column:    8,
				},
			},
		},
		{
			Code: `
class Foo extends Base {
  override foo() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    17,
				},
			},
		},

		// Additional invalid cases for comprehensive coverage
		{
			Code: `const foo = () => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					Column:    18,
				},
			},
		},
		{
			Code: `async function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					Column:    21,
				},
			},
		},
		{
			Code: `function* foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					Column:    16,
				},
			},
		},
		{
			Code: `
class Foo {
  get bar() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    12,
				},
			},
		},
		{
			Code: `
class Foo {
  set bar(value: string) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    25,
				},
			},
		},
		{
			Code: `
class Foo {
  async bar() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    14,
				},
			},
		},
		{
			Code: `
class Foo {
  *bar() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `const foo = function() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					Column:    23,
				},
			},
		},
		{
			Code: `const foo = function bar() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					Column:    27,
				},
			},
		},
	})
}