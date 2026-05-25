// TestJsxCurlyNewlineExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package jsx_curly_newline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxCurlyNewlineExtras(t *testing.T) {
	const (
		expectedAfter    = "Expected newline after '{'."
		expectedBefore   = "Expected newline before '}'."
		unexpectedAfter  = "Unexpected newline after '{'."
		unexpectedBefore = "Unexpected newline before '}'."
	)
	singleForbid := []interface{}{map[string]interface{}{"singleline": "forbid"}}
	singleRequire := []interface{}{map[string]interface{}{"singleline": "require"}}
	multiForbid := []interface{}{map[string]interface{}{"multiline": "forbid"}}
	// Discriminator option: singleline and multiline disagree, so the
	// single-/multi-line decision (which the paren-unwrap affects) is observable.
	singleReqMultiForbid := []interface{}{map[string]interface{}{"singleline": "require", "multiline": "forbid"}}
	multilineRequire := []interface{}{map[string]interface{}{"singleline": "consistent", "multiline": "require"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyNewlineRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized expression (single line, no newlines) ----
		{Code: "<div>{(foo)}</div>", Tsx: true},
		// ---- Dimension 4: TS type-expression wrappers (as / non-null) ----
		{Code: "<div>{foo as string}</div>", Tsx: true},
		{Code: "<div>{foo!}</div>", Tsx: true},
		// ---- Dimension 4: optional chain (no ChainExpression wrapper in tsgo) ----
		{Code: "<div>{foo?.bar}</div>", Tsx: true},
		// ---- Dimension 4: element access ----
		{Code: "<div>{foo['bar']}</div>", Tsx: true},
		// ---- Dimension 4: object literal — outer container braces, not the object's ----
		{Code: "<div style={{ color: 'red' }} />", Tsx: true},
		{Code: "<div style={{\n  color: 'red'\n}} />", Tsx: true},
		// ---- Dimension 4: nested JSX containers — listener fires per container, no bleed ----
		{Code: "<div>{cond && <span>{foo}</span>}</div>", Tsx: true},
		// ---- Dimension 4: deep nesting (3 container levels), all single-line ----
		{Code: "<div>{a && <span>{b && <i>{c}</i>}</span>}</div>", Tsx: true},
		// ---- Dimension 4: fragment children ----
		{Code: "<>{foo}</>", Tsx: true},
		{Code: "<>{\nfoo\n}</>", Tsx: true},
		// ---- Dimension 4: attribute container holding an element with its own container ----
		{Code: "<Foo bar={<Baz qux={x} />} />", Tsx: true},
		// ---- Real-user: the canonical multiline `.map` child — must NOT false-positive
		// under the default `consistent` (braces hug the call, content broke inside). ----
		{Code: "<ul>{items.map((i) =>\n<li>{i}</li>\n)}</ul>", Tsx: true},
		// ---- Dimension 4: empty container / graceful degradation ----
		{Code: "<div>{}</div>", Tsx: true},
		{Code: "<div>{}</div>", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<div>{\n}</div>", Tsx: true},
		{Code: "<div>{/* comment */}</div>", Tsx: true},
		// ---- Dimension 4: spread child — JSXSpreadChild in ESTree, not visited ----
		{Code: "<div>{...children}</div>", Tsx: true, Options: []interface{}{"never"}},
		// ---- Dimension 4: spread attribute — JsxSpreadAttribute, not a JsxExpression ----
		{Code: "<div {...props} />", Tsx: true, Options: []interface{}{"never"}},
		// Locks in shouldHaveNewlines(): multiline:'require' leaves singleline at
		// its 'consistent' default — a single-line expr without newlines is fine.
		{Code: "<div>{foo}</div>", Tsx: true, Options: multiForbid},
		// Exercises the bare-object options path (CLI single-option shape).
		{Code: "<div>{foo}</div>", Tsx: true, Options: map[string]interface{}{"multiline": "require"}},
	}, []rule_tester.InvalidTestCase{
		// Locks in shouldHaveNewlines() singleline='forbid' arm: single-line expr
		// with newlines on both sides is rejected.
		{
			Code:    "<div>{\nfoo\n}</div>",
			Tsx:     true,
			Options: singleForbid,
			Output:  []string{"<div>{foo}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},
		// Locks in shouldHaveNewlines() singleline='require' arm: single-line expr
		// missing newlines on both sides.
		{
			Code:    "<div>{foo}</div>",
			Tsx:     true,
			Options: singleRequire,
			Output:  []string{"<div>{\nfoo\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// Locks in shouldHaveNewlines() multiline='forbid' arm via the OBJECT
		// form (a distinct code path from the 'never' string shorthand).
		{
			Code:    "<div>{\nfoo &&\nbar\n}</div>",
			Tsx:     true,
			Options: multiForbid,
			Output:  []string{"<div>{foo &&\nbar}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// Locks in parseOptions() default arm: no options => 'consistent'.
		{
			Code:   "<div>{\nfoo}</div>",
			Tsx:    true,
			Output: []string{"<div>{\nfoo\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
			},
		},
		// ---- Dimension 4: empty container — isSingleLine of the JSXEmptyExpression
		// is decided by whether the braces share a line. A single-line `{ }` under
		// singleline='require' must report both sides (multiline defaults to
		// 'consistent', which would yield 0 errors if the braces were misjudged). ----
		{
			Code:    "<div>{ }</div>",
			Tsx:     true,
			Options: singleRequire,
			Output:  []string{"<div>{\n \n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
			},
		},
		// ---- Dimension 4: nested containers fixed independently — the inner
		// container violates 'never' while the outer is already compliant. ----
		{
			Code:    "<div>{<span>{\nfoo\n}</span>}</div>",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<div>{<span>{foo}</span>}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},
		// ---- Dimension 4: multi-byte content — tsgo positions are byte offsets, so
		// the fix ranges must slice on byte boundaries (UTF-16 offsets would corrupt
		// the CJK identifier or trip the brace guard). ----
		{
			Code:    "<div>{\n你好\n}</div>",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<div>{你好}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},
		// ---- Dimension 4: U+2028 LINE SEPARATOR counts as a line break (tsgo's
		// ECMALineMap and ESLint's LINEBREAK_MATCHER agree), so `}` is on a new line. ----
		{
			Code:    "<div>{foo\u2028}</div>",
			Tsx:     true,
			Options: []interface{}{"consistent"},
			Output:  []string{"<div>{foo}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
			},
		},
		// ---- Dimension 4: comment-only empty container spanning lines — both sides
		// report under 'never', and neither is fixed (the gap is the comment). ----
		{
			Code:    "<div>{/* a\nb */}</div>",
			Tsx:     true,
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 2, Column: 5, EndLine: 2, EndColumn: 6},
			},
		},
		// Locks in parseOptions() bare-string path ('never' not array-wrapped).
		{
			Code:    "<div>{\nfoo\n}</div>",
			Tsx:     true,
			Options: "never",
			Output:  []string{"<div>{foo}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},
		// ---- Dimension 4: parenthesized expression — isSingleLine uses the
		// UNWRAPPED inner node. Here the parens span lines but `foo` is single-
		// line, so singleline='require' applies (multiline='forbid' would give 0
		// errors if we wrongly judged by the ParenthesizedExpression's own span).
		{
			Code:    "<div>{(\nfoo\n)}</div>",
			Tsx:     true,
			Options: singleReqMultiForbid,
			Output:  []string{"<div>{\n(\nfoo\n)\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 3, Column: 2, EndLine: 3, EndColumn: 3},
			},
		},
		// ---- Real-user: eslint-plugin-react#3097 — multiline ternary gets
		// newlines added after { and before }, with NO indentation adjustment. ----
		{
			Code:    "<div>{cond\n? a\n: b}</div>",
			Tsx:     true,
			Options: multilineRequire,
			Output:  []string{"<div>{\ncond\n? a\n: b\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 3, Column: 4, EndLine: 3, EndColumn: 5},
			},
		},
		// ---- Real-user: eslint-stylistic#292 — closing brace moved to its own
		// line (indentation untouched) when content already broke after {. ----
		{
			Code:    "<div>{\ntrue && (\n<p/>\n)}</div>",
			Tsx:     true,
			Options: []interface{}{"consistent"},
			Output:  []string{"<div>{\ntrue && (\n<p/>\n)\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 4, Column: 2, EndLine: 4, EndColumn: 3},
			},
		},
	})
}
