package jsx_tag_spacing

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

func TestJsxTagSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxTagSpacingRule, []rule_tester.ValidTestCase{
		// All checks can be disabled together.
		{Code: "< App / >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// A comment-only gap satisfies never.
		{Code: "<App/* comment */></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Line comments make the closing bracket a new-line boundary.
		{Code: "<App // comment\n/>", Tsx: true, Options: []any{}},
		// Fragments are not tag-spacing targets.
		{Code: "<>text</>", Tsx: true, Options: []any{map[string]any{"closingSlash": "always", "afterOpening": "always", "beforeClosing": "always"}}},
		// A multiline string token is exempt from ordinary beforeClosing checks.
		{Code: "<App value=\"first\nsecond\"></App >", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}},
		// A newline between the final attribute and bracket is allowed with never.
		{Code: "<App value=\"first\nsecond\"\n></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "never"}}},
		// Unicode line separators satisfy allow-multiline.
		{Code: "<\u2028App />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow-multiline", "beforeClosing": "allow"}}},
		// A Unicode paragraph separator is a line break before self-closing.
		{Code: "<App\u2029/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}},
		// Unicode line breaks also satisfy checks that explicitly require spacing.
		{Code: "<\u2028App />", Tsx: true, Options: map[string]any{"afterOpening": "always"}},
		{Code: "<App /\u2029>", Tsx: true, Options: map[string]any{"closingSlash": "always"}},
		// A multiline string is a single token even with a `this` member tag name.
		{Code: "<this.Button value=\"first\nsecond\" ></this.Button>", Tsx: true, Options: map[string]any{"beforeClosing": "never"}},
	}, []rule_tester.InvalidTestCase{
		// The scanner skips these line separators without returning newline trivia.
		{Code: "<\u2028App />", Tsx: true, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
		}},
		{Code: "<App /\u2029>", Tsx: true, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 1, Column: 6, EndLine: 2, EndColumn: 2},
		}},
		// JSX-valued attributes do not need braces; both tag boundaries are checked.
		{Code: "<App child=<span/>/>", Tsx: true, Output: []string{"<App child=<span /> />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 19, EndLine: 1, EndColumn: 19},
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 17, EndLine: 1, EndColumn: 17},
		}},
		// An empty option object retains the runtime defaults.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{}}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 5},
		}},
		// Explicit defaults match omitted options.
		{Code: "<App/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "never", "beforeSelfClosing": "always", "afterOpening": "never", "beforeClosing": "allow"}}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 5},
		}},
		// Independent checks preserve all three edits. The native tester uses
		// listener order; ESLint sorts the same diagnostics by source position.
		{Code: "< App/ >", Tsx: true, Options: []any{}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 1, Column: 6, EndLine: 1, EndColumn: 9},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 6, EndLine: 1, EndColumn: 6},
		}},

		// Whitespace inside a comment does not count as spacing.
		{Code: "<App/* comment *//>", Tsx: true, Options: []any{}, Output: []string{"<App/* comment */ />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 18, EndLine: 1, EndColumn: 18},
		}},
		// Real whitespace adjacent to a comment counts and is removed with the gap.
		{Code: "<App /* comment */ />", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "never", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNoSpace", Message: "A space is forbidden before closing bracket", Line: 1, Column: 20, EndLine: 1, EndColumn: 20},
		}},
		// A comment-only gap does not satisfy always.
		{Code: "<App/* comment */></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<App/* comment */ ></App >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 18},
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 24, EndLine: 1, EndColumn: 24},
		}},
		// Dotted and namespaced tags use the complete name span.
		{Code: "<>< UI.Button></ UI.Button>< svg:path/></>", Tsx: true, Options: []any{}, Output: []string{"<><UI.Button></UI.Button><svg:path /></>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 3, EndLine: 1, EndColumn: 5},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 16, EndLine: 1, EndColumn: 18},
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 28, EndLine: 1, EndColumn: 30},
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 38, EndLine: 1, EndColumn: 38},
		}},
		// Namespaced and spread attributes retain their full node spans.
		{Code: "<svg:path xml:lang=\"en\" {...props}/>", Tsx: true, Options: []any{}, Output: []string{"<svg:path xml:lang=\"en\" {...props} />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 35, EndLine: 1, EndColumn: 35},
		}},
		// Multiline attribute contents count for proportional self-closing checks.
		{Code: "<App value={\n  thing\n}/>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "proportional-always", "afterOpening": "allow", "beforeClosing": "allow"}}, Output: []string{"<App value={\n  thing\n}\n/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 2, EndLine: 3, EndColumn: 2},
		}},
		// Ordinary beforeClosing checks use the last token, even in a multiline attribute.
		{Code: "<App value={\n thing\n}></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<App value={\n thing\n} ></App >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 3, Column: 2, EndLine: 3, EndColumn: 2},
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 3, Column: 8, EndLine: 3, EndColumn: 8},
		}},
		// Proportional beforeClosing checks use the whole attribute.
		{Code: "<App value={\n thing\n}></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<App value={\n thing\n}\n></App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 3, Column: 2, EndLine: 3, EndColumn: 2},
		}},
		// Upstream inserts another space on every pass for this single-line
		// proportional case; preserve its behavior through the ten-pass fix limit.
		{Code: "<App ></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<App  ></App>", "<App   ></App>", "<App    ></App>", "<App     ></App>", "<App      ></App>", "<App       ></App>", "<App        ></App>", "<App         ></App>", "<App          ></App>", "<App           ></App>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
		}},
		// Proportional closing tags can require a newline.
		{Code: "<App></\nApp>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<App></\nApp\n>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 2, Column: 4, EndLine: 2, EndColumn: 4},
		}},
		// Multiline member names use start and end lines separately.
		{Code: "<UI.\nButton></UI.Button>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "proportional-always"}}, Output: []string{"<UI.\nButton\n></UI.Button>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedNewline", Message: "A newline is required before closing bracket", Line: 2, Column: 7, EndLine: 2, EndColumn: 7},
		}},
		// CRLF and UTF-16 columns are preserved.
		{Code: "const é = \"😀\";\r\n<É prop=\"😀\"/ >", Tsx: true, Options: []any{}, Output: []string{"const é = \"😀\";\r\n<É prop=\"😀\" />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "selfCloseSlashNoSpace", Message: "Whitespace is forbidden between `/` and `>`; write `/>`", Line: 2, Column: 13, EndLine: 2, EndColumn: 16},
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 2, Column: 13, EndLine: 2, EndColumn: 13},
		}},
		// Non-ASCII whitespace is removed after the opening bracket.
		{Code: "<\u00a0App />", Tsx: true, Options: []any{}, Output: []string{"<App />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "afterOpenNoSpace", Message: "A space is forbidden after opening bracket", Line: 1, Column: 1, EndLine: 1, EndColumn: 3},
		}},
		// Type arguments retain upstream name-to-next-token semantics.
		{Code: "<App<T>/>", Tsx: true, Options: []any{}, Output: []string{"<App <T>/>"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 5, EndLine: 1, EndColumn: 5},
		}},
		// Type arguments do not replace the final attribute.
		{Code: "<App<T> value={(thing as T)}/>", Tsx: true, Options: []any{}, Output: []string{"<App<T> value={(thing as T)} />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 29, EndLine: 1, EndColumn: 29},
		}},
		// Ordinary closing checks use the type-argument closing token.
		{Code: "<App<T>></App>", Tsx: true, Options: []any{map[string]any{"closingSlash": "allow", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "always"}}, Output: []string{"<App<T> ></App >"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 8, EndLine: 1, EndColumn: 8},
			{MessageId: "beforeCloseNeedSpace", Message: "Whitespace is required before closing bracket", Line: 1, Column: 14, EndLine: 1, EndColumn: 14},
		}},
		// Nested JSX, templates, regexps and optional/computed accesses keep token boundaries.
		{Code: "<App child={<UI.Button value={obj?.[key] ?? /x>/.test(`a${value}`)}/>}/>", Tsx: true, Options: []any{}, Output: []string{"<App child={<UI.Button value={obj?.[key] ?? /x>/.test(`a${value}`)} />} />"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 71, EndLine: 1, EndColumn: 71},
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 68, EndLine: 1, EndColumn: 68},
		}},

		// Parenthesized TypeScript expressions do not alter attribute ranges.
		{Code: "(<App value={(thing satisfies T)}/>)", Tsx: true, Options: []any{}, Output: []string{"(<App value={(thing satisfies T)} />)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "beforeSelfCloseNeedSpace", Message: "A space is required before closing bracket", Line: 1, Column: 34, EndLine: 1, EndColumn: 34},
		}},
	})
}

func collectDiagnostics(t *testing.T, code string, options any, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.tsx", Path: "/test.tsx"}, code, core.ScriptKindTSX)
	comments := rule.NewCommentStore(sf)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{SourceFile: sf, Comments: comments, DisableManager: rule.NewDisableManager(sf, comments)}.WithDiagnosticConsumer(JsxTagSpacingRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
	})
	listeners := JsxTagSpacingRule.Run(ctx, rule_tester.ResolveTestCaseOptions(t, &JsxTagSpacingRule, options))
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		node.ForEachChild(visit)
		return false
	}
	visit(sf.AsNode())
	return diagnostics
}

func TestJsxTagSpacingClosingSlashFix(t *testing.T) {
	// These two upstream cases have a supported input but an output that tsgo
	// cannot parse. Verify the original diagnostics and edits without parsing the output.
	for _, test := range []struct {
		code, output string
		start        int
	}{
		{"<App prop=\"foo\"></App>", "<App prop=\"foo\">< /App>", 16},
		{"<Goodbye></Goodbye>", "<Goodbye>< /Goodbye>", 9},
	} {
		diagnostics := collectDiagnostics(t, test.code, map[string]any{
			"closingSlash": "always", "beforeSelfClosing": "allow", "afterOpening": "allow", "beforeClosing": "allow",
		}, rule.EditDemandAll)
		assert.Equal(t, len(diagnostics), 1)
		diagnostic := diagnostics[0]
		assert.Equal(t, diagnostic.Message.Id, "closeSlashNeedSpace")
		assert.Equal(t, diagnostic.Message.Description, "Whitespace is required between `<` and `/`; write `< /`")
		assert.Equal(t, diagnostic.Range, core.NewTextRange(test.start, test.start+2))
		assert.Assert(t, slices.Equal(diagnostic.Fixes(), []rule.RuleFix{{Range: core.NewTextRange(test.start+1, test.start+1), Text: " "}}))
		assert.Assert(t, diagnostic.Suggestions == nil)
		output, _, _ := linter.ApplyRuleFixes(test.code, diagnostics)
		assert.Equal(t, output, test.output)
	}
}

func TestJsxTagSpacingEditDemand(t *testing.T) {
	code := "< App/ >"
	all := collectDiagnostics(t, code, nil, rule.EditDemandAll)
	assert.Equal(t, len(all), 3)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostics := collectDiagnostics(t, code, nil, demand)
		assert.Equal(t, len(diagnostics), len(all))
		for index, diagnostic := range diagnostics {
			assert.DeepEqual(t, diagnostic.Message, all[index].Message)
			assert.Equal(t, diagnostic.Range, all[index].Range)
			assert.Equal(t, diagnostic.Severity, all[index].Severity)
			assert.Assert(t, diagnostic.Suggestions == nil)
			if demand&rule.EditDemandAutofix != 0 {
				assert.Assert(t, slices.Equal(diagnostic.Fixes(), all[index].Fixes()))
			} else {
				assert.Equal(t, len(diagnostic.Fixes()), 0)
			}
		}
	}
}

func TestJsxTagSpacingSchema(t *testing.T) {
	for _, options := range [][]any{
		{"always"},
		{map[string]any{"unknown": "always"}},
		{map[string]any{"closingSlash": "proportional-always"}},
		{map[string]any{"afterOpening": "proportional-always"}},
		{map[string]any{"beforeSelfClosing": "allow-multiline"}},
		{map[string]any{"beforeClosing": "allow-multiline"}},
		{map[string]any{}, map[string]any{}},
	} {
		assert.Assert(t, JsxTagSpacingRule.Schema.Validate(options) != nil, "options: %#v", options)
	}
}
