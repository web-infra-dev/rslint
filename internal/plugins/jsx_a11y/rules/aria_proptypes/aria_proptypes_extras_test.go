// cspell:ignore foobar

package aria_proptypes

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestAriaProptypesExtras covers tsgo-specific edge cases that upstream's
// JS test file doesn't exercise but which are likely sources of silent
// drift between rslint and ESLint:
//
//   - Case-insensitive attribute names (ARIA-HIDDEN, Aria-Hidden)
//   - Parenthesized values (tsgo preserves Parens; ESTree flattens)
//   - TS type-assertion wrappers (`as`, `!`, `satisfies`) — LITERAL_TYPES
//     has no entry → noop → null → step-3 skip (VALID)
//   - aria-haspopup heterogeneous list — boolean tokens alongside strings
//   - aria-current heterogeneous list — strings + boolean tokens
//   - aria-orientation's quirky string "undefined" as a valid token value
//   - aria-pressed (tristate, distinct from aria-checked)
//   - Token validity-check case sensitivity edge cases
//   - Multi-attribute elements / mixed valid+invalid attrs
//   - Nested JSX with invalid attrs at multiple depths
//   - Bigint literals
func TestAriaProptypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaProptypesRule,
		[]rule_tester.ValidTestCase{
			// Case-insensitive attribute names — upstream's `name.toLowerCase()`
			// applies before the prefix and aria-query lookups.
			{Code: `<div ARIA-HIDDEN="true" />`, Tsx: true},
			{Code: `<div Aria-Hidden="true" />`, Tsx: true},
			{Code: `<div ARIA-LEVEL={3} />`, Tsx: true},

			// Parenthesized values — tsgo preserves ParenthesizedExpression,
			// our SkipParentheses unwrap mirrors ESTree's parse-time flatten.
			{Code: `<div aria-hidden={(true)} />`, Tsx: true},
			{Code: `<div aria-hidden={((true))} />`, Tsx: true},
			{Code: `<div aria-level={(123)} />`, Tsx: true},

			// TS type-assertion wrappers — LITERAL_TYPES has no entry for
			// TSAsExpression / TSNonNullExpression / TSSatisfiesExpression /
			// TSTypeAssertion; all fall through to noop → null → step-3 skip.
			{Code: `<div aria-hidden={true as boolean} />`, Tsx: true},
			{Code: `<div aria-hidden={x!} />`, Tsx: true},
			{Code: `<div aria-hidden={x satisfies boolean} />`, Tsx: true},
			{Code: `<div aria-label={"x" as const} />`, Tsx: true},

			// aria-haspopup (token: [false, true, 'menu', 'listbox', 'tree', 'grid', 'dialog']).
			// Heterogeneous list — booleans AND strings both valid.
			{Code: `<div aria-haspopup={true} />`, Tsx: true},
			{Code: `<div aria-haspopup={false} />`, Tsx: true},
			{Code: `<div aria-haspopup="true" />`, Tsx: true},
			{Code: `<div aria-haspopup="false" />`, Tsx: true},
			{Code: `<div aria-haspopup="menu" />`, Tsx: true},
			{Code: `<div aria-haspopup="MENU" />`, Tsx: true},
			{Code: `<div aria-haspopup="listbox" />`, Tsx: true},
			{Code: `<div aria-haspopup="tree" />`, Tsx: true},
			{Code: `<div aria-haspopup="grid" />`, Tsx: true},
			{Code: `<div aria-haspopup="dialog" />`, Tsx: true},

			// aria-current (token: ['page','step','location','date','time', true, false]).
			{Code: `<div aria-current="page" />`, Tsx: true},
			{Code: `<div aria-current="STEP" />`, Tsx: true},
			{Code: `<div aria-current={true} />`, Tsx: true},
			{Code: `<div aria-current={false} />`, Tsx: true},
			{Code: `<div aria-current="true" />`, Tsx: true},
			{Code: `<div aria-current="false" />`, Tsx: true},

			// aria-orientation (token: ['vertical', 'undefined', 'horizontal']).
			// "undefined" is intentionally a STRING in upstream's list; it
			// matches the string literal, NOT the JS `undefined` keyword.
			{Code: `<div aria-orientation="vertical" />`, Tsx: true},
			{Code: `<div aria-orientation="HORIZONTAL" />`, Tsx: true},
			{Code: `<div aria-orientation="undefined" />`, Tsx: true},

			// aria-pressed (tristate, distinct rule entry from aria-checked).
			{Code: `<div aria-pressed={true} />`, Tsx: true},
			{Code: `<div aria-pressed={false} />`, Tsx: true},
			{Code: `<div aria-pressed="mixed" />`, Tsx: true},

			// aria-expanded / aria-grabbed / aria-selected — allowUndefined.
			// Identifier `undefined` triggers step-1 (PropValueIsNullish),
			// so the rule short-circuits before the allowUndefined branch.
			// Verifying these are valid lets a future refactor surface
			// regression if step-1 ever stops gating undefined.
			{Code: `<div aria-expanded={undefined} />`, Tsx: true},
			{Code: `<div aria-grabbed={undefined} />`, Tsx: true},
			{Code: `<div aria-selected={undefined} />`, Tsx: true},

			// aria-live (token: ['assertive', 'off', 'polite']).
			{Code: `<div aria-live="polite" />`, Tsx: true},
			{Code: `<div aria-live="POLITE" />`, Tsx: true},
			{Code: `<div aria-live="off" />`, Tsx: true},

			// aria-autocomplete (token: ['inline','list','both','none']).
			{Code: `<div aria-autocomplete="inline" />`, Tsx: true},
			{Code: `<div aria-autocomplete="LIST" />`, Tsx: true},

			// aria-dropeffect (tokenlist: ['copy','execute','link','move','none','popup']).
			{Code: `<div aria-dropeffect="copy" />`, Tsx: true},
			{Code: `<div aria-dropeffect="copy execute" />`, Tsx: true},
			{Code: `<div aria-dropeffect="copy execute link move" />`, Tsx: true},

			// Integer / number — numeric string with whitespace (JS Number()
			// trims). Upstream test doesn't include this; lock our parity
			// in via the JS Number coercion rules.
			{Code: `<div aria-level=" 3 " />`, Tsx: true},
			{Code: `<div aria-valuemax="3.14" />`, Tsx: true},
			{Code: `<div aria-valuemax={3.14} />`, Tsx: true},

			// Integer / number — empty string coerces to 0. JS:
			// `isNaN(Number("")) === false` is true → valid.
			{Code: `<div aria-level="" />`, Tsx: true},
			{Code: `<div aria-valuemax="" />`, Tsx: true},

			// Multi-attribute element — only the offending attribute reports;
			// peers stay clean.
			{Code: `<div aria-label="ok" aria-hidden={true} aria-level={3} />`, Tsx: true},

			// Boolean form is valid for boolean-typed attributes.
			{Code: `<div aria-busy />`, Tsx: true},
			{Code: `<div aria-disabled />`, Tsx: true},
			{Code: `<div aria-required />`, Tsx: true},

			// Aria-* attribute NOT in aria-query map — gate-1 skip even if
			// the value would have been invalid for some hypothetical type.
			{Code: `<div aria-tabindex="not-a-number" />`, Tsx: true},
			{Code: `<div aria-onclick={fn} />`, Tsx: true},
			// data-* prefix is not aria-*.
			{Code: `<div data-aria-hidden="yes" />`, Tsx: true},

			// JSXSpreadAttribute is not visited — invalid keys inside are
			// silently ignored, matching upstream's JSXAttribute listener.
			{Code: `<div {...{'aria-hidden': 'yes'}} />`, Tsx: true},

			// React patterns — verify the listener fires correctly through
			// nested elements / fragments / library wrappers.
			{Code: `<>{cond && <div aria-hidden>x</div>}</>`, Tsx: true},
			{Code: `function L({xs}) { return xs.map(x => <div aria-hidden key={x} />); }`, Tsx: true},
			{Code: `class C { render() { return <div aria-label="x" />; } }`, Tsx: true},

			// Custom tag — rule fires on the attribute regardless of tag.
			{Code: `<Custom aria-hidden={true} />`, Tsx: true},
			{Code: `<Foo.Bar aria-level={3} />`, Tsx: true},
			{Code: `<my-element aria-hidden />`, Tsx: true},

			// Identifier value that would be falsy at runtime — still
			// LITERAL_TYPES noop → null → step-3 skip. The rule never
			// inspects runtime values.
			{Code: `<div aria-hidden={someExpression} />`, Tsx: true},

			// CallExpression / NewExpression / TaggedTemplateExpression —
			// LITERAL_TYPES noop except TaggedTemplate which extracts the
			// inner quasi. Verify our wiring matches.
			{Code: `<div aria-hidden={fn()} />`, Tsx: true},
			{Code: `<div aria-hidden={new C()} />`, Tsx: true},
			// TaggedTemplate with NoSubstitutionTemplate text "true" —
			// literalPropValue routes to jsxAstUtilsLiteralCoerce → bool true → VALID for boolean.
			{Code: "<div aria-hidden={tag`true`} />", Tsx: true},

			// ConditionalExpression — LITERAL_TYPES noop → null → step-3
			// skip regardless of branch values.
			{Code: `<div aria-checked={cond ? true : "mixed"} />`, Tsx: true},
			{Code: `<div aria-level={cond ? 1 : 2} />`, Tsx: true},

			// LogicalExpression — LITERAL_TYPES noop → null → step-3 skip.
			{Code: `<div aria-hidden={a && b} />`, Tsx: true},
			{Code: `<div aria-hidden={a || b} />`, Tsx: true},
			{Code: `<div aria-hidden={a ?? b} />`, Tsx: true},

			// JSXNamespacedName — `<svg aria:hidden="true" />`. propName
			// returns "aria:hidden" (colon-separated). Lowercase → still
			// "aria:hidden", doesn't start with "aria-" → gate-1 skip.
			{Code: `<svg aria:hidden="true" />`, Tsx: true},
			{Code: `<svg aria:hidden="yes" />`, Tsx: true},

			// BigInt literal as a value — typeof === "bigint" doesn't match
			// any ARIA type check; integer/number's `Number(bigint)` would
			// throw in JS so jsx-ast-utils' LITERAL_TYPES yields the BigInt
			// itself. validityCheck returns false for every type → would be
			// invalid... BUT step-1 may gate first. Verify behavior is
			// stable rather than crashing.
			//
			// In upstream behavior, `<div aria-hidden={123n} />`:
			//   - getPropValue returns BigInt — not null → step-1 doesn't gate.
			//   - getLiteralPropValue returns BigInt — non-null → step-3 doesn't skip.
			//   - validityCheck: typeof "bigint" → not boolean → INVALID.
			// We classify this in extras as invalid below.

			// React library wrappers — verify the listener fires through
			// forwardRef / memo / useMemo / Suspense regardless of the
			// surrounding expression scaffolding.
			{
				Code: `const Btn = React.forwardRef<HTMLButtonElement, {}>((props, ref) => <button ref={ref} aria-pressed="true" />);`,
				Tsx:  true,
			},
			{
				Code: `const M = React.memo(function M() { return <div aria-label="x" />; });`,
				Tsx:  true,
			},
			{
				Code: `function F() { const v = React.useMemo(() => <div aria-hidden>x</div>, []); return v; }`,
				Tsx:  true,
			},
			{
				Code: `<Suspense fallback={<div aria-busy="true">loading</div>}><Page aria-label="x" /></Suspense>`,
				Tsx:  true,
			},

			// Tagged template — `tag\`true\`` routes through
			// literalPropValue's TaggedTemplateExpression arm, which
			// extracts the inner quasi via jsxAstUtilsLiteralCoerce, so a
			// "true"/"false" template coerces to boolean. For aria-hidden
			// (boolean), `tag\`true\`` should pass.
			{Code: "<div aria-hidden={tag`true`} />", Tsx: true},
			{Code: "<div aria-hidden={tag`false`} />", Tsx: true},

			// Numeric literal forms — hex / oct / bin / exponent. JS Number()
			// recognizes all; integer / number validity passes.
			{Code: `<div aria-level={0x10} />`, Tsx: true},
			{Code: `<div aria-level={0o10} />`, Tsx: true},
			{Code: `<div aria-level={0b10} />`, Tsx: true},
			{Code: `<div aria-valuemax={1e3} />`, Tsx: true},

			// Numeric strings with hex / exponent prefix (Number("0x10") = 16).
			{Code: `<div aria-level="0x10" />`, Tsx: true},
			{Code: `<div aria-level="1e3" />`, Tsx: true},

			// idlist allows ANY string — including arbitrary characters,
			// because the inner check is just "is string".
			{Code: `<div aria-labelledby="!@#$%^&*()" />`, Tsx: true},
			{Code: `<div aria-controls="a-b c.d" />`, Tsx: true},

			// PrefixUnary on numeric values for integer/number.
			{Code: `<div aria-level={-0} />`, Tsx: true},
			{Code: `<div aria-valuemin={-1.5} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Case-insensitive lookup — message uses the ORIGINAL (cased)
			// attribute name as written in source, not the normalized form.
			{
				Code: `<div ARIA-HIDDEN="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: "The value for ARIA-HIDDEN must be a boolean."}},
			},
			{
				Code: `<div Aria-Hidden={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: "The value for Aria-Hidden must be a boolean."}},
			},

			// Parenthesized invalid value still reports.
			{
				Code: `<div aria-hidden={(1234)} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},
			{
				Code: `<div aria-hidden={((1234))} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},

			// aria-haspopup invalid token.
			{
				Code: `<div aria-haspopup="dropdown" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-haspopup")}},
			},
			{
				Code: `<div aria-haspopup={123} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-haspopup")}},
			},

			// aria-current cross-class — strings and booleans valid; numbers / other strings not.
			{
				Code: `<div aria-current="foo" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-current")}},
			},

			// aria-orientation — non-listed string.
			{
				Code: `<div aria-orientation="diagonal" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-orientation")}},
			},

			// aria-pressed invalid (similar to aria-checked).
			{
				Code: `<div aria-pressed="partial" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-pressed")}},
			},
			{
				Code: `<div aria-pressed={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-pressed")}},
			},

			// aria-live invalid token.
			{
				Code: `<div aria-live="loud" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-live")}},
			},

			// aria-dropeffect — tokenlist, one unknown token in a multi-token list.
			{
				Code: `<div aria-dropeffect="copy fake" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-dropeffect")}},
			},

			// aria-colcount / aria-rowcount / aria-setsize / aria-posinset — integer.
			{
				Code: `<div aria-colcount="abc" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-colcount")}},
			},
			{
				Code: `<div aria-setsize={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-setsize")}},
			},

			// aria-valuenow / aria-valuemin — number.
			{
				Code: `<div aria-valuenow="abc" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuenow")}},
			},
			{
				Code: `<div aria-valuemin={false} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemin")}},
			},

			// aria-details / aria-errormessage — id (string).
			{
				Code: `<div aria-details={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-details")}},
			},
			{
				Code: `<div aria-errormessage={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-errormessage")}},
			},

			// aria-controls / aria-describedby — idlist (string).
			{
				Code: `<div aria-controls={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-controls")}},
			},
			{
				Code: `<div aria-describedby={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-describedby")}},
			},

			// Position assertions — confirm the diagnostic lands on the
			// JsxAttribute node, not the entire element.
			{
				Code: `<div aria-hidden="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidAriaPropType",
					Message:   upstreamErrorMessage("aria-hidden"),
					Line:      1, Column: 6,
				}},
			},
			{
				Code: `<div aria-label  =  "x"   aria-checked={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidAriaPropType",
					Message:   upstreamErrorMessage("aria-checked"),
					Line:      1, Column: 27,
				}},
			},
			// Multi-line — invalid attribute on line 2.
			{
				Code: "<div\n  aria-hidden=\"yes\"\n/>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidAriaPropType",
					Message:   upstreamErrorMessage("aria-hidden"),
					Line:      2, Column: 3,
				}},
			},

			// Nested JSX — invalid attrs at multiple depths each report.
			{
				Code: `<main><section><h2 aria-hidden="yes"><span aria-label={1} /></h2></section></main>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")},
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")},
				},
			},

			// Mixed valid + invalid attrs on a single element — only the
			// offending attribute reports.
			{
				Code: `<div aria-label="ok" aria-hidden="yes" aria-level={3} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")},
				},
			},

			// BigInt literal — `typeof 123n === "bigint"` matches no
			// ARIA type. integer / number / boolean / string / id / tristate
			// all reject; token / idlist / tokenlist need string. Report.
			{
				Code: `<div aria-hidden={123n} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},
			{
				Code: `<div aria-label={123n} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},

			// Library wrappers should not insulate the rule — invalid
			// values inside forwardRef / memo / useMemo / Suspense still
			// report.
			{
				Code: `const Btn = React.forwardRef<HTMLButtonElement, {}>((props, ref) => <button ref={ref} aria-pressed={1234} />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-pressed")},
				},
			},
			{
				Code: `<Suspense fallback={<div aria-busy="yes">loading</div>}><Page aria-label={1} /></Suspense>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-busy")},
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")},
				},
			},

			// Inside map callback — listener still fires per JSXAttribute.
			{
				Code: `function L({xs}) { return xs.map(x => <div aria-hidden="yes" key={x} />); }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")},
				},
			},

			// idlist with non-string value — boolean / number.
			{
				Code: `<div aria-labelledby={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-labelledby")}},
			},
			{
				Code: `<div aria-labelledby={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-labelledby")}},
			},

			// id type with non-string value.
			{
				Code: `<div aria-activedescendant={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-activedescendant")}},
			},
		})
}
