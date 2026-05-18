// TestNoUnnecessaryParameterPropertyAssignmentUpstream migrates the full
// valid/invalid suite from upstream typescript-eslint's
//
//	packages/eslint-plugin/tests/rules/no-unnecessary-parameter-property-assignment.test.ts
//
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases (Dimension 4 edge shapes, branch lock-ins,
// real-user issue shapes) live in
// no_unnecessary_parameter_property_assignment_extras_test.go.
package no_unnecessary_parameter_property_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryParameterPropertyAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryParameterPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `
class Foo {
  constructor(foo: string) {}
}
    `},
			{Code: `
class Foo {
  constructor(private foo: string) {}
}
    `},
			{Code: `
class Foo {
  constructor(private foo: string) {
    this.foo = bar;
  }
}
    `},
			{Code: `
class Foo {
  constructor(private foo: any) {
    this.foo = foo.bar;
  }
}
    `},
			{Code: `
class Foo {
  constructor(private foo: string) {
    this.foo = this.bar;
  }
}
    `},
			{Code: `
class Foo {
  foo: string;
  constructor(foo: string) {
    this.foo = foo;
  }
}
    `},
			{Code: `
class Foo {
  bar: string;
  constructor(private foo: string) {
    this.bar = foo;
  }
}
    `},
			{Code: `
class Foo {
  constructor(private foo: string) {
    this.bar = () => {
      this.foo = foo;
    };
  }
}
    `},
			{Code: "\nclass Foo {\n  constructor(private foo: string) {\n    this[`${foo}`] = foo;\n  }\n}\n    "},
			{Code: `
function Foo(foo) {
  this.foo = foo;
}
    `},
			{Code: `
const foo = 'foo';
this.foo = foo;
    `},
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo += foo;
    this.foo -= foo;
    this.foo *= foo;
    this.foo /= foo;
    this.foo %= foo;
    this.foo **= foo;
  }
}
    `},
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo += 1;
    this.foo = foo;
  }
}
    `},
			{Code: `
class Foo {
  constructor(
    public foo: number,
    bar: boolean,
  ) {
    if (bar) {
      this.foo += 1;
    } else {
      this.foo = foo;
    }
  }
}
    `},
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo = foo;
  }
  init = (this.foo += 1);
}
    `},
			{Code: `
class Foo {
  constructor(public foo: number) {
    {
      const foo = 1;
      this.foo = foo;
    }
  }
}
    `},
			{Code: `
declare const name: string;
class Foo {
  constructor(public foo: number) {
    this[name] = foo;
  }
}
    `},
			{Code: `
declare const name: string;
class Foo {
  constructor(public foo: number) {
    Foo.foo = foo;
  }
}
    `},
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo = foo;
  }
  init = (() => {
    this.foo += 1;
  })();
}
    `},
			{Code: `
declare const name: string;
class Foo {
  constructor(public foo: number) {
    this[name] = foo;
  }
  init = (this[name] = 1);
  init2 = (Foo.foo = 1);
}
    `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo?: string) {
    this.foo = foo!;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo?: string) {
    this.foo = foo as any;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo = '') {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo = '') {
    this.foo = foo;
    this.foo += 'foo';
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo ||= foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo ??= foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo &&= foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    this['foo'] = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    function bar() {
      this.foo = foo;
    }
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 7, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    this.bar = () => {
      this.foo = foo;
    };
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 7, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    class Bar {
      constructor(private foo: string) {
        this.foo = foo;
      }
    }
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 6, Column: 9},
					{MessageId: "unnecessaryAssign", Line: 9, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    this.foo = foo;
  }
  bar = () => {
    this.foo = 'foo';
  };
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    this.foo = foo;
  }
  init = foo => {
    this.foo = foo;
  };
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    this.foo = foo;
  }
  init = class Bar {
    constructor(private foo: string) {
      this.foo = foo;
    }
  };
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 8, Column: 7},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    {
      this.foo = foo;
    }
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 7},
				},
			},
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    (() => {
      this.foo = foo;
    })();
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 7},
				},
			},
		},
	)
}
