package no_namespace

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNamespaceRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNamespaceRule, []rule_tester.ValidTestCase{
		{Code: `
// Regular module declaration (not namespace)
declare module "foo" {
  export const bar: string;
}
    `},
		{Code: `
// Global module augmentation
declare global {
  interface Window {
    foo: string;
  }
}
    `},
		{Code: `
// Ambient module declaration
declare module "bar" {
  export const baz: number;
}
    `},
		{
			Code: `
// Declare namespace (allowed when allowDeclarations is true)
declare namespace Test {
  export const value = 1;
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": true,
			},
		},
		{Code: `
// Regular TypeScript code without namespaces
const value = 1;
function test() {
  return value;
}
class Test {
  constructor() {}
}
    `},
		{Code: `
// Module with exports (not namespace)
export const value = 1;
export function test() {
  return value;
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
// Basic namespace usage
namespace Test {
  export const value = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Nested namespace
namespace Outer {
  namespace Inner {
    export const value = 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Namespace with interface
namespace Test {
  export interface Config {
    value: string;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Namespace with class
namespace Test {
  export class MyClass {
    constructor() {}
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Namespace with function
namespace Test {
  export function myFunction() {
    return "test";
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Declare namespace (not allowed by default)
declare namespace Test {
  export const value = 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Multiple namespaces
namespace A {
  export const a = 1;
}

namespace B {
  export const b = 2;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
				{
					MessageId: "noNamespace",
				},
			},
		},
		{
			Code: `
// Namespace with complex content
namespace Utils {
  export interface Options {
    debug?: boolean;
    timeout?: number;
  }

  export class Helper {
    static process(options: Options): void {
      // implementation
    }
  }

  export function validate(input: string): boolean {
    return input.length > 0;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNamespace",
				},
			},
		},
	})
}
