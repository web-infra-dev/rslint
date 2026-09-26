package jsx_newline

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

func TestJsxNewlineEditDemand(t *testing.T) {
	for _, test := range []struct {
		code    string
		options []any
		message rule.RuleMessage
		range_  core.TextRange
		fix     rule.RuleFix
	}{
		{
			code:    "<><A/>\n<B/></>",
			message: rule.RuleMessage{Id: "require", Description: "JSX element should start in a new line"},
			range_:  core.NewTextRange(7, 11),
			fix:     rule.RuleFix{Range: core.NewTextRange(6, 7), Text: "\n\n"},
		},
		{
			code:    "<><A/>\n\n<B/></>",
			options: []any{map[string]any{"prevent": true}},
			message: rule.RuleMessage{Id: "prevent", Description: "JSX element should not start in a new line"},
			range_:  core.NewTextRange(8, 12),
			fix:     rule.RuleFix{Range: core.NewTextRange(6, 8), Text: "\n"},
		},
		{
			code:    "<><A/>\n<B\n/></>",
			options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			message: rule.RuleMessage{Id: "allowMultilines", Description: "Multiline JSX elements should start in a new line"},
			range_:  core.NewTextRange(7, 12),
			fix:     rule.RuleFix{Range: core.NewTextRange(6, 7), Text: "\n\n"},
		},
		{
			code:    "<><A/> <B/></>",
			message: rule.RuleMessage{Id: "require", Description: "JSX element should start in a new line"},
			range_:  core.NewTextRange(7, 11),
			fix:     rule.RuleFix{Range: core.NewTextRange(6, 7), Text: " "},
		},
	} {
		for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/demand.tsx", Path: "/demand.tsx"}, test.code, core.ScriptKindTSX)
			comments := rule.NewCommentStore(sourceFile)
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{SourceFile: sourceFile, Comments: comments, DisableManager: rule.NewDisableManager(sourceFile, comments)}.WithDiagnosticConsumer("react/jsx-newline", rule.SeverityError, rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			})
			root := sourceFile.Statements.Nodes[0].AsExpressionStatement().Expression
			options := rule_tester.ResolveTestCaseOptions(t, &JsxNewlineRule, test.options)
			JsxNewlineRule.Run(ctx, options)[ast.KindJsxFragment](root)
			assert.Equal(t, len(diagnostics), 1)
			diagnostic := diagnostics[0]
			assert.DeepEqual(t, diagnostic.Message, test.message)
			assert.Equal(t, diagnostic.Range, test.range_)
			if demand&rule.EditDemandAutofix != 0 {
				assert.Equal(t, len(diagnostic.Fixes()), 1)
				assert.Equal(t, diagnostic.Fixes()[0], test.fix)
			} else {
				assert.Equal(t, len(diagnostic.Fixes()), 0)
			}
			assert.Assert(t, diagnostic.Suggestions == nil)
		}
	}
}

func TestJsxNewlineSchema(t *testing.T) {
	for _, options := range [][]any{
		{map[string]any{"prevent": "true"}},
		{map[string]any{"allowMultilines": "true"}},
		{map[string]any{"unknown": true}},
		{map[string]any{}, map[string]any{}},
		{map[string]any{"allowMultilines": true}},
		{map[string]any{"prevent": false, "allowMultilines": true}},
	} {
		assert.Assert(t, JsxNewlineRule.Schema.Validate(options) != nil, "options must be rejected: %#v", options)
	}
}

func TestJsxNewlineExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNewlineRule, []rule_tester.ValidTestCase{
		// No text sibling means adjacent elements and expressions are ignored.
		{Code: "<><A/><B/>{value}<C/></>", Tsx: true},
		// Only one child, leading whitespace, and trailing whitespace are ignored.
		{Code: `<>

<A/>

</>`, Tsx: true,
		},
		// JSX attribute expressions are not children.
		{Code: "<A value={<B/>} other={value}/>", Tsx: true},
		// A fragment is not checked as the left sibling.
		{Code: `<><><A/></>
<B/></>`, Tsx: true,
		},
		// A spread child is not checked as the left sibling.
		{Code: `<>{...items}
<B/></>`, Tsx: true,
		},
		// An immediately braced block comment is skipped, even before an expression.
		{Code: `<>{/*comment*/ value}
<B/></>`, Tsx: true,
		},
		// Trailing comments have no next element (upstream issue 3573).
		{Code: `<>
<A/>
{/*comment*/}
</>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
		},
		// prevent alone accepts a space between one-line children.
		{Code: "<><A/> <B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": false}},
		},
		// Decoded line-feed entities count as a blank line.
		{Code: "<><A/>&#10;&#10;<B/></>", Tsx: true},
		// Hexadecimal and decimal entities can form the same blank line.
		{Code: "<><A/>&#xA;\t&#10;<B/></>", Tsx: true},
		// Supplementary characters do not count as blank lines in prevent mode.
		{Code: "<><A/>&#65546;&#65546;<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
		// Leading zeros do not prevent numeric references from decoding to LF.
		{Code: "<><A/>&#00000000010;&#x000000000A;<B/></>", Tsx: true},
		// A blank line can occur after non-whitespace text in the same sibling.
		{Code: "<><A/>\ntext\n\n<B/></>", Tsx: true},
		// JavaScript whitespace includes the byte order mark and non-breaking spaces.
		{Code: "<><A/>\n\ufeff \n<B/></>", Tsx: true},
		// Decoded entities can supply the whitespace between line feeds.
		{Code: `<><A/>
&nbsp;
<B/></>`, Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// U+1000A is a supplementary character, not a line feed. These entity
		// cases follow JSX semantics and upstream with @typescript-eslint/parser.
		{Code: "<><A/>&#65546;&#65546;<B/></>", Tsx: true,
			Output: slices.Repeat([]string{"<><A/>&#65546;&#65546;<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
			},
		},
		// Hexadecimal references preserve the same character in multiline mode.
		{Code: "<><A/>&#x1000A;&#x1000A;<B\n/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output:  slices.Repeat([]string{"<><A/>&#x1000A;&#x1000A;<B\n/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 1, Column: 25, EndLine: 2, EndColumn: 3},
			},
		},
		// U+10020 is not whitespace; fixes preserve its original reference.
		{Code: "<><A/>\n&#x10020;\n<B/></>", Tsx: true,
			Output: []string{"<><A/>\n&#x10020;\n\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// Long references still count as blank lines, but raw-source fixes cannot
		// remove their decoded line feeds, just as with shorter references.
		{Code: "<><A/>&#00000000010;&#x000000000A;<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output:  slices.Repeat([]string{"<><A/>&#00000000010;&#x000000000A;<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 1, Column: 35, EndLine: 1, EndColumn: 39},
			},
		},
		// Invalid digits and missing semicolons leave the reference as text.
		{Code: "<><A/>&#x1two;&#10<B/></>", Tsx: true,
			Output: slices.Repeat([]string{"<><A/>&#x1two;&#10<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 1, Column: 19, EndLine: 1, EndColumn: 23},
			},
		},
		// Type arguments, namespaced tags and member tags stay within their elements.
		{Code: "<><List<Value> />\n<svg:path></svg:path>\n<UI.Item /></>", Tsx: true,
			Output: []string{"<><List<Value> />\n\n<svg:path></svg:path>\n\n<UI.Item /></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 22},
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 12},
			},
		},
		// Entity decoding is not recursive: escaped references remain visible text.
		{Code: "<><A/>&amp;#10;&amp;#10;<B/></>", Tsx: true,
			Output: slices.Repeat([]string{"<><A/>&amp;#10;&amp;#10;<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
			},
		},
		// Non-whitespace text between blank lines is preserved across fix passes.
		{Code: "<><A/>\n\ntext\n\n<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output:  []string{"<><A/>\n\ntext\n<B/></>", "<><A/>\ntext\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 1, EndLine: 5, EndColumn: 5},
			},
		},
		// A fragment can still be the reported right sibling.
		{Code: `<><A/>
<><B/></></>`, Tsx: true,
			Output: []string{`<><A/>

<><B/></></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
			},
		},
		// A spread child can still be the reported right sibling.
		{Code: `<><A/>
{...items}</>`, Tsx: true,
			Output: []string{`<><A/>

{...items}</>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
			},
		},
		// An empty expression is still an expression container.
		{Code: `<>{}
<B/></>`, Tsx: true,
			Output: []string{`<>{}

<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		// Whitespace after the brace prevents the block-comment exemption.
		{Code: `<>{ /*comment*/}
<B/></>`, Tsx: true,
			Output: []string{`<>{ /*comment*/}

<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		// A line comment is not exempt.
		{Code: `<>{// comment
}
<B/></>`, Tsx: true,
			Output: []string{`<>{// comment
}

<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// Multiline lookahead skips comments, fragments and spread children.
		{Code: `<><A/>
{/*comment*/}
<>
<B/>
</>
{...items}
<C
 prop={x}/></>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`<><A/>

{/*comment*/}
<>
<B/>
</>
{...items}
<C
 prop={x}/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
			},
		},
		// A one-line fragment does not make the preceding multiline child exempt.
		{Code: `<><A
/>
<></></>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`<><A
/>

<></></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 6},
			},
		},
		// Parentheses and TypeScript syntax remain inside expression-container ranges.
		{Code: `<>{(value?.[key] as string)}
{(
value!
)}</>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`<>{(value?.[key] as string)}

{(
value!
)}</>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 2, Column: 1, EndLine: 4, EndColumn: 3},
			},
		},
		// Explicit default options match omitted options.
		{Code: `<><A/>
<B/></>`, Tsx: true,
			Options: []any{map[string]any{"prevent": false, "allowMultilines": false}},
			Output: []string{`<><A/>

<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		// An empty option object uses both defaults.
		{Code: `<><A/>
<B/></>`, Tsx: true,
			Options: []any{map[string]any{}},
			Output: []string{`<><A/>

<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		// A space between children reports without changing text (upstream issue 3296).
		{Code: "<p>{t('text')} <span>{email}</span></p>", Tsx: true,
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"<p>{t('text')} <span>{email}</span></p>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 1, Column: 16, EndLine: 1, EndColumn: 36},
			},
		},
		// Entity-only blank lines cannot be removed by the raw-source fixer.
		{Code: "<><A/>&#10;&#10;<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"<><A/>&#10;&#10;<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
			},
		},
		// JavaScript whitespace excludes NEL.
		{Code: "<><A/>\n\u0085\n<B/></>", Tsx: true,
			Output: []string{"<><A/>\n\u0085\n\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// CRLF diagnostics and raw fix behavior match upstream.
		{Code: "<><A/>\r\n<B/></>", Tsx: true,
			Output: []string{"<><A/>\r\n\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
			},
		},
		// CRLF blank lines are reported but unchanged in prevent mode.
		{Code: "<><A/>\r\n\r\n<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"<><A/>\r\n\r\n<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// Indented blank lines are reported but unchanged in prevent mode.
		{Code: "<><A/>\n \n<B/></>", Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"<><A/>\n \n<B/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// Multiple empty lines are removed over multiple passes.
		{Code: `<><A/>



<B/></>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`<><A/>


<B/></>`, `<><A/>

<B/></>`, `<><A/>
<B/></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 1, EndLine: 5, EndColumn: 5},
			},
		},
		// A CR stops the replacement lookahead.
		{Code: "<><A/>\ntext\r\n<B/></>", Tsx: true,
			Output: []string{"<><A/>\n\ntext\r\n\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 1, EndLine: 3, EndColumn: 5},
			},
		},
		// Unicode line separators stop the replacement lookahead.
		{Code: "<><A/>\ntext\u2028\ntext\u2029\n<B/></>", Tsx: true,
			Output: []string{"<><A/>\n\ntext\u2028\n\ntext\u2029\n\n<B/></>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 6, Column: 1, EndLine: 6, EndColumn: 5},
			},
		},
		// Diagnostic columns count UTF-16 units, including supplementary characters.
		{Code: "<>你好😀<A/> <名字/></>", Tsx: true,
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"<>你好😀<A/> <名字/></>"}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
			},
		},
		// Multiline detection recognizes a Unicode line separator.
		{Code: "<><A/>\n{value\u2028+ other}</>", Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output:  []string{"<><A/>\n\n{value\u2028+ other}</>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 2, Column: 1, EndLine: 3, EndColumn: 9},
			},
		},
	})
}
