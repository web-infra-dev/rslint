// TestJsxOneExpressionPerLineExtras locks in branches and edge shapes that
// the upstream test suite does not fully exercise. The upstream migration
// cases live in the sibling jsx_one_expression_per_line_upstream_test.go file.
package jsx_one_expression_per_line

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

func TestJsxOneExpressionPerLineExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxOneExpressionPerLineRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized JSX expression ----
		{Code: "const value = (<App>\n  <Foo />\n</App>);", Tsx: true},
		// ---- Dimension 4: nested element and fragment boundaries ----
		{Code: "const value = <App>\n  <Foo>\n    <Bar />\n  </Foo>\n</App>;", Tsx: true},
		{Code: "const value = <>\n  <Foo />\n  <Bar />\n</>;", Tsx: true},
		// ---- Dimension 4: self-closing child ----
		{Code: "<App>\n  <Foo />\n</App>", Tsx: true},
		// ---- Dimension 4: empty children and whitespace-only text ----
		{Code: `<App></App>`, Tsx: true},
		{Code: "<App>\n  \t\n</App>", Tsx: true},
		// ---- Dimension 4: member and namespaced tag names ----
		{Code: "<App>\n  <Foo.Bar />\n</App>", Tsx: true},
		{Code: "<App>\n  <svg:path />\n</App>", Tsx: true},
		// ---- Dimension 4: multiline attributes do not become children ----
		{Code: "<App\n  foo=\"bar\"\n>\n  <Foo />\n</App>", Tsx: true},
		// ---- Dimension 4: async / TypeScript-containing expression child ----
		{Code: `<App>{(value as string)}</App>`, Tsx: true, Options: map[string]any{"allow": "single-child"}},
		// ---- Real-user: issue #1835, text after a component ----
		{Code: "<div>\n  <MyComponent>\n    a\n  </MyComponent>\n  <MyOther>\n    {a}\n  </MyOther>\n</div>", Tsx: true},
		// ---- Real-user: issue #1893, CRLF line endings ----
		{Code: "<div>\r\n  <Foo />\r\n</div>", Tsx: true},
		// ---- Real-user: issue #2318, Gatsby-style text children ----
		{Code: "<Layout>\n  <p>\n    Welcome to your new Gatsby site.\n  </p>\n  <p>\n    Now go build something great.\n  </p>\n</Layout>", Tsx: true},
		// Locks in upstream handleJSX() arm 2: non-jsx returns when no direct JSX child exists.
		{Code: `<App>text {value}</App>`, Tsx: true, Options: map[string]any{"allow": "non-jsx"}},
	}, []rule_tester.InvalidTestCase{
		{Code: `<App><Foo.Bar /></App>`, Tsx: true, Output: []string{"<App>\n<Foo.Bar />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Foo.Bar` must be placed on a new line"},
		}},
		{Code: `<App><svg:path /></App>`, Tsx: true, Output: []string{"<App>\n<svg:path />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`svg:path` must be placed on a new line"},
		}},
		// Locks in upstream handleJSX() arm 1: empty children return without a report.
		{Code: `<App><Foo />text</App>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 13},
			{MessageId: "moveToNewLine", Message: "`text` must be placed on a new line", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
		}, Output: []string{"<App>\n<Foo />\ntext\n</App>"}},
		// Locks in upstream handleJSX() arm 3: a direct JSX child keeps the rule active in non-jsx mode.
		{Code: `<App>text <Foo /></App>`, Tsx: true, Options: map[string]any{"allow": "non-jsx"}, Output: []string{"<App>\ntext\n<Foo />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`text ` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 11},
			{MessageId: "moveToNewLine", Message: "`Foo` must be placed on a new line", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
		}},
		{Code: `<App>foo {value}</App>`, Tsx: true, Output: []string{"<App>\nfoo\n{value}\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`foo ` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`{value}` must be placed on a new line"},
		}},
		{Code: `<App>{x}a {y}</App>`, Tsx: true, Output: []string{"<App>\n{x}\na \n{' '}\n{y}\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`{x}` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`a ` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`{y}` must be placed on a new line"},
		}},
		{Code: `<A>{x}a </A>`, Tsx: true, Output: []string{"<A>\n{x}\na\n{' '}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`{x}` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`a ` must be placed on a new line"},
		}},
		{Code: `<App>foo <A /> bar <B /></App>`, Tsx: true, Output: []string{"<App>\nfoo\n<A />\n{' '}\nbar\n<B />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`foo ` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`A` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "` bar ` must be placed on a new line"},
			{MessageId: "moveToNewLine", Message: "`B` must be placed on a new line"},
		}},
		{Code: "<App>foo \n</App>", Tsx: true, Output: []string{"<App>\nfoo \n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
		}},
		{Code: "<App>foo \r\n</App>\r\n", Tsx: true, Output: []string{"<App>\nfoo \r\n</App>\r\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
		}},
		{Code: "<App>foo \r</App>", Tsx: true, Output: []string{"<App>\nfoo \r\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
		}},
		// Locks in upstream single-child allow arms: literal allows text, not an expression container.
		{Code: `<App>{value}</App>`, Tsx: true, Options: map[string]any{"allow": "literal"}, Output: []string{"<App>\n{value}\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`{value}` must be placed on a new line", Line: 1, Column: 6, EndLine: 1, EndColumn: 13},
		}},
		// Locks in upstream grouping arm: a multiline text child is represented on both endpoint lines.
		{Code: "<App>\n  foo <Bar />\n</App>", Tsx: true, Output: []string{"<App>\n  foo \n{' '}\n<Bar />\n</App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Message: "`Bar` must be placed on a new line", Line: 2, Column: 7, EndLine: 2, EndColumn: 14},
		}},
		// ---- Dimension 4: a multiline JSX child gets one report per conflicting sibling relation ----
		{Code: "<App>\n  <Foo>\n    text\n  </Foo><Bar />\n</App>", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine", Line: 4, Column: 9, EndLine: 4, EndColumn: 16},
		}, Output: []string{"<App>\n  <Foo>\n    text\n  </Foo>\n<Bar />\n</App>"}},
		// ---- Adjacent-fix parity: eslint-plugin-react only lands every other
		// ---- report per pass when children abut, which decides whether a raw
		// ---- trailing space is trimmed or kept alongside a `{' '}` marker. ----
		{Code: `<A><B />{x}a {y}</A>`, Tsx: true, Output: []string{"<A>\n<B />\n{x}\na\n{y}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B /><C />a {y}</A>`, Tsx: true, Output: []string{"<A>\n<B />\n<C />\na\n{y}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B />{x}a {y}b {z}</A>`, Tsx: true, Output: []string{"<A>\n<B />\n{x}\na\n{y}\nb\n{z}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A>{x}a {y} b {z}</A>`, Tsx: true, Output: []string{"<A>\n{x}\na \n{' '}\n{y}\n{' '}\nb \n{' '}\n{z}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B />{x} a {y}</A>`, Tsx: true, Output: []string{"<A>\n<B />\n{x}\n{' '}\na\n{y}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B />foo <C /> bar <D /></A>`, Tsx: true, Output: []string{"<A>\n<B />\nfoo \n{' '}\n<C />\n{' '}\nbar \n{' '}\n<D />\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B /> a {x}</A>`, Tsx: true, Output: []string{"<A>\n<B />\n{' '}\na \n{' '}\n{x}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A>{x}<B /> {y}</A>`, Tsx: true, Output: []string{"<A>\n{x}\n<B /> \n{' '}\n{y}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A><B />{x}a </A>`, Tsx: true, Output: []string{"<A>\n<B />\n{x}\na\n{' '}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
		{Code: `<A>a {x}b </A>`, Tsx: true, Output: []string{"<A>\na\n{x}\nb\n{' '}\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
			{MessageId: "moveToNewLine"},
		}},
	})
}

func TestJsxOneExpressionPerLineTriviaAndSpacing(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, prefix := range []string{"\n", "\r\n", "\r", "\u2028", "\u2029", "/* comment\n */", "// comment\n"} {
		for _, allow := range []string{"literal", "single-child"} {
			for _, jsx := range []string{"<A>text</A>", "<>text</>"} {
				valid = append(valid, rule_tester.ValidTestCase{
					Code: "const view = () => (" + prefix + jsx + ");", Tsx: true,
					Options: map[string]any{"allow": allow},
				})
			}
		}
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: "const view = () => (\n<A>{value as string}</A>\n);", Tsx: true, Options: map[string]any{"allow": "single-child"}},
		rule_tester.ValidTestCase{Code: "const view = () => (\n<><A /></>\n);", Tsx: true, Options: map[string]any{"allow": "single-child"}},
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxOneExpressionPerLineRule, valid, []rule_tester.InvalidTestCase{
		{Code: "const view = () => (\n<A>{value}</A>\n);", Tsx: true, Options: map[string]any{"allow": "literal"},
			Output: []string{"const view = () => (\n<A>\n{value}\n</A>\n);"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine", Line: 2, Column: 4, EndLine: 2, EndColumn: 11}},
		},
		{Code: "<A\n>text</A>", Tsx: true, Options: map[string]any{"allow": "literal"},
			Output: []string{"<A\n>\ntext\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}},
		},
		{Code: "<A>text</\nA>", Tsx: true, Options: map[string]any{"allow": "single-child"},
			Output: []string{"<A>\ntext\n</\nA>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}},
		},
		{Code: "<A>text\ntext</A>", Tsx: true, Options: map[string]any{"allow": "single-child"},
			Output: []string{"<A>\ntext\ntext\n</A>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}},
		},
		{Code: "<A>{x}text  {y}</A>", Tsx: true,
			Output: []string{"<A>\n{x}\ntext  \n{' '}\n{y}\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}},
		},
		{Code: "<A><B />text  {y}</A>", Tsx: true, Options: map[string]any{"allow": "non-jsx"},
			Output: []string{"<A>\n<B />\ntext  \n{' '}\n{y}\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}},
		},
		{Code: "<A>{x}text\n more\n \t </A>", Tsx: true,
			Output: []string{"<A>\n{x}\ntext\n more\n \t</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}},
		},
		// Upstream's boundary adjustment recognizes LF, not a bare CR.
		{Code: "<A>{x}text\r  {y}</A>", Tsx: true,
			Output: []string{"<A>\n{x}\ntext\r  \n{' '}\n{y}\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}},
		},
		// A suppressed report must not change which following text fixes land.
		{Code: "<A>\n{/* eslint-disable-next-line react/jsx-one-expression-per-line */}\n{x}{(\nx\n)} text  {y}\n</A>", Tsx: true,
			Output: []string{"<A>\n{/* eslint-disable-next-line react/jsx-one-expression-per-line */}\n{x}{(\nx\n)}\n{' '}\ntext\n{y}\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine", Line: 5, Column: 3, EndLine: 5, EndColumn: 10}, {MessageId: "moveToNewLine", Line: 5, Column: 10, EndLine: 5, EndColumn: 13}},
		},
		{Code: "<A>{/* eslint-disable react/jsx-one-expression-per-line */}<B>{x}\n</B> text {y}{/* eslint-enable react/jsx-one-expression-per-line */} text {y}</A>", Tsx: true,
			Output: []string{"<A>\n{/* eslint-disable react/jsx-one-expression-per-line */}<B>{x}\n</B> text {y}{/* eslint-enable react/jsx-one-expression-per-line */}\n{' '}\ntext\n{y}\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}, {MessageId: "moveToNewLine"}},
		},
	})
}

func TestJsxOneExpressionPerLineEditDemand(t *testing.T) {
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/demand.tsx", Path: "/demand.tsx",
		}, "<A>{x}text  {y}</A>", core.ScriptKindTSX)
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{SourceFile: sourceFile, Comments: comments, DisableManager: rule.NewDisableManager(sourceFile, comments)}.WithDiagnosticConsumer(ruleName, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
		})
		listeners := JsxOneExpressionPerLineRule.Run(ctx, nil)
		listeners[ast.KindJsxElement](sourceFile.Statements.Nodes[0].AsExpressionStatement().Expression)
		assert.Equal(t, len(diagnostics), 3)
		for i, expected := range []struct {
			text       string
			start, end int
		}{{"{x}", 3, 6}, {"text  ", 6, 12}, {"{y}", 12, 15}} {
			d := diagnostics[i]
			assert.Equal(t, d.Message.Description, "`"+expected.text+"` must be placed on a new line")
			assert.Equal(t, d.Range, core.NewTextRange(expected.start, expected.end))
			assert.Equal(t, len(d.Fixes()) > 0, demand&rule.EditDemandAutofix != 0)
			assert.Assert(t, d.Suggestions == nil)
		}
	}
}
