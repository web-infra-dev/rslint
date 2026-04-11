package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMethodHOF(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// MethodDeclaration returning function — HOF allows outer
		{
			Code: `
class Foo {
  method() {
    return (): void => {};
  }
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// Object method shorthand returning function
		{
			Code: `
const obj = {
  method() {
    return (): void => {};
  },
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// Getter with return type — inner arrow typed by ancestor
		{
			Code: `class A { get prop(): () => void { return () => {}; } }`,
		},
		// Method with return type — inner arrow typed by ancestor
		{
			Code: `
class A {
  method(): () => void {
    return () => {};
  }
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// MethodDeclaration returning function — inner still needs type
		{
			Code: `
class Foo {
  method() {
    return () => {};
  }
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 15}},
		},
		// Object method NOT HOF — not all returns are functions
		{
			Code: `
const obj = {
  method(x: boolean) {
    if (x) return 'string';
    return (): void => {};
  },
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// Getter without return type
		{
			Code: `
class A {
  get prop() { return 1; }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
	})
}

func TestAncestorReturnType(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// Arrow with expression body inside property of returned object — ancestor has return type
		{
			Code: `
interface Foo { bar: () => string; }
function foo(): Foo {
  return { bar: () => 'test' };
}
			`,
		},
		{
			Code: `
interface Foo { bar: () => string; }
const foo = (): Foo => ({ bar: () => 'test' });
			`,
		},
		// Method/Getter with return type — inner arrow typed by ancestor
		{
			Code: `
interface Foo { bar: () => string; }
class A {
  method(): Foo {
    return { bar: () => 'test' };
  }
}
			`,
		},
		{
			Code: `
interface Foo { bar: () => string; }
class A {
  get prop(): Foo {
    return { bar: () => 'test' };
  }
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// Arrow with BLOCK body — ancestorHasReturnType returns false
		// (the PropertyAssignment.Initializer is the arrow, but it has block body so isBodylessArrow=false)
		{
			Code: `
interface Foo { bar: () => string; }
function foo(): Foo {
  return { bar: () => { return 'test'; } };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 12}},
		},
		// Function expression in property — not a bodyless arrow
		{
			Code: `
interface Foo { bar: () => string; }
function foo(): Foo {
  return { bar: function() { return 'test'; } };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 12}},
		},
		// Method shorthand in returned object — always needs return type
		{
			Code: `
interface Foo { bar(): string; }
function foo(): Foo {
  return { bar() { return 'test'; } };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 12}},
		},
	})
}
