package jsx_curly_newline

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

func TestJsxCurlyNewlineExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyNewlineRule, []rule_tester.ValidTestCase{
		// JSX spread children are represented as JsxExpression by tsgo but are
		// not JSXExpressionContainers in ESTree, so upstream does not inspect them.
		{Code: "<App>{...items}</App>", Tsx: true, Options: []any{"never"}},
		// A TypeScript wrapper may span lines, so multiline selection follows the
		// expression range rather than just the first and last inner tokens.
		{Code: "<App>{\n(value as string)\n}</App>", Tsx: true, Options: []any{map[string]any{"multiline": "require"}}},
		// Empty expressions use their authored interior range when selecting the
		// singleline or multiline option.
		{Code: "<App>{\n}</App>", Tsx: true, Options: []any{map[string]any{"singleline": "forbid", "multiline": "require"}}},
	}, []rule_tester.InvalidTestCase{
		{Code: "<App>{value}</App>", Tsx: true, Options: []any{map[string]any{"singleline": "require"}}, Output: []string{"<App>{\nvalue\n}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("expectedAfter", "Expected newline after '{'.", 1, 6),
			errorAt("expectedBefore", "Expected newline before '}'.", 1, 12),
		}},
		// Line comments are skipped by ESLint's token APIs but deliberately keep
		// newline-removal diagnostics non-fixable.
		{Code: "<App>{// comment\nvalue}</App>", Tsx: true, Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedAfter", "Unexpected newline after '{'.", 1, 6),
		}},
		{Code: "<App>{value // comment\n}</App>", Tsx: true, Options: []any{"never"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedBefore", "Unexpected newline before '}'.", 2, 1),
		}},
		// Brace-adjacent whitespace is not part of the ESTree expression span.
		{Code: "<App>{\nvalue}</App>", Tsx: true, Options: []any{map[string]any{"singleline": "forbid", "multiline": "require"}}, Output: []string{"<App>{value}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedAfter", "Unexpected newline after '{'.", 1, 6),
		}},
		{Code: "<App>{\n(value)}</App>", Tsx: true, Options: []any{map[string]any{"singleline": "forbid", "multiline": "require"}}, Output: []string{"<App>{(value)}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedAfter", "Unexpected newline after '{'.", 1, 6),
		}},
		{Code: "<App>{\nvalue as string}</App>", Tsx: true, Options: []any{map[string]any{"singleline": "forbid", "multiline": "require"}}, Output: []string{"<App>{value as string}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedAfter", "Unexpected newline after '{'.", 1, 6),
		}},
		// Parser-backed comment ranges leave comment-like text in literals alone,
		// preserving upstream's safe newline-removal fixes.
		{Code: "<App>{'http://host'\n}</App>", Tsx: true, Options: []any{"never"}, Output: []string{"<App>{'http://host'}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedBefore", "Unexpected newline before '}'.", 2, 1),
		}},
		{Code: "<App>{/x*/\n}</App>", Tsx: true, Options: []any{"never"}, Output: []string{"<App>{/x*/}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("unexpectedBefore", "Unexpected newline before '}'.", 2, 1),
		}},
		{Code: "<App>{// comment\u2028value}</App>", Tsx: true, Output: []string{"<App>{// comment\u2028value\n}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("expectedBefore", "Expected newline before '}'.", 2, 6),
		}},
		{Code: "<App>{// comment\u2029value}</App>", Tsx: true, Output: []string{"<App>{// comment\u2029value\n}</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			errorAt("expectedBefore", "Expected newline before '}'.", 2, 6),
		}},
	})
}

func TestJsxCurlyNewlineEditDemand(t *testing.T) {
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/demand.tsx", Path: "/demand.tsx"}, "<App>{value}</App>", core.ScriptKindTSX)
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: sourceFile, Comments: comments, DisableManager: rule.NewDisableManager(sourceFile, comments)}.WithDiagnosticConsumer("react/jsx-curly-newline", rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		})
		root := sourceFile.Statements.Nodes[0].AsExpressionStatement().Expression
		child := reactutil.GetJsxChildren(root)[0]
		listeners := JsxCurlyNewlineRule.Run(ctx, []any{map[string]any{"singleline": "require"}})
		listeners[ast.KindJsxExpression](child)

		assert.Equal(t, len(diagnostics), 2)
		for index, expected := range []struct {
			id     string
			range_ core.TextRange
		}{{"expectedAfter", core.NewTextRange(5, 6)}, {"expectedBefore", core.NewTextRange(11, 12)}} {
			diagnostic := diagnostics[index]
			assert.Equal(t, diagnostic.Message.Id, expected.id)
			assert.Equal(t, diagnostic.Range, expected.range_)
			assert.Equal(t, len(diagnostic.Fixes()) > 0, demand&rule.EditDemandAutofix != 0)
			assert.Assert(t, diagnostic.Suggestions == nil)
		}
	}
}
