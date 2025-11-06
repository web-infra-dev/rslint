package no_extraneous_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraneousClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraneousClassRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `class Foo {
  public prop = 1;
  constructor() {}
}`},
			{Code: `export class CClass extends BaseClass {
  public static helper(): void {}
  private static privateHelper(): boolean {
    return true;
  }
  constructor() {}
}`},
			{Code: `class Foo {
  constructor(public bar: string) {}
}`},
			{Code: `class Foo {}`, Options: map[string]interface{}{"allowEmpty": true}},
			{Code: `class Foo {
  constructor() {}
}`, Options: map[string]interface{}{"allowConstructorOnly": true}},
			{Code: `export class Bar {
  public static helper(): void {}
  private static privateHelper(): boolean {
    return true;
  }
}`, Options: map[string]interface{}{"allowStaticOnly": true}},
			{Code: `export default class {
  hello() {
    return 'I am foo!';
  }
}`},
			{Code: `abstract class Foo {
  abstract property: string;
}`},
			{Code: `abstract class Foo {
  abstract method(): string;
}`},
			{Code: `class Foo {
  prop: string;
}`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `class Foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "empty"},
				},
			},
			{
				Code: `class Foo {
  constructor() {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyConstructor"},
				},
			},
			{
				Code: `export class Bar {
  public static helper(): void {}
  private static privateHelper(): boolean {
    return true;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyStatic"},
				},
			},
			{
				Code: `export default class {
  static hello() {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyStatic"},
				},
			},
			{
				Code: `abstract class Foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "empty"},
				},
			},
			{
				Code: `abstract class Foo {
  static property: string;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyStatic"},
				},
			},
			{
				Code: `abstract class Foo {
  constructor() {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyConstructor"},
				},
			},
		},
	)
}
