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
		// Test array format options
		{
			Code: `
// Declare namespace with array options format
declare namespace Test {
  export const value = 1;
}
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allowDeclarations": true,
				},
			},
		},
		// Test empty options object
		{
			Code: `
// Regular code with empty options
const value = 1;
      `,
			Options: map[string]interface{}{},
		},
		// Test nil options
		{
			Code: `
// Regular code with nil options
const value = 1;
      `,
			Options: nil,
		},
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
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
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
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test allowDeclarations explicitly set to false
		{
			Code: `
// Declare namespace with allowDeclarations explicitly set to false
declare namespace Test {
  export const value = 1;
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test array options format but allowDeclarations false
		{
			Code: `
// Declare namespace with array options format but allowDeclarations false
declare namespace Test {
  export const value = 1;
}
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allowDeclarations": false,
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test mix of namespace and module declaration
		{
			Code: `
// Mix of namespace and module declaration
namespace Test {
  export const value = 1;
}

declare module "external" {
  export const externalValue = 2;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test mix of namespace and global declaration
		{
			Code: `
// Mix of namespace and global declaration
namespace Test {
  export const value = 1;
}

declare global {
  interface GlobalInterface {
    prop: string;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test deeply nested namespaces
		{
			Code: `
// Deeply nested namespaces
namespace Level1 {
  namespace Level2 {
    namespace Level3 {
      export const value = 1;
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test namespace with type aliases
		{
			Code: `
// Namespace with type aliases
namespace Types {
  export type StringOrNumber = string | number;
  export type Callback<T> = (value: T) => void;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test namespace with enums
		{
			Code: `
// Namespace with enums
namespace Constants {
  export enum Status {
    Active = "active",
    Inactive = "inactive"
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		// Test regular namespace should be reported even when allowDefinitionFiles is true
		{
			Code: `
// Regular namespace should be reported even when allowDefinitionFiles is true
namespace Test {
  export const value = 1;
}
      `,
			Options: map[string]interface{}{
				"allowDefinitionFiles": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
	})
}

// Test options parsing logic
func TestNoNamespaceOptionsParsing(t *testing.T) {
	// Test default options
	opts := defaultNoNamespaceOptions
	if *opts.AllowDeclarations != false {
		t.Errorf("Expected default AllowDeclarations to be false, got %v", *opts.AllowDeclarations)
	}
	if *opts.AllowDefinitionFiles != true {
		t.Errorf("Expected default AllowDefinitionFiles to be true, got %v", *opts.AllowDefinitionFiles)
	}
}

// Test message building
func TestNoNamespaceMessage(t *testing.T) {
	message := buildNoNamespaceMessage()
	if message.Id != "moduleSyntaxIsPreferred" {
		t.Errorf("Expected message ID to be 'moduleSyntaxIsPreferred', got %s", message.Id)
	}
	if message.Description != "Namespace is not allowed." {
		t.Errorf("Expected message description to be 'Namespace is not allowed.', got %s", message.Description)
	}
}
