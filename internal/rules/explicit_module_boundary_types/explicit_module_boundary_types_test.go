package explicit_module_boundary_types

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestExplicitModuleBoundaryTypesRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: "function test(): void { return; }"},
		{Code: "export function test(): void { return; }"},
		{Code: "export var fn = function (): number { return 1; };"},
		{Code: "export var arrowFn = (): string => 'test';"},
		{Code: `
class Test {
  constructor(one) {}
  get prop(one) {
    return 1;
  }
  set prop(one) {}
  method(one) {
    return;
  }
  arrow = one => 'arrow';
  abstract abs(one);
}`},
		{Code: `
export class Test {
  constructor(one: string) {}
  get prop(one: string): void {
    return 1;
  }
  set prop(one: string): void {}
  method(one: string): void {
    return;
  }
  arrow = (one: string): string => 'arrow';
  abstract abs(one: string): void;
}`},
		{Code: `
export class Test {
  private constructor(one) {}
  private get prop(one) {
    return 1;
  }
  private set prop(one) {}
  private method(one) {
    return;
  }
  private arrow = one => 'arrow';
  private abstract abs(one);
}`},
		{Code: "export class PrivateProperty { #property = () => null; }"},
		{Code: "export class PrivateMethod { #method() {} }"},
		{Code: `
export class Test {
  constructor();
  constructor(value?: string) {
    console.log(value);
  }
}`},
		{Code: `
declare class MyClass {
  constructor(options?: MyClass.Options);
}
export { MyClass };`},
		{Code: `
export function test(): void {
  nested();
  return;

  function nested() {}
}`},
		{Code: `
export function test(): string {
  const nested = () => 'value';
  return nested();
}`},
		{Code: `
export function test(): string {
  class Nested {
    public method() {
      return 'value';
    }
  }
  return new Nested().method();
}`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `
export function test(a: number, b: number) {
  return;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 8}},
		},
		{
			Code: `
export function test() {
  return;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 8}},
		},
		{
			Code: `
export var fn = function () {
  return 1;
};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 17}},
		},
		{
			Code: `export var arrowFn = () => 'test';`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 25}},
		},
		{
			Code: `
export class Test {
  constructor() {}
  get prop() {
    return 1;
  }
  set prop(value) {}
  method() {
    return;
  }
  arrow = arg => 'arrow';
  private method() {
    return;
  }
  abstract abs(arg);
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3},
				{MessageId: "missingArgType", Line: 7, Column: 12},
				{MessageId: "missingReturnType", Line: 8, Column: 3},
				{MessageId: "missingReturnType", Line: 11, Column: 3},
				{MessageId: "missingArgType", Line: 11, Column: 11},
				{MessageId: "missingReturnType", Line: 15, Column: 15},
				{MessageId: "missingArgType", Line: 15, Column: 16},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitModuleBoundaryTypesRule, validTestCases, invalidTestCases)
}