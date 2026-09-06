package prefer_called_with

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestPreferCalledWithMessageDataAndEditDemand(t *testing.T) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", Path: "/test.ts"}, `expect(fn).toBeCalled();`, core.ScriptKindTS)
	node := source.Statements.Nodes[0].AsExpressionStatement().Expression
	accessor := node.AsCallExpression().Expression.AsPropertyAccessExpression().Name()
	for _, autofix := range []bool{false, true} {
		testRule := NewRule(Config{
			Name: "test/prefer-called-with", Autofix: autofix,
			Replacements: map[string]string{"toBeCalled": "toBeCalledWith"},
			Prepare: func(rule.RuleContext) Runtime {
				return Runtime{Parse: func(*ast.Node) *ExpectCall {
					return &ExpectCall{Matcher: "toBeCalled", MatcherEntry: testFramework.MemberEntry{Node: accessor}}
				}}
			},
		})
		if err := testRule.Schema.Validate(nil); err != nil {
			t.Fatal(err)
		}
		if err := testRule.Schema.Validate([]any{map[string]any{}}); err == nil {
			t.Fatal("the rule must reject options")
		}
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
			var diagnostics []rule.RuleDiagnostic
			ctx := (rule.RuleContext{SourceFile: source}).WithDiagnosticConsumer(testRule.Name, rule.SeverityWarning, rule.DiagnosticConsumer{
				Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			})
			testRule.Run(ctx, nil)[ast.KindCallExpression](node)
			if len(diagnostics) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
			}
			diagnostic := diagnostics[0]
			wantData := "toBeCalled"
			if autofix {
				wantData = "toBeCalledWith"
			}
			if diagnostic.Message.Data["matcherName"] != wantData || diagnostic.Message.Description != "Prefer toBeCalledWith(/* expected args */)" {
				t.Fatalf("unexpected message: %#v", diagnostic.Message)
			}
			wantFix := autofix && demand&rule.EditDemandAutofix != 0
			if (diagnostic.FixesPtr != nil) != wantFix {
				t.Fatalf("autofix %t, demand %d: unexpected fixes", autofix, demand)
			}
		}
	}
}

func TestPreferCalledWithReportWithoutFix(t *testing.T) {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", Path: "/test.ts"}, `expect(fn)[key]();`, core.ScriptKindTS)
	node := source.Statements.Nodes[0].AsExpressionStatement().Expression
	key := node.AsCallExpression().Expression.AsElementAccessExpression().ArgumentExpression
	for _, accessor := range []*ast.Node{nil, key, node} {
		testRule := NewRule(Config{
			Name: "test/prefer-called-with", Autofix: true,
			Replacements: map[string]string{"toBeCalled": "toBeCalledWith"},
			Prepare: func(rule.RuleContext) Runtime {
				return Runtime{Parse: func(*ast.Node) *ExpectCall {
					return &ExpectCall{Matcher: "toBeCalled", MatcherEntry: testFramework.MemberEntry{Node: accessor}}
				}}
			},
		})
		var diagnostics []rule.RuleDiagnostic
		ctx := (rule.RuleContext{SourceFile: source}).WithReporter(testRule.Name, rule.SeverityError, func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		})
		testRule.Run(ctx, nil)[ast.KindCallExpression](node)
		if len(diagnostics) != 1 {
			t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
		}
		if fixes := diagnostics[0].FixesPtr; fixes != nil && len(*fixes) != 0 {
			t.Fatal("unsafe accessor received a fix")
		}
	}
}
