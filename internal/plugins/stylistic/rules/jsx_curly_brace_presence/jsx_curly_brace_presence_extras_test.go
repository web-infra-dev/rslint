// TestJsxCurlyBracePresenceExtras locks in branches and edge shapes the upstream
// @stylistic suite doesn't exercise. Grouped into:
//
//  1. @stylistic variant delta — the quote gate. This rule is built on the
//     shared react/jsx-curly-brace-presence implementation with
//     stylisticQuotes=true, so the ONLY behavioral difference is that an
//     attribute string literal whose value contains a quote character is left
//     wrapped instead of reported. Those cases are valid here but invalid under
//     react — the reason the rule exists separately, locked in first.
//  2. Dimension 4 — the tsgo↔ESTree shape surface the rule classifies on:
//     ParenthesizedExpression (ESTree flattens), TS type-expression wrappers
//     (as / satisfies / non-null), the Kind*Literal forms ESTree collapses into
//     `Literal`, and JSX element / fragment cross-nesting.
//  3. Real-user shapes from the upstream GitHub issue tracker.
//  4. Diagnostic contract — exact positions and message text.
//  5. Options JSON path — array-wrapped (CLI / rule-tester) shapes.
package jsx_curly_brace_presence

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxCurlyBracePresenceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyBracePresenceRule, []rule_tester.ValidTestCase{
		// ===== @stylistic variant delta: quote-bearing attribute values stay
		// wrapped (react would report+unwrap these). =====
		{Code: `<Foo bar={"'"} />`, Tsx: true, Options: map[string]interface{}{"props": "never", "children": "never", "propElementValues": "never"}},
		{Code: `<App prop={'say "hi"'} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App prop={"it's"} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App prop={"it's"} />`, Tsx: true, Options: "never"},
		// Parenthesized receiver + quote: SkipParentheses still finds the quote.
		{Code: `<App prop={("it's")} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		// Template path is NOT affected by the variant — a quote in the cooked
		// value already blocks unwrapping in BOTH react and @stylistic.
		{Code: "<App prop={`has \"q\" inside`} />", Tsx: true, Options: map[string]interface{}{"props": "never"}},

		// ---- Dimension 4: TS type-expression wrappers. After SkipParentheses,
		// As/Satisfies/NonNull are not string/template/JSX-like, so the braces
		// stay (the attribute classification gate bails). ----
		{Code: `<App prop={('x' as string)} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App prop={('x' satisfies string)} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App prop={x!} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App>{('x' as string)}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},

		// ---- Dimension 4: literal kinds ESTree collapses into `Literal`. Only
		// string-valued literals are unwrap candidates. ----
		{Code: `<App>{123}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		{Code: `<App>{null}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		{Code: `<App>{foo}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		{Code: `<App>{[]}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		{Code: `<App prop={123} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<App prop={true} />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},

		// ---- Dimension 4: TemplateExpression (substitution present) never unwraps. ----
		{Code: "<App prop={`a ${x} b`} />", Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: "<App>{`a ${x} b`}</App>", Tsx: true, Options: map[string]interface{}{"children": "never"}},

		// ---- Dimension 4: optional chain / element access ----
		// N/A: the rule only inspects JsxExpression contents, JsxAttribute
		// initializers, and JsxText — never member/element-access receivers, so
		// `X?.y`, `X['y']`, `X?.()` cannot reach a classification branch.

		// ---- Dimension 4: graceful degradation ----
		{Code: `<App {...rest} />`, Tsx: true, Options: "never"},                                           // JsxSpreadAttribute not visited
		{Code: `<App>{}</App>`, Tsx: true, Options: "never"},                                               // empty container, no crash
		{Code: `<App readOnly />`, Tsx: true, Options: map[string]interface{}{"props": "always"}},          // bare boolean attr, nil Initializer
		{Code: "<App>\n  <b />\n</App>", Tsx: true, Options: map[string]interface{}{"children": "always"}}, // surrounding whitespace JsxText not wrapped

		// ---- Dimension 4: attribute-name shapes (rule operates on the
		// initializer; the name shape is irrelevant). ----
		{Code: `<svg xmlns:xlink="http://www.w3.org/1999/xlink" />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},
		{Code: `<my-elem data-foo="bar" />`, Tsx: true, Options: map[string]interface{}{"props": "never"}},

		// ---- Comment inside `{…}` suppresses unwrap (full-range scan). ----
		{Code: `<App>{'foo' /* trailing */}</App>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},

		// ===== Real-user shapes (jsx-eslint/eslint-plugin-react issue tracker) =====
		// #2299: multi-line substitution-free template literal stays wrapped.
		{
			Code: "<textarea value={`First line\nsecond line\nthird line`} />",
			Tsx:  true, Options: map[string]interface{}{"props": "never", "children": "never"},
		},
		// #2885: a comment before a JSX element — braces must stay (else the
		// comment is deleted).
		{
			Code: `
        <App>
          {
            /* This is a very important note on the use of the below component. */
            <Component />
          }
        </App>
      `,
			Tsx: true, Options: map[string]interface{}{"children": "never"},
		},
		// #2427 / #2454: the `{' '}` whitespace-injection idiom stays wrapped.
		{Code: `<MyComponent>{' '}</MyComponent>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		{Code: `<MyComponent>{'    '}</MyComponent>`, Tsx: true, Options: map[string]interface{}{"children": "never"}},
		// #3228 / #3184: a JSX element prop value stays wrapped under the default
		// propElementValues='ignore'.
		{Code: `<App extra={<Foo />} />`, Tsx: true, Options: map[string]interface{}{"props": "never", "children": "never"}},
		{
			Code: `
        <CollapsibleTitle
          extra={<span className="activity-type">{activity.type}</span>}
        />
      `,
			Tsx: true, Options: "never",
		},

		// ---- Options JSON path: array-wrapped string shorthand (CLI / multi-element shape). ----
		{Code: `<MyComponent prop="bar" attr='foo' />`, Tsx: true, Options: []interface{}{"never"}},
	}, []rule_tester.InvalidTestCase{
		// ===== Variant boundary: the quote gate is ATTRIBUTE-only. Quote payloads
		// in CHILDREN still report; quote-free attribute values still unwrap. =====
		// (Also Real-user #3214: this unwrap is the source of the reported
		// no-unescaped-entities conflict — rslint matches upstream's unwrap.)
		{
			Code:    `<App>{'foo "bar"'}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App>foo "bar"</App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    `<App>{"it's"}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App>it's</App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    `<App prop={'plain'} />`,
			Tsx:     true,
			Options: map[string]interface{}{"props": "never"},
			Output:  []string{`<App prop="plain" />`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},

		// ---- Dimension 4: multi-level ParenthesizedExpression — SkipParentheses
		// must peel every layer before classifying / emitting the fix. ----
		{
			Code:    `<App>{(('foo'))}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App>foo</App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    "<App>{(`foo`)}</App>",
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App>foo</App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    `<App>{(<Foo />)}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App><Foo /></App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},

		// ---- Dimension 4: JSX element / fragment cross-nesting; the listener
		// fires regardless of parent kind and must not bleed across boundaries. ----
		{
			Code:    `<App>{<>x</>}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App><>x</></App>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		{
			Code:    `<><Inner prop={'x'} /></>`,
			Tsx:     true,
			Options: map[string]interface{}{"props": "never"},
			Output:  []string{`<><Inner prop="x" /></>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
		// Same-kind nesting: both prop values report; listener doesn't bleed.
		{
			Code:    `<Outer prop={'a'}><Inner prop={'b'} /></Outer>`,
			Tsx:     true,
			Options: map[string]interface{}{"props": "never"},
			Output:  []string{`<Outer prop="a"><Inner prop="b" /></Outer>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryCurly"},
				{MessageId: "unnecessaryCurly"},
			},
		},
		// 3-level deep children=always wraps at every level.
		{
			Code:    `<A><B><C>foo</C></B></A>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "always"},
			Output:  []string{`<A><B><C>{"foo"}</C></B></A>`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCurly"}},
		},

		// ---- Locks in upstream propElementValues='never' arm: reports
		// unnecessaryCurly on the `{<div />}` container. SKIP: the only fix
		// output is the braceless `horror=<div />`, which the TS/JSX grammar
		// rejects (upstream gates the equivalent on `features: ['no-ts']`). ----
		{
			Code:    `<App horror={<div />} />`,
			Tsx:     true,
			Options: map[string]interface{}{"propElementValues": "never"},
			Output:  []string{`<App horror=<div /> />`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
			Skip:    true,
		},
		// ---- Real-user #2885 variant: a leading comment before a STRING literal
		// also suppresses the unwrap. SKIP: valid-in-disguise (0 diagnostics);
		// the valid element form is asserted above. ----
		{
			Code:    `<App>{/* keep me */ 'foo'}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Errors:  []rule_tester.InvalidTestCaseError{},
			Skip:    true,
		},

		// ===== Diagnostic contract: exact positions + message text. =====
		{
			Code:    `<App prop={'bar'}>foo</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"props": "never"},
			Output:  []string{`<App prop="bar">foo</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryCurly", Message: "Curly braces are unnecessary here.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			},
		},
		{
			Code:    `<App>{'foo'}</App>`,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output:  []string{`<App>foo</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryCurly", Line: 1, Column: 6, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code: `
        <MyComponent>
          {'%'}
        </MyComponent>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"children": "never"},
			Output: []string{`
        <MyComponent>
          %
        </MyComponent>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryCurly", Line: 3, Column: 11, EndLine: 3, EndColumn: 16},
			},
		},
		{
			Code:    `<App prop='foo' />`,
			Tsx:     true,
			Options: map[string]interface{}{"props": "always"},
			Output:  []string{`<App prop={"foo"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurly", Message: "Need to wrap this literal in a JSX expression.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
			},
		},

		// ---- Options JSON path: array-wrapped object (CLI / multi-element shape). ----
		{
			Code:    `<App prop={'bar'} />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"props": "never"}},
			Output:  []string{`<App prop="bar" />`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryCurly"}},
		},
	})
}
