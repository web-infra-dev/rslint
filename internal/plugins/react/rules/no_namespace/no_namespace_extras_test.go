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
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamespaceSourceOnlyRespectsShadowing(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		want int
	}{
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

		// Real-user: eslint-plugin-react#3082 — React's validator passes undefined
		// as a component type; the rule must remain silent and must not crash.
		{Code: `React.createElement(undefined);`, Tsx: true},
		// Real-user: eslint-plugin-react#3082 — React's validator also passes an
		// object component type with fields.
		{Code: `React.createElement({ x: 17 });`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
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
