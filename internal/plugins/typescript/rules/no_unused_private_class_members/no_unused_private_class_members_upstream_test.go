// TestNoUnusedPrivateClassMembersUpstream migrates the full valid/invalid suite
// from typescript-eslint's tests/rules/no-unused-private-class-members.test.ts
// 1:1. Position assertions cover line/column for every invalid case. rslint-
// specific lock-in cases live in no_unused_private_class_members_extras_test.go.
package no_unused_private_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPrivateClassMembersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedPrivateClassMembersRule, []rule_tester.ValidTestCase{
		{Code: `class Foo {}`},
		{Code: `
class Foo {
  publicMember = 42;
}
    `},
		{Code: `
class Foo {
  public publicMember = 42;
}
    `},
		{Code: `
class Foo {
  protected publicMember = 42;
}
    `},
		{Code: `
class C {
  #usedInInnerClass;

  method(a) {
    return class {
      foo = a.#usedInInnerClass;
    };
  }
}
    `},
		{Code: `
class C {
  private accessor accessorMember = 42;

  method() {
    return this.accessorMember;
  }
}
    `},
		{Code: `
class C {
  private static staticMember = 42;

  static method() {
    return this.staticMember;
  }
}
    `},
		{Code: `
class C {
  private static staticMember = 42;

  method() {
    return C.staticMember;
  }
}
      `},
		{Code: `
class Test1 {
  constructor(private parameterProperty: number) {}
  method() {
    return this.parameterProperty;
  }
}
    `},
		{Code: `
class Test1 {
  constructor(private readonly parameterProperty: number) {}
  method() {
    return this.parameterProperty;
  }
}
    `},
		{Code: `
class Test1 {
  constructor(private readonly parameterProperty: number = 1) {}
  method() {
    return this.parameterProperty;
  }
}
    `},
		{Code: `
class Foo {
  private prop: number;

  method(thing: Foo) {
    return thing.prop;
  }
}
    `},
		{Code: `
class Foo {
  private static staticProp: number;

  method(thing: typeof Foo) {
    return thing.staticProp;
  }
}
    `},
		{Code: `
class Foo {
  private prop: number;

  method() {
    const self = this;
    return self.prop;
  }
}
    `},
		{Code: `
class Foo {
  #privateMember = 42;
  method() {
    return this.#privateMember;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 42;
  method() {
    return this.privateMember;
  }
}
    `},
		{Code: `
class Foo {
  #privateMember = 42;
  anotherMember = this.#privateMember;
}
    `},
		{Code: `
class Foo {
  private privateMember = 42;
  anotherMember = this.privateMember;
}
    `},
		{Code: `
class Foo {
  #privateMember = 42;
  foo() {
    anotherMember = this.#privateMember;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 42;
  foo() {
    anotherMember = this.privateMember;
  }
}
    `},
		{Code: `
class C {
  #privateMember;

  foo() {
    bar((this.#privateMember += 1));
  }
}
    `},
		{Code: `
class C {
  private privateMember;

  foo() {
    bar((this.privateMember += 1));
  }
}
    `},
		{Code: `
class Foo {
  #privateMember = 42;
  method() {
    return someGlobalMethod(this.#privateMember);
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 42;
  method() {
    return someGlobalMethod(this.privateMember);
  }
}
    `},
		{Code: `
class C {
  #privateMember;

  foo() {
    return class {};
  }

  bar() {
    return this.#privateMember;
  }
}
    `},
		{Code: `
class C {
  private privateMember;

  foo() {
    return class {};
  }

  bar() {
    return this.privateMember;
  }
}
    `},
		{Code: `
class Foo {
  #privateMember;
  method() {
    for (const bar in this.#privateMember) {
    }
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;
  method() {
    for (const bar in this.privateMember) {
    }
  }
}
    `},
		{Code: `
class Foo {
  #privateMember;
  method() {
    for (const bar of this.#privateMember) {
    }
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;
  method() {
    for (const bar of this.privateMember) {
    }
  }
}
    `},
		{Code: `
class Foo {
  #privateMember;
  method() {
    [bar = 1] = this.#privateMember;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;
  method() {
    [bar = 1] = this.privateMember;
  }
}
    `},
		{Code: `
class Foo {
  #privateMember;
  method() {
    [bar] = this.#privateMember;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;
  method() {
    [bar] = this.privateMember;
  }
}
    `},
		{Code: `
class C {
  #privateMember;

  method() {
    ({ [this.#privateMember]: a } = foo);
  }
}
    `},
		{Code: `
class C {
  private privateMember;

  method() {
    ({ [this.privateMember]: a } = foo);
  }
}
    `},
		{Code: `
class C {
  set #privateMember(value) {
    doSomething(value);
  }
  get #privateMember() {
    return something();
  }
  method() {
    this.#privateMember += 1;
  }
}
    `},
		{Code: `
class C {
  private set privateMember(value) {
    doSomething(value);
  }
  private get privateMember() {
    return something();
  }
  method() {
    this.privateMember += 1;
  }
}
    `},
		{Code: `
class Foo {
  set #privateMember(value) {}

  method(a) {
    [this.#privateMember] = a;
  }
}
    `},
		{Code: `
class Foo {
  private set privateMember(value) {}

  method(a) {
    [this.privateMember] = a;
  }
}
    `},
		{Code: `
class C {
  get #privateMember() {
    return something();
  }
  set #privateMember(value) {
    doSomething(value);
  }
  method() {
    this.#privateMember += 1;
  }
}
    `},
		{Code: `
class C {
  private get privateMember() {
    return something();
  }
  private set privateMember(value) {
    doSomething(value);
  }
  method() {
    this.privateMember += 1;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;
  private privateMember2;

  method() {
    const { privateMember, privateMember2 } = this;
    console.log(privateMember, privateMember2);
  }
}
    `},
		{Code: `
class Foo {
  private static staticMember = 1;
  static method() {
    const { staticMember } = this;
    console.log(staticMember);
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 1;
  method() {
    const { privateMember } = this;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 1;
  method() {
    const { privateMember: privateMember2 } = this;
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 1;

  method() {
    let privateMember;
    ({ privateMember } = this);
  }
}
    `},
		{Code: `
class Foo {
  private privateMember = 1;

  method() {
    const foo = ({ privateMember } = this) => {};
  }
}
    `},
		{Code: `
class Foo {
  private privateMember;

  method() {
    const { privateMember: used } = this;
  }
}
    `},

		// ---- Method definitions (upstream group) ----
		{Code: `
class Foo {
  #privateMember() {
    return 42;
  }
  anotherMethod() {
    return this.#privateMember();
  }
}
    `},
		{Code: `
class Foo {
  private privateMember() {
    return 42;
  }
  anotherMethod() {
    return this.privateMember();
  }
}
    `},
		{Code: `
class C {
  set #privateMember(value) {
    doSomething(value);
  }

  foo() {
    this.#privateMember = 1;
  }
}
    `},
		{Code: `
class C {
  private set privateMember(value) {
    doSomething(value);
  }

  foo() {
    this.privateMember = 1;
  }
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class C {
  #unusedInOuterClass;

  foo() {
    return class D {
      #unusedInOuterClass;

      bar() {
        return this.#unusedInOuterClass;
      }
    };
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class C {
  #unusedOnlyInSecondNestedClass;

  foo() {
    return class {
      #unusedOnlyInSecondNestedClass;

      bar() {
        return this.#unusedOnlyInSecondNestedClass;
      }
    };
  }

  baz() {
    return this.#unusedOnlyInSecondNestedClass;
  }

  bar() {
    return class {
      #unusedOnlyInSecondNestedClass;
    };
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 21, Column: 7},
			},
		},
		{
			Code: `
class C {
  #usedOnlyInTheSecondInnerClass;

  method(a) {
    return class {
      #usedOnlyInTheSecondInnerClass;

      method2(b) {
        foo = b.#usedOnlyInTheSecondInnerClass;
      }

      method3(b) {
        foo = b.#usedOnlyInTheSecondInnerClass;
      }
    };
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class C {
  private accessor accessorMember = 42;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 20},
			},
		},
		{
			Code: `
class C {
  private static staticMember = 42;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 18},
			},
		},
		{
			Code: `
class Test1 {
  constructor(private parameterProperty: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 23},
			},
		},
		{
			Code: `
class Test1 {
  constructor(private readonly parameterProperty: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 32},
			},
		},
		{
			Code: `
class Test1 {
  constructor(private readonly parameterProperty: number = 1) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 32},
			},
		},
		// Usage of a property outside the class — bracket and dot notation.
		{
			Code: `
class C {
  private usedOutsideClass;
}

const instance = new C();
console.log(instance.usedOutsideClass);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class C {
  private usedOutsideClass;
}

const instance = new C();
console.log(instance['usedOutsideClass']);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		// Too much indirection (`const self1 = this; const self2 = self1;`) —
		// the rule intentionally bails after one hop.
		{
			Code: `
class Foo {
  private prop: number;

  method() {
    const self1 = this;
    const self2 = self1;
    return self2.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		{
			Code: `
class Foo {
  #privateMember = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class First {}
class Second {
  #privateMember = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 3},
			},
		},
		{
			Code: `
class First {}
class Second {
  private privateMember = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},
		{
			Code: `
class First {
  #privateMember = 5;
}
class Second {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class First {
  private privateMember = 5;
}
class Second {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class First {
  #privateMember = 5;
  #privateMember2 = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 3},
			},
		},
		{
			Code: `
class First {
  private privateMember = 5;
  private privateMember2 = 5;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember = 5;
  method() {
    this.#privateMember = 42;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember = 5;
  method() {
    this.privateMember = 42;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember = 5;
  method() {
    this.#privateMember += 42;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember = 5;
  method() {
    this.privateMember += 42;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class C {
  #privateMember;

  foo() {
    this.#privateMember++;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class C {
  private privateMember;

  foo() {
    this.privateMember++;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Unused method definitions (upstream group) ----
		{
			Code: `
class Foo {
  #privateMember() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember() {}
  #privateMemberUsed() {
    return 42;
  }
  publicMethod() {
    return this.#privateMemberUsed();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember() {}
  private privateMemberUsed() {
    return 42;
  }
  publicMethod() {
    return this.privateMemberUsed();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  set #privateMember(value) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 7},
			},
		},
		{
			Code: `
class Foo {
  private set privateMember(value) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 15},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    for (this.#privateMember in bar) {
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    for (this.privateMember in bar) {
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    for (this.#privateMember of bar) {
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    for (this.privateMember of bar) {
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    ({ x: this.#privateMember } = bar);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    ({ x: this.privateMember } = bar);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    [...this.#privateMember] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    [...this.privateMember] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    [this.#privateMember = 1] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    [this.privateMember = 1] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  #privateMember;
  method() {
    [this.#privateMember] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;
  method() {
    [this.privateMember] = bar;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
class Foo {
  private privateMember;

  method() {
    const { unused: privateMember } = this;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},
		{
			Code: `
const foo = 'bar';
class Foo {
  private foo = 1;

  method() {
    const { [foo]: test } = this;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},
		{
			Code: `
const foo = 'bar';
class Foo {
  private foo = 1;
  private bar = 2;

  method() {
    const { [foo]: test } = this;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
				{MessageId: "unusedPrivateClassMember", Line: 5, Column: 11},
			},
		},
	})
}
