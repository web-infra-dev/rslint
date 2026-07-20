package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
)

func TestIsNameShadowedBetweenEnumDeclaration(t *testing.T) {
	source := `enum value {
  A = value.present,
}
value.outside;
`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	shadowed := AccessExpressionObject(findNodeWithText(t, sourceFile, "value.present"))
	if shadowed == nil || !IsNameShadowedBetween(shadowed, sourceFile.AsNode(), "value") {
		t.Fatal("expected enum declaration to shadow the namespace before the source-file boundary")
	}

	outsideRef := AccessExpressionObject(findNodeWithText(t, sourceFile, "value.outside"))
	if outsideRef == nil || IsNameShadowedBetween(outsideRef, sourceFile.AsNode(), "value") {
		t.Fatal("expected reference outside enum scope not to be shadowed")
	}
}

func TestGetConstVariableInitializer(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "const-variable-initializer.ts")
	code := `
const direct = [];
consumeDirect(direct);
consumeParenthesized((direct));
let mutable = {};
consumeMutable(mutable);
var legacy = {};
consumeLegacy(legacy);
const assertedSource = [];
consumeAsserted(assertedSource as unknown[]);
const {destructured} = source;
consumeDestructured(destructured);
{
	const shadowed = {};
	consumeShadowed(shadowed);
}
consumeAfter(after);
const after = new Array();
declare const missing: unknown;
consumeMissing(missing);
declare const duplicate: unknown;
declare const duplicate: unknown;
consumeDuplicate(duplicate);
function callable() {}
consumeCallable(callable);
const parameter = [];
function parameterScope(parameter: unknown) {
	consumeParameter(parameter);
}
const NamedFunction = [];
const functionHolder = function NamedFunction() {
	consumeNamedFunction(NamedFunction);
};
const NamedClass = [];
const classHolder = class NamedClass {
	method() {
		consumeNamedClass(NamedClass);
	}
};
const caught = [];
try {} catch (caught) {
	consumeCaught(caught);
}
`

	fs := NewOverlayVFSForFile(filePath, code)
	program, err := CreateProgram(true, fs, rootDir, "tsconfig.json", CreateCompilerHost(rootDir, fs))
	if err != nil {
		t.Fatalf("CreateProgram() error = %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatal("program did not contain test source file")
	}
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	tests := []struct {
		name     string
		callText string
		want     string
	}{
		{name: "direct const", callText: "consumeDirect(direct)", want: "[]"},
		{name: "parenthesized reference", callText: "consumeParenthesized((direct))", want: "[]"},
		{name: "let declaration", callText: "consumeMutable(mutable)"},
		{name: "var declaration", callText: "consumeLegacy(legacy)"},
		{name: "assertion wrapper is significant", callText: "consumeAsserted(assertedSource as unknown[])"},
		{name: "destructured declaration", callText: "consumeDestructured(destructured)", want: "source"},
		{name: "nested shadowing", callText: "consumeShadowed(shadowed)", want: "{}"},
		{name: "declaration after reference", callText: "consumeAfter(after)", want: "new Array()"},
		{name: "missing initializer", callText: "consumeMissing(missing)"},
		{name: "multiple declarations", callText: "consumeDuplicate(duplicate)"},
		{name: "non-variable declaration", callText: "consumeCallable(callable)"},
		{name: "parameter shadow", callText: "consumeParameter(parameter)"},
		{name: "named function expression shadow", callText: "consumeNamedFunction(NamedFunction)"},
		{name: "named class expression shadow", callText: "consumeNamedClass(NamedClass)"},
		{name: "catch binding shadow", callText: "consumeCaught(caught)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			call := findNodeWithText(t, sourceFile, test.callText)
			if !ast.IsCallExpression(call) || len(call.Arguments()) != 1 {
				t.Fatalf("%q is not a one-argument call", test.callText)
			}
			initializer := GetConstVariableInitializer(call.Arguments()[0], typeChecker)
			if test.want == "" {
				if initializer != nil {
					t.Fatalf("GetConstVariableInitializer() = %q, want nil",
						TrimmedNodeText(sourceFile, initializer))
				}
				if withoutChecker := GetConstVariableInitializer(
					call.Arguments()[0],
					nil,
				); withoutChecker != nil {
					t.Fatalf("GetConstVariableInitializer(nil checker) = %q, want nil",
						TrimmedNodeText(sourceFile, withoutChecker))
				}
				return
			}
			if initializer == nil {
				t.Fatalf("GetConstVariableInitializer() = nil, want %q", test.want)
			}
			if got := TrimmedNodeText(sourceFile, initializer); got != test.want {
				t.Fatalf("GetConstVariableInitializer() = %q, want %q", got, test.want)
			}

			initializerWithoutChecker := GetConstVariableInitializer(call.Arguments()[0], nil)
			if initializerWithoutChecker == nil {
				t.Fatalf("GetConstVariableInitializer(nil checker) = nil, want %q", test.want)
			}
			if got := TrimmedNodeText(sourceFile, initializerWithoutChecker); got != test.want {
				t.Fatalf("GetConstVariableInitializer(nil checker) = %q, want %q",
					got, test.want)
			}
		})
	}
}
