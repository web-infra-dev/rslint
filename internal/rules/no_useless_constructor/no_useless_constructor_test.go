package no_useless_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUselessConstructorRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: "class A {}"},
		{Code: `
class A {
  constructor() {
    doSomething();
  }
}`},
		{Code: `
class A extends B {
  constructor() {}
}`},
		{Code: `
class A extends B {
  constructor() {
    super('foo');
  }
}`},
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo, bar, 1);
  }
}`},
		{Code: `
class A extends B {
  constructor() {
    super();
    doSomething();
  }
}`},
		{Code: `
class A extends B {
  constructor(...args) {
    super(...args);
    doSomething();
  }
}`},
		{Code: `
class A {
  dummyMethod() {
    doSomething();
  }
}`},
		{Code: `
class A extends B.C {
  constructor() {
    super(foo);
  }
}`},
		{Code: `
class A extends B.C {
  constructor([a, b, c]) {
    super(...arguments);
  }
}`},
		{Code: `
class A extends B.C {
  constructor(a = f()) {
    super(...arguments);
  }
}`},
		{Code: `
class A extends B {
  constructor(a, b, c) {
    super(a, b);
  }
}`},
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo);
  }
}`},
		{Code: `
class A extends B {
  constructor(test) {
    super();
  }
}`},
		{Code: `
class A extends B {
  constructor() {
    foo;
  }
}`},
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(bar);
  }
}`},
		{Code: `
declare class A {
  constructor();
}`},
		{Code: `
class A {
  constructor();
}`},
		{Code: `
abstract class A {
  constructor();
}`},
		{Code: `
class A {
  constructor(private name: string) {}
}`},
		{Code: `
class A {
  constructor(public name: string) {}
}`},
		{Code: `
class A {
  constructor(protected name: string) {}
}`},
		{Code: `
class A {
  private constructor() {}
}`},
		{Code: `
class A {
  protected constructor() {}
}`},
		{Code: `
class A extends B {
  public constructor() {}
}`},
		{Code: `
class A extends B {
  protected constructor(foo, bar) {
    super(bar);
  }
}`},
		{Code: `
class A extends B {
  private constructor(foo, bar) {
    super(bar);
  }
}`},
		{Code: `
class A extends B {
  public constructor(foo) {
    super(foo);
  }
}`},
		{Code: `
class A extends B {
  public constructor(foo) {}
}`},
		{Code: `
class A {
  constructor(foo);
}`},
		{Code: `
class A extends Object {
  constructor(@Foo foo: string) {
    super(foo);
  }
}`},
		{Code: `
class A extends Object {
  constructor(foo: string, @Bar() bar) {
    super(foo, bar);
  }
}`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `
class A {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor(foo) {
    super(foo);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo, bar);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor(...args) {
    super(...args);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B.C {
  constructor() {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor(a, b, ...c) {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A extends B {
  constructor(a, b, ...c) {
    super(a, b, ...c);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
		{
			Code: `
class A {
  public constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noUselessConstructor", Line: 3, Column: 3}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessConstructorRule, validTestCases, invalidTestCases)
}