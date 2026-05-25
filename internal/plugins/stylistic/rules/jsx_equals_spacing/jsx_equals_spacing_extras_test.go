// TestJsxEqualsSpacingExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Upstream's entire suite is self-closing single-line ASCII; the cases here add
// the tsgo-specific shapes (JsxOpeningElement vs JsxSelfClosingElement,
// namespaced & hyphenated attribute names, nested JSX, JSX-valued attributes,
// cross-line `=`, Unicode/tab whitespace), realistic multi-attribute component
// shapes, and the option-parsing / branch lock-ins.
package jsx_equals_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxEqualsSpacingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxEqualsSpacingRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: declaration/container form — JsxOpeningElement (with children), not self-closing ----
		{Code: "<App foo=\"x\"></App>", Tsx: true},
		{Code: "<App foo = \"x\"></App>", Tsx: true, Options: []interface{}{"always"}},
		// ---- Dimension 4: namespaced attribute name (JsxAttributeName = Identifier | JsxNamespacedName) ----
		{Code: "<App a:b=\"c\" />", Tsx: true},
		{Code: "<App a:b = \"c\" />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Real-user: data-* / aria-* hyphenated attribute names (high-frequency React shapes) ----
		{Code: "<App data-foo=\"x\" aria-label=\"y\" />", Tsx: true},
		{Code: "<App data-foo = \"x\" aria-label = \"y\" />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Dimension 4: nesting boundary — outer + inner both clean ----
		{Code: "<App foo=\"x\"><Bar baz=\"y\" /></App>", Tsx: true},
		// ---- Dimension 4: JSX element as attribute value — inner element checked independently ----
		{Code: "<App foo={<Bar baz=\"y\" />} />", Tsx: true},
		// ---- Real-user: realistic multi-attribute component (mixed string / expression / boolean props) ----
		{Code: "<Button type=\"submit\" onClick={fn} disabled />", Tsx: true},
		// ---- Real-user: JSX returned from an arrow / inside a ternary (render-context shapes) ----
		{Code: "const C = () => <App foo=\"x\" />", Tsx: true},
		{Code: "const x = cond ? <A foo=\"x\" /> : <B bar=\"y\" />", Tsx: true},
		// ---- Dimension 4: graceful degradation — mixed spread / valued / valueless attributes ----
		{Code: "<App {...a} foo=\"x\" bar />", Tsx: true},
		{Code: "<App {...a} foo = \"x\" bar />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Dimension 4: graceful degradation — fragment has no attributes, must not crash ----
		{Code: "<>foo</>", Tsx: true},
		{Code: "<><App foo=\"x\" /></>", Tsx: true},
		// ---- Dimension 4: value-node kind is irrelevant (member / element-access expressions inside {}) ----
		{Code: "<App foo={a.b.c} />", Tsx: true},
		{Code: "<App foo={a['b']} />", Tsx: true},
		// ---- Dimension 4: cross-line — newline counts as a space (no same-line gate); `always` satisfied ----
		{Code: "<App foo\n= {bar} />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Dimension 4 / autofix: multiple spaces satisfy `always` (isSpaceBetween only checks existence) ----
		{Code: "<App foo  =  \"x\" />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Dimension 4: tab is whitespace — satisfies `always` ----
		{Code: "<App foo\t=\t\"x\" />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Real-user: comment between name and `=` with NO surrounding space is not a "space" (isSpaceBetween walks tokens) ----
		{Code: "<App foo/* c */={bar} />", Tsx: true},
		// ---- Option shape: bare string (single-option CLI form) exercises the non-array parseOption path ----
		{Code: "<App foo=\"x\" />", Tsx: true, Options: "never"},
		{Code: "<App foo = \"x\" />", Tsx: true, Options: "always"},
		// ---- Branch A lock-in: spread attribute is skipped under both configs ----
		{Code: "<App {...props} />", Tsx: true, Options: []interface{}{"always"}},
		// ---- Branch A lock-in: valueless attribute is skipped under both configs ----
		{Code: "<App foo bar />", Tsx: true, Options: []interface{}{"always"}},
		// ---- N/A: receiver/expression wrappers (paren, non-null `!`, optional-chain `?.`) — rule inspects attribute name/`=`/value position, never a member/call receiver; value internals are not parsed ----
		// ---- N/A: string / numeric / computed / private-identifier keys — a JSX attribute name is only Identifier or JsxNamespacedName ----
		// ---- N/A: element access X['y'] as the matched input — rule does not do dotted member access ----
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: declaration/container form — JsxOpeningElement (with children) ----
		{
			Code:   "<App foo = \"x\"></App>",
			Tsx:    true,
			Output: []string{"<App foo=\"x\"></App>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Dimension 4: namespaced attribute name — name.End() lands past `b`, `=` located correctly ----
		{
			Code:   "<App a:b = \"c\" />",
			Tsx:    true,
			Output: []string{"<App a:b=\"c\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Real-user: data-* hyphenated name with spaced `=` (verifies name.End() spans the full hyphenated identifier, not just `data`) ----
		{
			Code:   "<App data-foo = \"x\" />",
			Tsx:    true,
			Output: []string{"<App data-foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
			},
		},
		// ---- Dimension 4: nesting boundary — outer JsxOpeningElement + inner JsxSelfClosingElement both reported, no bleed ----
		{
			Code:   "<App foo = \"x\"><Bar baz = \"y\" /></App>",
			Tsx:    true,
			Output: []string{"<App foo=\"x\"><Bar baz=\"y\" /></App>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
			},
		},
		// ---- Dimension 4: JSX element as attribute value — only inner element has spacing; outer foo={...} is tight ----
		{
			Code:   "<App foo={<Bar baz = \"y\" />} />",
			Tsx:    true,
			Output: []string{"<App foo={<Bar baz=\"y\" />} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
			},
		},
		// ---- Real-user: realistic component, only the spaced prop reported (string / boolean props untouched) ----
		{
			Code:   "<Button type=\"submit\" onClick = {fn} disabled />",
			Tsx:    true,
			Output: []string{"<Button type=\"submit\" onClick={fn} disabled />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
			},
		},
		// ---- Real-user: JSX in arrow-return render context — the `const C =` and `=>` tokens are NOT JSX `=`, only `foo`'s is reported ----
		{
			Code:   "const C = () => <App foo = \"x\" />",
			Tsx:    true,
			Output: []string{"const C = () => <App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
			},
		},
		// ---- Dimension 4 / autofix: cross-line — newline before `=` is a space; fix deletes the newline ----
		{
			Code:    "<App foo\n={bar} />",
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
			},
		},
		// ---- Dimension 4 / autofix: multiple spaces around `=` — never deletes ALL of them ----
		{
			Code:   "<App foo  =  \"x\" />",
			Tsx:    true,
			Output: []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
			},
		},
		// ---- Dimension 4 / autofix: tab whitespace around `=` is deleted under never ----
		{
			Code:   "<App foo\t=\t\"x\" />",
			Tsx:    true,
			Output: []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Dimension 4: graceful degradation — spread + valueless attributes skipped, only `foo` reported ----
		{
			Code:   "<App {...a} foo = \"x\" bar />",
			Tsx:    true,
			Output: []string{"<App {...a} foo=\"x\" bar />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// ---- Dimension 4: value-node kind irrelevant — member-expression value with spaced `=` ----
		{
			Code:   "<App foo = {a.b.c} />",
			Tsx:    true,
			Output: []string{"<App foo={a.b.c} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Real-user: always mode, multi-attribute mix of missing-after (type) and missing-both (onClick) ----
		{
			Code:    "<Button type =\"submit\" onClick={fn} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<Button type = \"submit\" onClick = {fn} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				{MessageId: "needSpaceBefore", Message: "A space is required before '='", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
			},
		},
		// ---- Real-user: a space BEFORE a comment counts as a space (token-walk gap); fix removes the whole span incl. the comment ----
		{
			Code:   "<App foo /* c */={bar} />",
			Tsx:    true,
			Output: []string{"<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
			},
		},
		// ---- Option shape: bare string "never" (CLI single-option form) ----
		{
			Code:    "<App foo = \"x\" />",
			Tsx:     true,
			Options: "never",
			Output:  []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Option shape: bare string "always" (CLI single-option form) ----
		{
			Code:    "<App foo=\"x\" />",
			Tsx:     true,
			Options: "always",
			Output:  []string{"<App foo = \"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "needSpaceBefore", Message: "A space is required before '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				{MessageId: "needSpaceAfter", Message: "A space is required after '='", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
			},
		},
		// ---- Locks in parseOption fallback: out-of-enum option degrades to the "never" default ----
		{
			Code:    "<App foo = \"x\" />",
			Tsx:     true,
			Options: "bogus",
			Output:  []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Locks in parseOption fallback: empty option array degrades to the "never" default ----
		{
			Code:    "<App foo = \"x\" />",
			Tsx:     true,
			Options: []interface{}{},
			Output:  []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Dimension 4: Unicode whitespace (NBSP) around `=` — ESLint's isSpaceBetween counts any non-token char as a space (cf. arrow-spacing's NBSP lock-in) ----
		{
			Code:   "<App foo\u00a0=\u00a0\"x\" />",
			Tsx:    true,
			Output: []string{"<App foo=\"x\" />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Dimension 4: deep 3-level nesting (opening + opening + self-closing) — every level reported, no bleed / no miss ----
		{
			Code:   "<A a = \"1\"><B b = \"2\"><C c = \"3\" /></B></A>",
			Tsx:    true,
			Output: []string{"<A a=\"1\"><B b=\"2\"><C c=\"3\" /></B></A>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				{MessageId: "noSpaceBefore", Message: "There should be no space before '='", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
				{MessageId: "noSpaceAfter", Message: "There should be no space after '='", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
			},
		},
	})
}
