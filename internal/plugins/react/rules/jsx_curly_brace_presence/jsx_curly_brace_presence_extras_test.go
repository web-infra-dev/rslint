// This file contains rslint-specific regressions and edge cases. Migrated
// upstream coverage remains in jsx_curly_brace_presence_test.go.
package jsx_curly_brace_presence

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestJsxCurlyBracePresenceSourceTextEdges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyBracePresenceRule, []rule_tester.ValidTestCase{
		// JavaScript /^\s|\s$/ protects whitespace whose removal from a
		// multiline JSX expression would change the rendered child text.
		{Code: "<div>\n{\"\u00a0hello\"}\n</div>", Tsx: true},
		{Code: "<div>\n{\"hello\u00a0\"}\n</div>", Tsx: true},
		{Code: "<div>\n{\"" + "\ufeff" + "hello\"}\n</div>", Tsx: true},
		{Code: "<div>\n{\"hello" + "\ufeff" + "\"}\n</div>", Tsx: true},
		{Code: "<div>\n{\"\vhello\"}\n</div>", Tsx: true},
		{Code: "<div>\n{\"hello\v\"}\n</div>", Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// A tsgo StringLiteral Pos includes leading trivia. Fixes must slice
		// from the token start so trivia never becomes attribute content.
		{
			Code:   `<X title={ 'hello' }/>`,
			Tsx:    true,
			Output: []string{`<X title="hello"/>`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:   `<X title={	"hello" }/>`,
			Tsx:    true,
			Output: []string{`<X title="hello"/>`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:   "<X title={\n  'hello'\n}/>",
			Tsx:    true,
			Output: []string{`<X title="hello"/>`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    `<X title= "hello"/>`,
			Tsx:     true,
			Options: "always",
			Output:  []string{`<X title= {"hello"}/>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCurly"}},
		},
		{
			Code:    `<X title=	'hello'/>`,
			Tsx:     true,
			Options: "always",
			Output:  []string{`<X title=	{"hello"}/>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCurly"}},
		},
		{
			Code:    "<X title=\n  \"hello\"/>",
			Tsx:     true,
			Options: "always",
			Output:  []string{"<X title=\n  {\"hello\"}/>"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCurly"}},
		},
	})
}

func TestJsxCurlyBracePresenceEditDemand(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		code       string
		options    any
		wantOutput string
	}{
		{
			name:       "remove braces after leading trivia",
			code:       `<X title={ 'hello' }/>`,
			wantOutput: `<X title="hello"/>`,
		},
		{
			name:       "add braces after leading trivia",
			code:       `<X title= "hello"/>`,
			options:    "always",
			wantOutput: `<X title= {"hello"}/>`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/edit-demand.tsx",
				Path:     "/edit-demand.tsx",
			}, testCase.code, core.ScriptKindTSX)

			run := func(demand rule.EditDemand) rule.RuleDiagnostic {
				t.Helper()

				comments := rule.NewCommentStore(sourceFile)
				var diagnostics []rule.RuleDiagnostic
				ctx := rule.RuleContext{
					SourceFile:     sourceFile,
					Comments:       comments,
					DisableManager: rule.NewDisableManager(sourceFile, comments),
				}.WithDiagnosticConsumer(JsxCurlyBracePresenceRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(diagnostic rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, diagnostic)
					},
				})
				listeners := JsxCurlyBracePresenceRule.Run(ctx, rule_tester.ResolveTestCaseOptions(t, &JsxCurlyBracePresenceRule, testCase.options))
				utils.VisitDescendants(sourceFile.AsNode(), func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return true
				})
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
				}
				return diagnostics[0]
			}

			diagnosticsOnly := run(rule.EditDemandNone)
			autofixOnly := run(rule.EditDemandAutofix)
			suggestionOnly := run(rule.EditDemandSuggestion)
			allEdits := run(rule.EditDemandAll)

			for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
				rule.EditDemandAutofix:    autofixOnly,
				rule.EditDemandSuggestion: suggestionOnly,
				rule.EditDemandAll:        allEdits,
			} {
				if diagnostic.Range != diagnosticsOnly.Range || diagnostic.Message.Id != diagnosticsOnly.Message.Id ||
					diagnostic.Message.Description != diagnosticsOnly.Message.Description || diagnostic.RuleName != diagnosticsOnly.RuleName {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
			}
			if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
				t.Fatal("non-autofix demand materialized fixes")
			}
			if autofixOnly.FixesPtr == nil || allEdits.FixesPtr == nil {
				t.Fatal("autofix demand did not materialize fixes")
			}
			for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
				if diagnostic.Suggestions != nil {
					t.Fatal("autofix-only rule materialized suggestions")
				}
			}
			output, _, fixed := linter.ApplyRuleFixes(testCase.code, []rule.RuleDiagnostic{allEdits})
			if !fixed || output != testCase.wantOutput {
				t.Errorf("fixed output = %q, %v; want %q, true", output, fixed, testCase.wantOutput)
			}
		})
	}
}
