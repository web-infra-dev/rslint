package no_namespace

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
		{Code: `module "foo" {}`},
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
		{
			Code: `
declare namespace Outer {
  namespace Inner {}
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": true,
			},
		},
		{
			Code: `
declare global {
  namespace Test {}
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": true,
			},
		},
		{
			Code: `
declare module Test {
  namespace Inner {}
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": true,
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
			Code: `namespace Test {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Message:   "ES2015 module syntax is preferred over namespaces.",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `module Test {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		{
			Code: `module Test {}`,
			Options: map[string]interface{}{
				"allowDeclarations": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		{
			Code: `namespace Foo.Bar {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
				},
			},
		},
		{
			Code: `
namespace Foo.Bar {
  namespace Baz.Bas {}
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
		{
			Code: `
namespace Outer {
  declare namespace Inner {
    namespace Nested {}
  }
}
      `,
			Options: map[string]interface{}{
				"allowDeclarations": true,
			},
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

func TestNoNamespaceDefinitionFileOptions(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		source   string
		options  []any
		want     int
	}{
		{name: "d.ts default", fileName: "/file.d.ts", source: `namespace Test {}`},
		{name: "d.mts default", fileName: "/file.d.mts", source: `namespace Test {}`},
		{name: "d.cts default", fileName: "/file.d.cts", source: `namespace Test {}`},
		{
			name:     "definition files disabled",
			fileName: "/file.d.ts",
			source:   `namespace Test {}`,
			options:  []any{map[string]interface{}{"allowDefinitionFiles": false}},
			want:     1,
		},
		{
			name:     "implicit ambient is not explicit declare",
			fileName: "/file.d.ts",
			source:   `namespace Test {}`,
			options: []any{map[string]interface{}{
				"allowDeclarations":    true,
				"allowDefinitionFiles": false,
			}},
			want: 1,
		},
		{
			name:     "explicit declare remains allowed",
			fileName: "/file.d.ts",
			source:   `declare namespace Test {}`,
			options: []any{map[string]interface{}{
				"allowDeclarations":    true,
				"allowDefinitionFiles": false,
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := len(runNoNamespaceRule(test.fileName, test.source, test.options)); got != test.want {
				t.Fatalf("got %d diagnostics, want %d", got, test.want)
			}
		})
	}
}

func TestNoNamespaceReportRangeMatchesTrimmedNode(t *testing.T) {
	sources := []string{
		`namespace Test {}`,
		"\n// leading comment\nexport namespace Test {}\n",
		"\ufeff/* leading block */\nmodule Test {}",
		"/** namespace docs */\nnamespace Outer.Inner {}",
	}

	for _, source := range sources {
		diagnostics := runNoNamespaceRule("/file.ts", source, nil)
		if len(diagnostics) != 1 {
			t.Fatalf("source %q: got %d diagnostics, want 1", source, len(diagnostics))
		}

		sourceFile, ok := diagnostics[0].SourceFile.(*ast.SourceFile)
		if !ok {
			t.Fatalf("source %q: unexpected source file type %T", source, diagnostics[0].SourceFile)
		}
		var declaration *ast.Node
		sourceFile.AsNode().ForEachChild(func(node *ast.Node) bool {
			if node.Kind == ast.KindModuleDeclaration {
				declaration = node
				return true
			}
			return false
		})
		if declaration == nil {
			t.Fatalf("source %q: module declaration not found", source)
		}

		got := diagnostics[0].Range
		want := utils.TrimNodeTextRange(sourceFile, declaration)
		if got.Pos() != want.Pos() || got.End() != want.End() {
			t.Fatalf(
				"source %q: report range [%d,%d), want trimmed node range [%d,%d)",
				source,
				got.Pos(),
				got.End(),
				want.Pos(),
				want.End(),
			)
		}
	}
}

func TestNoNamespaceDisableDirective(t *testing.T) {
	source := "// rslint-disable-next-line test\nnamespace Test {}"
	if diagnostics := runNoNamespaceRule("/file.ts", source, nil); len(diagnostics) != 0 {
		t.Fatalf("disabled namespace produced %d diagnostics", len(diagnostics))
	}
}

func runNoNamespaceRule(fileName string, source string, options []any) []rule.RuleDiagnostic {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
	}, source, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithReporter("test", rule.SeverityError, func(diagnostic rule.RuleDiagnostic) {
		diagnostics = append(diagnostics, diagnostic)
	})

	listener := NoNamespaceRule.Run(ctx, options)[ast.KindModuleDeclaration]
	if listener == nil {
		return diagnostics
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindModuleDeclaration {
			listener(node)
		}
		node.ForEachChild(visit)
		return false
	}
	visit(sourceFile.AsNode())
	return diagnostics
}

// Test options parsing logic
func TestNoNamespaceOptionsParsing(t *testing.T) {
	// Test default options
	opts := defaultNoNamespaceOptions
	if opts.AllowDeclarations {
		t.Errorf("Expected default AllowDeclarations to be false, got %v", opts.AllowDeclarations)
	}
	if !opts.AllowDefinitionFiles {
		t.Errorf("Expected default AllowDefinitionFiles to be true, got %v", opts.AllowDefinitionFiles)
	}
}

// Test message building
func TestNoNamespaceMessage(t *testing.T) {
	message := noNamespaceMessage
	if message.Id != "moduleSyntaxIsPreferred" {
		t.Errorf("Expected message ID to be 'moduleSyntaxIsPreferred', got %s", message.Id)
	}
	if message.Description != "ES2015 module syntax is preferred over namespaces." {
		t.Errorf("Expected upstream message, got %s", message.Description)
	}
}
