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
	filePath := tspath.ResolvePath(rootDir.Dir, "const-variable-initializer.ts")
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

	fs := NewOverlayVFS(rootDir.FS, map[string]string{filePath: code})
	program, err := CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", CreateCompilerHost(rootDir.Dir, fs))
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

// TestShadowingScopeModels locks in the three scope models the shadowing
// helpers implement, using `Target()` as the single reference in each snippet:
//
//   - IsShadowed follows runtime lexical semantics, and additionally treats a
//     namespace body as a scope of its own.
//   - IsShadowedFromParameterInitializer adds the declarations scope-manager
//     puts in a function scope but runtime semantics keep out of a parameter
//     environment.
//   - HasEnclosingTypeParameter adds the type parameters scope-manager keeps
//     in the lexical chain.
//
// A parameter decorator cuts across all three: a reference sitting directly in
// one still acquires the decorated function's scope, but scope-manager attaches
// a scope created inside the decorator to the enclosing class instead, so the
// decorated function drops out of the chain entirely.
func TestShadowingScopeModels(t *testing.T) {
	tests := []struct {
		name                   string
		code                   string
		shadowed               bool
		fromParameterInit      bool
		enclosingTypeParameter bool
	}{
		// A namespace body is a scope of its own.
		{
			name:     "namespace const",
			code:     `namespace N { const Target = f; Target(); }`,
			shadowed: true,
		},
		{
			name:     "namespace var hoists to the module block",
			code:     `namespace N { var Target = 1; { Target(); } }`,
			shadowed: true,
		},
		{
			name:     "namespace import equals",
			code:     `namespace N { import Target = globalThis.x; Target(); }`,
			shadowed: true,
		},
		{
			name: "namespace declaration does not reach the outer scope",
			code: `namespace N { const Target = f; } Target();`,
		},

		// A body declaration doesn't shadow a default value at runtime, but
		// scope-manager puts both in one function scope.
		{
			name:              "parameter default, body var",
			code:              `function f(a = Target()) { var Target; }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body let",
			code:              `function f(a = Target()) { let Target; }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, var hoisted from a nested block",
			code:              `function f(a = Target()) { if (b) { var Target; } }`,
			fromParameterInit: true,
		},
		{
			name: "parameter default, let in a nested block stays there",
			code: `function f(a = Target()) { { let Target; } }`,
		},
		{
			name:              "parameter default, reference nested in an arrow",
			code:              `function f(a = () => Target()) { var Target; }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body type alias",
			code:              `function f(a = Target()) { type Target = {}; }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body interface",
			code:              `function f(a = Target()) { interface Target {} }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body import equals",
			code:              `function f(a = Target()) { import Target = require("x"); }`,
			fromParameterInit: true,
		},

		// A reference directly in a parameter decorator acquires the decorated
		// function's scope.
		{
			name:              "parameter decorator, body var",
			code:              `class C { m(@dec(Target()) x: number) { var Target; } }`,
			fromParameterInit: true,
		},
		{
			name:     "parameter decorator, sibling parameter",
			code:     `class C { m(@dec(Target()) x: number, Target: any) { } }`,
			shadowed: true,
		},
		{
			name:                   "parameter decorator, method type parameter",
			code:                   `class C { m<Target>(@dec(Target()) x: number) { } }`,
			enclosingTypeParameter: true,
		},

		// A scope created inside a parameter decorator belongs to the enclosing
		// class, so the decorated function drops out of the chain.
		{
			name: "nested scope in a parameter decorator, body var",
			code: `class C { m(@dec(() => Target()) x: number) { var Target; } }`,
		},
		{
			name: "nested scope in a parameter decorator, sibling parameter",
			code: `class C { m(@dec(() => Target()) x: number, Target: any) { } }`,
		},
		{
			name: "nested scope in a parameter decorator, method type parameter",
			code: `class C { m<Target>(@dec(() => Target()) x: number) { } }`,
		},
		{
			name:     "nested scope in a parameter decorator, class name",
			code:     `class Target { m(@dec(() => Target()) x: number) { } }`,
			shadowed: true,
		},
		{
			name:                   "nested scope in a parameter decorator, class type parameter",
			code:                   `class C<Target> { m(@dec(() => Target()) x: number) { } }`,
			enclosingTypeParameter: true,
		},

		// A member's decorators and computed name belong to the enclosing class
		// or object literal, so the member's own scope never reaches them.
		{
			name: "method decorator, body var",
			code: `class C { @dec(Target()) m() { var Target; } }`,
		},
		{
			name: "method decorator, parameter",
			code: `class C { @dec(Target()) m(Target: any) { } }`,
		},
		{
			name: "method decorator, method type parameter",
			code: `class C { @dec(Target()) m<Target>() { } }`,
		},
		{
			name:                   "method decorator, class type parameter",
			code:                   `class C<Target> { @dec(Target()) m() { } }`,
			enclosingTypeParameter: true,
		},
		{
			name: "accessor decorator, body var",
			code: `class C { @dec(Target()) get m() { var Target; return 1; } }`,
		},
		{
			name: "computed method name, body var",
			code: `class C { [Target()]() { var Target; } }`,
		},
		{
			name: "computed method name, parameter",
			code: `class C { [Target()](Target: any) { } }`,
		},
		{
			name: "computed method name, method type parameter",
			code: `class C { [Target()]<Target>() { } }`,
		},
		{
			name:                   "computed method name, class type parameter",
			code:                   `class C<Target> { [Target()]() { } }`,
			enclosingTypeParameter: true,
		},
		{
			name:     "computed method name, class name",
			code:     `class Target { [Target()]() { } }`,
			shadowed: true,
		},
		{
			name: "computed object method name, body var",
			code: `const o = { [Target()]() { var Target; } };`,
		},

		// A function declaration with no block to hold it is defined in the
		// innermost scope that already exists at its position.
		{
			name:              "parameter default, body function in an if branch",
			code:              `function f(a = Target()) { if (b) function Target() {} }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body function in an else branch",
			code:              `function f(a = Target()) { if (b) ; else function Target() {} }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body function in a labelled statement",
			code:              `function f(a = Target()) { lbl: function Target() {} }`,
			fromParameterInit: true,
		},
		{
			name:              "parameter default, body function in an unbraced loop",
			code:              `function f(a = Target()) { while (b) function Target() {} }`,
			fromParameterInit: true,
		},
		{
			name: "parameter default, body function in a nested block stays there",
			code: `function f(a = Target()) { { function Target() {} } }`,
		},
		{
			name: "parameter default, body function in a let-scoped loop stays there",
			code: `function f(a = Target()) { for (let i; ;) function Target() {} }`,
		},
		{
			name:     "function in an if branch reaches the function scope",
			code:     `function f() { if (b) function Target() {} Target(); }`,
			shadowed: true,
		},
		{
			name:     "function in an if branch reaches the source file scope",
			code:     `if (b) function Target() {} Target();`,
			shadowed: true,
		},
		{
			name:     "function in an if branch reaches the namespace scope",
			code:     `namespace N { if (b) function Target() {} Target(); }`,
			shadowed: true,
		},
		{
			name:     "function in an if branch reaches the switch case scope",
			code:     `switch (a) { case 1: if (b) function Target() {} case 2: Target(); }`,
			shadowed: true,
		},
		{
			name: "function in a nested block stays there",
			code: `function f() { { function Target() {} } Target(); }`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, test.code, core.ScriptKindTS)

			callee := findNodeWithText(t, sourceFile, "Target()").Expression()
			if got := IsShadowed(callee, "Target"); got != test.shadowed {
				t.Errorf("IsShadowed() = %v, want %v", got, test.shadowed)
			}
			if got := IsShadowedFromParameterInitializer(callee, "Target"); got != test.fromParameterInit {
				t.Errorf("IsShadowedFromParameterInitializer() = %v, want %v", got, test.fromParameterInit)
			}
			if got := HasEnclosingTypeParameter(callee, "Target"); got != test.enclosingTypeParameter {
				t.Errorf("HasEnclosingTypeParameter() = %v, want %v", got, test.enclosingTypeParameter)
			}
		})
	}
}
