package no_namespace

// TestNoNamespaceExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Each group names the branch or input shape it
// protects; the upstream mirror lives in no_namespace_upstream_test.go.

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoNamespaceSourceOnlyRespectsShadowing(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		want int
	}{
		{
			name: "inner type definition wins by name",
			code: `import { createElement } from "react";
function f() { type createElement = {}; createElement("ns:Panel"); }`,
			want: 0,
		},
		{
			name: "parameter default uses the latest body definition",
			code: `const createElement = React.createElement;
function f(x = createElement("ns:Panel")) { const createElement = other; }`,
			want: 0,
		},
		{
			name: "first child definitions follow upstream lookup",
			code: `function child() { const createElement = React.createElement; }
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "cached bindings stay separate across functions",
			code: `import { createElement } from "react";
createElement("ns:One");
function f(createElement) { createElement("ns:Ignored"); }
createElement("ns:Two");`,
			want: 2,
		},
		{
			name: "optional require is not a React definition",
			code: `const { createElement } = require?.("react"); createElement("ns:Panel");`,
			want: 0,
		},
		{
			name: "local createElement shadows imported binding",
			code: `import { createElement } from "react";
function f() {
  const createElement = () => null;
  createElement("ns:Panel");
}`,
			want: 0,
		},
		{
			name: "imported createElement is recognized",
			code: `import { createElement } from "react";
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "default React import is recognized",
			code: `import createElement from "react";
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "namespace React import is recognized",
			code: `import * as createElement from "react";
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "computed React member is recognized",
			code: `const createElement = React["anything"];
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "optional React member is not recognized",
			code: `const createElement = React?.anything;
createElement("ns:Panel");`,
			want: 0,
		},
		{
			name: "wrapped React member is not recognized",
			code: `const createElement = React.anything as any;
createElement("ns:Panel");`,
			want: 0,
		},
		{
			name: "latest variable definition wins over function",
			code: `function createElement() {}
var createElement = React.anything;
createElement("ns:Panel");`,
			want: 1,
		},
		{
			name: "latest unrelated redeclaration suppresses imported binding",
			code: `var createElement = React.createElement;
var createElement = other;
createElement("ns:Panel");`,
			want: 0,
		},
		{
			name: "latest React redeclaration is recognized",
			code: `var createElement = other;
var createElement = React.createElement;
createElement("ns:Panel");`,
			want: 1,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/react.tsx",
				Path:     tspath.Path("/react.tsx"),
			}, testCase.code, core.ScriptKindTSX)
			binder.BindSourceFile(sourceFile)
			refs := rule.NewRefStore(sourceFile, &core.CompilerOptions{}, nil, rule.RefStoreInit{})
			comments := rule.NewCommentStore(sourceFile)
			var diagnostics []rule.RuleDiagnostic
			ctx := (rule.RuleContext{
				SourceFile:     sourceFile,
				Comments:       comments,
				Refs:           refs,
				DisableManager: rule.NewDisableManager(sourceFile, comments),
			}).WithReporter(NoNamespaceRule.Name, rule.SeverityError, func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			})

			listeners := NoNamespaceRule.Run(ctx, nil)
			var visit func(node *ast.Node) bool
			visit = func(node *ast.Node) bool {
				if listener := listeners[node.Kind]; listener != nil {
					listener(node)
				}
				node.ForEachChild(visit)
				return false
			}
			sourceFile.Node.ForEachChild(visit)

			if len(diagnostics) != testCase.want {
				t.Fatalf("got %d diagnostics, want %d: %#v", len(diagnostics), testCase.want, diagnostics)
			}
		})
	}
}

func TestNoNamespaceExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// The file pragma takes precedence over settings for createElement calls.
		{
			Code:     `/* @jsx h */ React.createElement("ns:Panel");`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
		},

		// ---- Dimension 4: paired JSX opening element ----
		{Code: `const x = <Outer><Child /></Outer>;`, Tsx: true},
		// ---- Dimension 4: nested namespace and member-expression siblings ----
		{Code: `const x = <Outer><Inner /><Outer.Inner /></Outer>;`, Tsx: true},
		// ---- Dimension 4: JSX text and fragments are traversal boundaries ----
		{Code: `const x = <><One /><Two>text</Two></>;`, Tsx: true},

		// Locks in upstream isCreateElement() member-access arm: custom pragma.
		{
			Code:     `h.createElement("Panel");`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}},
		},
		// Locks in upstream isCreateElement() member-access arm: wrong pragma.
		{Code: `React.render("ns:Panel");`, Tsx: true},
		{Code: `h.createElement("ns:Panel");`, Tsx: true},
		// Locks in upstream isCreateElement() member-access arm: computed property.
		{Code: `React["createElement"]("ns:Panel");`, Tsx: true},
		// Locks in upstream isCreateElement() member-access arm: empty arguments.
		{Code: `React.createElement();`, Tsx: true},
		// Locks in upstream isCreateElement() member-access arm: no namespace.
		{Code: `React.createElement("ns.Panel");`, Tsx: true},

		// Locks in upstream isCreateElement() bare-callee arm: an unbound
		// createElement is not treated as React.createElement.
		{Code: `createElement("ns:Panel");`, Tsx: true},
		// Locks in upstream isCreateElement() bare-callee arm: unrelated import.
		{Code: `import { createElement } from "other"; createElement("ns:Panel");`, Tsx: true},

		// Dimension 4: parenthesized receiver and argument (ESTree flattens both).
		{Code: `React.createElement(("Panel"));`, Tsx: true},
		{Code: `(React).createElement("Panel");`, Tsx: true},
		// Dimension 4: TypeScript wrapper and template literal are not Literal.
		{Code: `React.createElement("ns:Panel" as string);`, Tsx: true},
		{Code: "React.createElement(`ns:Panel`);", Tsx: true},
		// Dimension 4: optional chains retain the createElement member shape.
		{Code: `React?.createElement("Panel");`, Tsx: true},
		{Code: `(React?.createElement)("Panel");`, Tsx: true},
		// Authored TypeScript wrappers remain visible to ESTree and therefore do
		// not match require()'s string/callee shape.
		{Code: `const { createElement } = require("react" as string); createElement("ns:Panel");`, Tsx: true},
		{Code: `const { createElement } = (require as any)("react"); createElement("ns:Panel");`, Tsx: true},

		// Real-user: eslint-plugin-react#3082 — React's validator passes undefined
		// as a component type; the rule must remain silent and must not crash.
		{Code: `React.createElement(undefined);`, Tsx: true},
		// Real-user: eslint-plugin-react#3082 — React's validator also passes an
		// object component type with fields.
		{Code: `React.createElement({ x: 17 });`, Tsx: true},
		// A type-only definition is still the latest ESLint scope definition.
		{Code: `const createElement = React.createElement; type createElement = {}; createElement("ns:Panel");`, Tsx: true},
		{Code: `const createElement = React.createElement; interface createElement {} createElement("ns:Panel");`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// A JSDoc cast is absent from ESTree, so it is transparent around the
		// first argument and still leaves a Literal to inspect.
		{
			Code:     `React.createElement(/** @type {any} */ ("ns:Panel"));`,
			FileName: "jsdoc-create-element-argument.js",
			TSConfig: "tsconfig.allow-js.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNamespace"}},
		},
		// The same ESTree runtime view applies to pragma initializers.
		{
			Code:     `const createElement = /** @type {any} */ (React.anything); createElement("ns:Panel");`,
			FileName: "jsdoc-create-element-initializer.js",
			TSConfig: "tsconfig.allow-js.json",
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNamespace"}},
		},
		// Upstream uses the latest scope definition even when it is type-only;
		// an import type therefore still follows the import declaration path.
		{
			Code:   `import type { createElement } from "react"; createElement("ns:Panel");`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamespace"}},
		},
		// The file pragma selects the createElement factory used by the rule.
		{
			Code:     `/* @jsx h */ h.createElement("ns:Panel");`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 14, EndLine: 1, EndColumn: 41,
			}},
		},

		// Locks in JSXOpeningElement's namespaced-name arm for a paired element.
		{
			Code: `const x = <ns:Panel></ns:Panel>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 11, EndLine: 1, EndColumn: 21,
			}},
		},
		// Custom pragma's createElement branch reports the complete call.
		{
			Code:     `h.createElement("ns:Panel");`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 28,
			}},
		},
		// Bare createElement imported from React takes the upstream second path.
		{
			Code: `import { createElement } from "react";
createElement("Ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:Panel must not be in a namespace, as React does not support them",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
			}},
		},
		{
			Code: `import createElement from "react";
createElement("Ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:Panel must not be in a namespace, as React does not support them",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
			}},
		},
		{
			Code: `import * as createElement from "react";
createElement("Ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:Panel must not be in a namespace, as React does not support them",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
			}},
		},
		{
			Code: `React[createElement]("ns:Panel");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 33,
			}},
		},
		{
			Code: `const createElement = React["anything"];
createElement("Ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component Ns:Panel must not be in a namespace, as React does not support them",
				Line:      2, Column: 1, EndLine: 2, EndColumn: 26,
			}},
		},
		// Dimension 4: parenthesized argument and receiver stay equivalent to
		// the unwrapped ESTree Literal / MemberExpression shapes.
		{
			Code: `React.createElement(("ns:Panel"));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 34,
			}},
		},
		{
			Code: `(React).createElement("ns:Panel");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 34,
			}},
		},
		// Dimension 4: optional member access keeps the upstream member shape.
		{
			Code: `React?.createElement("ns:Panel");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNamespace",
				Message:   "React component ns:Panel must not be in a namespace, as React does not support them",
				Line:      1, Column: 1, EndLine: 1, EndColumn: 33,
			}},
		},
		// Traversal boundary: both namespace elements are independently reported.
		{
			Code: `const x = <><ns:One /><ns:Two>text</ns:Two></>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: "React component ns:One must not be in a namespace, as React does not support them", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				{MessageId: "noNamespace", Message: "React component ns:Two must not be in a namespace, as React does not support them", Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNamespaceRule, valid, invalid)
}

func TestNoNamespaceDoesNotUseMergedGlobalDeclarationOrder(t *testing.T) {
	root := fixtures.GetRootDir()
	declarationFile := tspath.ResolvePath(root.Dir, "declaration.ts")
	usageFile := tspath.ResolvePath(root.Dir, "usage.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{
		declarationFile: `var createElement = React.createElement;`,
		usageFile: `var createElement = other;
createElement("ns:Panel");`,
	})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("create fixture program: %v", err)
	}

	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program: lintprogram.NewFromCompiler(program),
		Files:   []string{usageFile},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     NoNamespaceRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoNamespaceRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostic count = %d, want 0: %+v", len(diagnostics), diagnostics)
	}
}

func TestNoNamespaceDoesNotUseCrossFileCheckerBinding(t *testing.T) {
	root := fixtures.GetRootDir()
	declarationFile := tspath.ResolvePath(root.Dir, "declaration.ts")
	usageFile := tspath.ResolvePath(root.Dir, "usage.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{
		declarationFile: `var createElement = React.createElement;`,
		usageFile:       `createElement("ns:Panel");`,
	})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("create fixture program: %v", err)
	}

	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program: lintprogram.NewFromCompiler(program),
		Files:   []string{usageFile},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     NoNamespaceRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoNamespaceRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostic count = %d, want 0: %+v", len(diagnostics), diagnostics)
	}
}

// TestNoNamespaceBindingCompatibility is checked against eslint-plugin-react
// 7.37.5 and its current upstream implementation with ESLint 10 and the TS parser.
func TestNoNamespaceBindingCompatibility(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// An inner type declaration wins over an outer value import.
		{Code: `import {createElement} from "react";
function f(){ type createElement = {}; createElement("ns:Panel") }`,
			Tsx: true},
		// Interfaces participate in the same name lookup as runtime declarations.
		{Code: `import {createElement} from "react";
function f(){ interface createElement {} createElement("ns:Panel") }`,
			Tsx: true},
		// Type parameters shadow the imported name.
		{Code: `import {createElement} from "react";
function f<createElement>(){ createElement("ns:Panel") }`,
			Tsx: true},
		// Parameter defaults see the function body declarations in the upstream lookup.
		{Code: `import {createElement} from "react";
function f(x = createElement("ns:Panel")){ const createElement = other; }`,
			Tsx: true},
		// The first child block is searched before the enclosing import.
		{Code: `import {createElement} from "react";
function f(){ { const createElement = other; } createElement("ns:Panel") }`,
			Tsx: true},
		// A class decorator acquires the class environment, including its type parameters.
		{Code: `import {createElement} from "react";
@(createElement("ns:Panel")) class C<createElement> {}`,
			Tsx: true},
		// Computed method keys skip the method environment but still search its first child.
		{Code: `import {createElement} from "react";
class C { [createElement("ns:Panel")](){ const createElement = other; } }`,
			Tsx: true},
		// Parameter decorators acquire the function environment.
		{Code: `import {createElement} from "react";
class C { method(@(createElement("ns:Panel")) x, createElement){} }`,
			Tsx: true},
		// A non-React body declaration also blocks an otherwise unbound default call.
		{Code: `function f(x = createElement("ns:Panel")){ const createElement = other; }`,
			Tsx: true},
		// The latest definition wins even when it has no initializer.
		{Code: `var createElement=React.x; var createElement; createElement("ns:Panel");`,
			Tsx: true},
		// An optional require call is an ESTree ChainExpression, not a require initializer.
		{Code: `const {createElement} = require?.("react");createElement("ns:Panel");`,
			Tsx: true},
		// Parentheses do not turn an optional require call into a plain CallExpression.
		{Code: `const createElement = (require?.("react")).x;createElement("ns:Panel");`,
			Tsx: true},
		// JSDoc casts must preserve the optional-chain boundary.
		{Code: `const {createElement} = /** @type {any} */ (require?.("react")); createElement("ns:Panel");`,
			FileName: "compatibility-16.js",
			TSConfig: "tsconfig.allow-js.json"},
		// Configured script globals stop lookup before child declarations.
		{Code: `function inner(){const createElement=React.x;} createElement("ns:Panel");`,
			FileName:        "compatibility-17.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Globals:         map[string]any{"createElement": "readonly"}},
		// Inline globals have the same effect as configured globals.
		{Code: `/* global createElement:readonly */function inner(){const createElement=React.x;} createElement("ns:Panel");`,
			FileName:        "compatibility-19.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// A with environment puts this declaration beyond the two-child search limit.
		{Code: `with(obj){function f(){const createElement=React.x;}} createElement("ns:Panel");`,
			FileName:        "compatibility-21.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// Loop initializers search the loop body before outer environments.
		{Code: `function f(){for(let i=createElement("ns:Panel"); i<1; i++){const createElement=other;}}`,
			Tsx: true},
		// Switch discriminants acquire the switch environment.
		{Code: `switch(createElement("ns:Panel")){case 0:const createElement=other;}`,
			Tsx: true},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			Code:   `class C { #createElement; f(React) { React.#createElement("ns:Panel"); } }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamespace", Line: 1, Column: 38, EndLine: 1, EndColumn: 70}},
		},
		{
			Code: `React.createElement("ns:\ud800");`,
			Tsx:  true,
			// Compiler strings preserve a lone surrogate as WTF-8.
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamespace", Message: "React component ns:\xed\xa0\x80 must not be in a namespace, as React does not support them", Line: 1, Column: 1, EndLine: 1, EndColumn: 33}},
		},
		// Only the first child and its first child are searched.
		{Code: `import {createElement} from "react";
function f(){ function child(){ function grandchild(){ const createElement = React.x; } } createElement("ns:Panel") }`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 2, Column: 91, EndLine: 2, EndColumn: 116},
			}},
		// A third child level is skipped, leaving the outer import visible.
		{Code: `import {createElement} from "react";
function f(){ function child(){ function grandchild(){ function ignored(){ const createElement = React.x; } } } createElement("ns:Panel") }`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 2, Column: 113, EndLine: 2, EndColumn: 138},
			}},
		// Later sibling environments are skipped, leaving the outer import visible.
		{Code: `import {createElement} from "react";
function f(){ function first(){} function second(){ const createElement = React.x; } createElement("ns:Panel") }`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 2, Column: 86, EndLine: 2, EndColumn: 111},
			}},
		// An otherwise unbound call can find the first child and grandchild.
		{Code: `function child(){ {const createElement=React.x;} } createElement("ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 1, Column: 52, EndLine: 1, EndColumn: 77},
			}},
		// Module bindings are searched before configured globals.
		{Code: `function inner(){const createElement=React.x;} createElement("ns:Panel");`,
			FileName:        "compatibility-18.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			Globals:         map[string]any{"createElement": "readonly"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 1, Column: 48, EndLine: 1, EndColumn: 73},
			}},
		// A disabled global does not stop child lookup.
		{Code: `function inner(){const createElement=React.x;} createElement("ns:Panel");`,
			FileName:        "compatibility-20.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Globals:         map[string]any{"createElement": "off"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 1, Column: 48, EndLine: 1, EndColumn: 73},
			}},
		// A with statement and its body are two separate child environments.
		{Code: `with(obj){const createElement=React.x;} createElement("ns:Panel");`,
			FileName:        "compatibility-22.js",
			TSConfig:        "tsconfig.allow-js.json",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 1, Column: 41, EndLine: 1, EndColumn: 66},
			}},
		// A dotted namespace introduces one environment for all name segments.
		{Code: `namespace N.M.O {const createElement=React.x;} createElement("ns:Panel");`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNamespace", Message: `React component ns:Panel must not be in a namespace, as React does not support them`, Line: 1, Column: 48, EndLine: 1, EndColumn: 73},
			}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNamespaceRule, valid, invalid)
}
