package id_denylist

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdDenylistTypeScriptGlobalsAndConstructors(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdDenylistRule,
		[]rule_tester.ValidTestCase{
			// ESTree exposes only identifier/private-identifier method keys to
			// id-denylist. ts-go represents a string key named constructor with
			// ConstructorDeclaration syntax, so its token kind must still win.
			{Code: `class C { "constructor"() {} }`, Options: deny("constructor")},
			{Code: `class C { ["constructor"]() {} }`, Options: deny("constructor")},

			// TypeScript's default lib catalog supplies type-space globals that do
			// not appear in ESLint's ordinary languageOptions.globals table.
			{Code: `type X = Record<string, unknown>; type Y = ReadonlyArray<string>; type Z = Partial<{ a: string }>;`, Options: deny("Record", "ReadonlyArray", "Partial")},
			{Code: `type X = Record<string, unknown>;`, TSConfig: "tsconfig.noLib.json", Options: deny("Record")},
			{Code: `type X = import("pkg").Box<Record<string, unknown>>;`, Options: deny("Record")},
			{Code: `export default Record;`, Options: deny("Record")},
			{Code: `export default (Record);`, Options: deny("Record")},
			{Code: `export default ((Record));`, Options: deny("Record")},
			{Code: `export = (Record);`, Options: deny("Record")},

			// A renamed local export keeps separate local-reference and exported
			// name nodes, so its implicit-global local target remains exempt.
			{Code: `export { Array as allowed };`, Options: deny("Array")},
			{Code: `export type { Record as Safe };`, Options: deny("Record")},

			// Namespace exports are scope-manager references to an existing global.
			{Code: `export as namespace Array;`, FileName: "global-array.d.ts", Options: deny("Array")},
			{Code: `export as namespace MyGlobal;`, FileName: "global-configured.d.ts", Globals: map[string]any{"MyGlobal": "readonly"}, Options: deny("MyGlobal")},
		},
		[]rule_tester.InvalidTestCase{
			// typescript-eslint attaches a simple declaration's direct type
			// annotation to its Identifier range. Rest and binding-pattern names
			// remain bare identifiers.
			{Code: `let Bad!: number;`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}}},
			{Code: `function f(Bad?: number) {}`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 12, EndLine: 1, EndColumn: 24}}},
			{Code: `function f(Bad?) {}`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 12, EndLine: 1, EndColumn: 16}}},
			{Code: `function f(Bad: number = 0) {}`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 12, EndLine: 1, EndColumn: 23}}},
			{Code: `function f(...Bad: number[]) {}`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 15, EndLine: 1, EndColumn: 18}}},
			{Code: `const [Bad]: [number] = value;`, Options: deny("Bad"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restricted", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}}},

			// The scanner starts after modifiers/decorators and skips trivia, so
			// comments cannot steal the constructor diagnostic range.
			{Code: `class C { /* constructor */ constructor() {} }`, Options: deny("constructor"), Errors: []rule_tester.InvalidTestCaseError{restricted("constructor", 1, 29)}},
			{Code: `class C { public /* constructor */ constructor() {} }`, Options: deny("constructor"), Errors: []rule_tester.InvalidTestCaseError{restricted("constructor", 1, 36)}},

			// The shorthand node is also the exported name, which id-denylist
			// checks even when its local-reference role resolves to a global.
			{Code: `export { Array };`, Options: deny("Array"), Errors: []rule_tester.InvalidTestCaseError{restricted("Array", 1, 10)}},
			{Code: `export type { Record };`, Options: deny("Record"), Errors: []rule_tester.InvalidTestCaseError{restricted("Record", 1, 15)}},
			{Code: `export { Array as Array };`, Options: deny("Array"), Errors: []rule_tester.InvalidTestCaseError{restricted("Array", 1, 19)}},

			// ImportType qualifiers name module members, not TypeScript lib globals.
			{Code: `type X = import("pkg").Array;`, Options: deny("Array"), Errors: []rule_tester.InvalidTestCaseError{restricted("Array", 1, 24)}},
			{Code: `type X = import("pkg", { with: { type: "json" } }).Array<Record<string, unknown>>;`, Options: deny("with", "type", "Array", "Record"), Errors: []rule_tester.InvalidTestCaseError{restricted("with", 1, 26), restricted("type", 1, 34), restricted("Array", 1, 52)}},
			{Code: `type X = import("pkg", { /* with */ with: { /* type */ type: "json" } }).Array;`, Options: deny("with", "type", "Array"), Errors: []rule_tester.InvalidTestCaseError{restricted("with", 1, 37), restricted("type", 1, 56), restricted("Array", 1, 74)}},

			// Capabilities are space-sensitive: Record is type-only in esnext.
			{Code: `Record;`, Options: deny("Record"), Errors: []rule_tester.InvalidTestCaseError{restricted("Record", 1, 1)}},
			{Code: `Record;`, FileName: "plain.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Record"), Errors: []rule_tester.InvalidTestCaseError{restricted("Record", 1, 1)}},

			{Code: `export as namespace Unknown;`, FileName: "global-unknown.d.ts", Options: deny("Unknown"), Errors: []rule_tester.InvalidTestCaseError{restricted("Unknown", 1, 21)}},
			{Code: `export as namespace Record;`, FileName: "global-record.d.ts", Options: deny("Record"), Errors: []rule_tester.InvalidTestCaseError{restricted("Record", 1, 21)}},
			{Code: `export as namespace Array;`, FileName: "global-array-off.d.ts", Globals: map[string]any{"Array": "off"}, Options: deny("Array"), Errors: []rule_tester.InvalidTestCaseError{restricted("Array", 1, 21)}},
		},
	)
}

func TestConstructorNameRangeHandlesEscapedKeywordToken(t *testing.T) {
	const source = `class C { constr\u0075ctor() {} }`
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/escaped-constructor.ts",
		Path:     tspath.Path("/escaped-constructor.ts"),
	}, source, core.ScriptKindTS)
	var constructor *ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindConstructor {
			constructor = node
			return true
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if constructor == nil {
		t.Fatal("parser did not retain the escaped constructor declaration")
	}
	nameRange, ok := constructorNameRange(file, constructor)
	if !ok {
		t.Fatal("constructorNameRange did not recognize the escaped keyword token")
	}
	if got, want := source[nameRange.Pos():nameRange.End()], `constr\u0075ctor`; got != want {
		t.Fatalf("constructor range text = %q, want %q", got, want)
	}
}
