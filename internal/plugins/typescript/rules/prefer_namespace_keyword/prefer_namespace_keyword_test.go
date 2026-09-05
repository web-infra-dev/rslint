package prefer_namespace_keyword

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNamespaceKeywordRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNamespaceKeywordRule, []rule_tester.ValidTestCase{
		{Code: `declare module 'foo';`},
		{Code: `declare module 'foo' {}`},
		{Code: `declare /* before keyword */ module /* before name */ 'foo' {}`},
		{Code: `namespace foo {}`},
		{Code: `declare namespace foo {}`},
		{Code: `declare global {}`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`namespace foo {}`},
		},
		{
			Code: `declare module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`declare namespace foo {}`},
		},
		{
			Code: `
declare module foo {
  declare module bar {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
				},
				{
					MessageId: "useNamespace",
				},
			},
			Output: []string{`
declare namespace foo {
  declare namespace bar {}
}`},
		},
		{
			Code: `export declare /* before keyword */ module /* before name */ foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`export declare /* before keyword */ namespace /* before name */ foo {}`},
		},
		{
			Code: `module foo.bar {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`namespace foo.bar {}`},
		},
		{
			Code: `/* before declaration */
module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      2,
					Column:    1,
				},
			},
			Output: []string{`/* before declaration */
namespace foo {}`},
		},
	})
}

func TestPreferNamespaceKeywordDefersAutofix(t *testing.T) {
	const source = `export declare /* before keyword */ module foo {}`

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/test.ts",
			Path:     "/test.ts",
		}, source, core.ScriptKindTS)
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(
			PreferNamespaceKeywordRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)
		listener := PreferNamespaceKeywordRule.Run(ctx, nil)[ast.KindModuleDeclaration]
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

	diagnosticsOnly := run(rule.EditDemandNone)
	if len(diagnosticsOnly) != 1 {
		t.Fatalf("diagnostics-only run produced %d diagnostics, want 1", len(diagnosticsOnly))
	}
	if diagnosticsOnly[0].FixesPtr != nil {
		t.Fatal("diagnostics-only run materialized an autofix")
	}

	withAutofix := run(rule.EditDemandAutofix)
	if len(withAutofix) != 1 {
		t.Fatalf("autofix run produced %d diagnostics, want 1", len(withAutofix))
	}
	fixes := withAutofix[0].Fixes()
	if len(fixes) != 1 {
		t.Fatalf("autofix run produced %d fixes, want 1", len(fixes))
	}
	fix := fixes[0]
	if got := source[fix.Range.Pos():fix.Range.End()]; got != "module" {
		t.Fatalf("autofix replaces %q, want module", got)
	}
	if fix.Text != "namespace" {
		t.Fatalf("autofix replacement = %q, want namespace", fix.Text)
	}
}
