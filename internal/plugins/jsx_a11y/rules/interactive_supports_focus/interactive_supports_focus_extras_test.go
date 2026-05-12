//nolint:misspell // cspell:ignore utton idden resentation alse
package interactive_supports_focus

// This file is self-contained — it only exercises shapes OUTSIDE the
// upstream test file (tsgo↔ESTree AST quirks, position lock-ins,
// configuration edges, branches upstream doesn't test, real-world patterns).
// The upstream suite mirror lives in *_upstream_test.go. Helpers below are
// duplicated from there intentionally so each file can be read and audited
// independently — there is no cross-file reference.

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// extrasComponentsSettings mirrors the `<Div>` → `div` mapping that
// upstream rule tests use to keep the component-map path covered.
// Local to this file (duplicated from the upstream-mirror file so the
// extras suite is self-contained).
var extrasComponentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Div": "div",
		},
	},
}

// extrasPolymorphicSettings mirrors the `polymorphicPropName` setting other
// jsx-a11y rule ports exercise — `as="button"` resolves the JSX component
// to a native element, which then trips (or skips) the rule by the same
// paths a static lowercase tag would.
var extrasPolymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// extrasTabbableMessage / extrasFocusableMessage build the rule's two
// diagnostic message variants. Duplicated from the upstream-mirror file so
// extras can be audited without cross-file lookup.
func extrasTabbableMessage(role string) string {
	return fmt.Sprintf("Elements with the '%s' interactive role must be tabbable.", role)
}

func extrasFocusableMessage(role string) string {
	return fmt.Sprintf("Elements with the '%s' interactive role must be focusable.", role)
}

// extrasFocusableSuggestionsTwo is a compact constructor for the two-suggestion
// shape carried by every `focusable` diagnostic — keeps the test bodies tight.
func extrasFocusableSuggestionsTwo(outZero, outNegOne string) []rule_tester.InvalidTestCaseSuggestion {
	return []rule_tester.InvalidTestCaseSuggestion{
		{MessageId: "tabIndexZero", Output: outZero},
		{MessageId: "tabIndexNegOne", Output: outNegOne},
	}
}

// TestInteractiveSupportsFocusExtras covers shapes outside upstream's test
// file — tsgo↔ESTree AST differences, position assertions, configuration
// edges, and lock-ins for branches the upstream suite doesn't exercise.
func TestInteractiveSupportsFocusExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &InteractiveSupportsFocusRule,
		[]rule_tester.ValidTestCase{
			// ---- Custom (non-DOM) components — rule bails at the `dom.has(type)` gate.
			{Code: `<MyComp role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<X.Y role="button" onClick={() => {}} />`, Tsx: true},

			// ---- No mouse / keyboard handlers — rule bails at hasInteractiveProps.
			{Code: `<div role="button" onFocus={() => {}} />`, Tsx: true},
			{Code: `<div role="button" onBlur={() => {}} onScroll={() => {}} />`, Tsx: true},

			// ---- Disabled (HTML5 `disabled` resolves to non-undefined).
			{Code: `<div role="button" onClick={() => {}} disabled />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} disabled="" />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} disabled={false} />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} disabled={someVar} />`, Tsx: true},

			// ---- aria-disabled cases that count as disabled.
			{Code: `<div role="button" onClick={() => {}} aria-disabled />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} aria-disabled="true" />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} aria-disabled={true} />`, Tsx: true},

			// ---- aria-hidden disagreements: "true" (string) hides, anything else does not.
			{Code: `<div role="button" onClick={() => {}} aria-hidden="true" />`, Tsx: true},

			// ---- Presentation / none roles — IsPresentationRole bails the rule.
			{Code: `<div role="presentation" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="none" onClick={() => {}} />`, Tsx: true},

			// ---- role string-split: first valid role is non-interactive.
			//      "heading" is the first valid role — non-interactive → bail.
			{Code: `<div role="heading button" onClick={() => {}} />`, Tsx: true},

			// ---- Non-literal role expressions don't trip the interactive-role
			//      gate (LiteralStringValue returns false).
			{Code: `<div role={someRole} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={"button" + ""} onClick={() => {}} />`, Tsx: true},

			// ---- tabIndex set to a non-numeric expression — upstream getTabIndex
			//      step-2 returns the identifier name string → !== undefined.
			{Code: `<div role="button" onClick={() => {}} tabIndex={foo} />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} tabIndex={someFn()} />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} tabIndex={foo.bar} />`, Tsx: true},

			// ---- Inherently interactive element with role override — element
			//      already focusable so the rule bails at isInteractiveElement.
			{Code: `<button role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<a href="#" role="button" onClick={() => {}} />`, Tsx: true},

			// ---- Inherently non-interactive element with overlapping role —
			//      bails at isNonInteractiveElement (e.g. <h1 role="button">).
			{Code: `<h1 role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<nav role="button" onClick={() => {}} />`, Tsx: true},

			// ---- Polymorphic as resolves to interactive native element.
			{Code: `<Foo as="button" onClick={() => {}} />`, Tsx: true, Settings: extrasPolymorphicSettings},

			// ---- Triggering handler is set via literal-spread — upstream's
			//      hasAnyProp default `spreadStrict: true` treats this as
			//      opaque so the rule never enters. Equivalent to "no handler".
			{Code: `<div role="button" {...{onClick: () => {}}} />`, Tsx: true},

			// ---- Paired form (JsxElement → JsxOpeningElement listener path)
			//      vs self-closing form. Both must classify the same way.
			{Code: `<div role="button" tabIndex="0" onClick={() => {}}>label</div>`, Tsx: true},

			// ---- role lookup is case-insensitive at the rolesMap.has check —
			//      upstream `String(role).toLowerCase().split(' ')`. "Button" /
			//      "BUTTON" normalize to "button" and trip the rule. Locks
			//      [IsInteractiveRole]'s strings.ToLower path. (Listed as VALID
			//      here only when the element is already focusable, e.g. with
			//      tabIndex; the focus-required failure path is exercised in
			//      the invalid block below.)
			{Code: `<div role="BUTTON" onClick={() => {}} tabIndex="0" />`, Tsx: true},
			{Code: `<div role="Button" onClick={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- Multiple trigger handlers — element with tabIndex is fine.
			{Code: `<div role="button" onClick={() => {}} onMouseDown={() => {}} onKeyDown={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- role passed via literal-spread object. upstream
			//      `getProp(attributes, 'role')` default `spreadStrict: false`
			//      walks literal spreads, so this resolves to `role="button"`
			//      and tabIndex="0" already satisfies focus.
			{Code: `<div {...{role: "button"}} onClick={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- Multi-space / leading-trailing-space role — upstream's split(' ')
			//      filters empty entries via rolesMap.has, so "  button " resolves
			//      to first valid role "button" (interactive). Here element is
			//      already focusable, so it stays valid.
			{Code: `<div role=" button " onClick={() => {}} tabIndex="0" />`, Tsx: true},
			{Code: `<div role="button  link" onClick={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- Nested JSX where the OUTER element is focusable and inner
			//      is a non-interactive child — only the listener for each
			//      JsxOpeningElement fires independently.
			{Code: `<div role="button" tabIndex="0" onClick={() => {}}><span>inner</span></div>`, Tsx: true},

			// ---- role values that don't resolve to a valid interactive role.
			//      Upstream `String(value).toLowerCase().split(' ')` → filtered
			//      by rolesMap.has — empty / non-role tokens drop out, leaving
			//      0 valid roles → IsInteractiveRole = false → bail.
			{Code: `<div role="" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="not-a-real-role" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="custom widget" onClick={() => {}} />`, Tsx: true},

			// ---- Non-string role expressions (LITERAL_TYPES.noop → null).
			//      Identifier, conditional, logical, member, call → all noop
			//      under LITERAL_TYPES for the role-extractor path → IsInteractiveRole=false.
			{Code: `<div role={someRole} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={cond ? "button" : "link"} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={"button" || "link"} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={roles.primary} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={getRole()} onClick={() => {}} />`, Tsx: true},

			// ---- Numeric / boolean role values — getLiteralPropValue returns
			//      the JS primitive, String(it) doesn't match any role.
			{Code: `<div role={0} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={42} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={true} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={false} onClick={() => {}} />`, Tsx: true},
			// ---- Boolean attribute form `<div role />` — extractValue null-attr
			//      → true → String(true) = "true" → not a role.
			{Code: `<div role onClick={() => {}} />`, Tsx: true},

			// ---- Comments between attributes — JSX is whitespace-tolerant,
			//      the rule should not be confused by trivia tokens.
			{Code: `<div /* comment */ role="button" onClick={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- TypeScript type assertions inside the role JsxExpression —
			//      upstream's getLiteralPropValue does NOT strip TSAsExpression /
			//      TSNonNullExpression (it noop-s to null on those wrappers),
			//      so the role becomes null → not interactive → bail.
			{Code: `<div role={"button" as const} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={"button" as "button"} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={"button"!} onClick={() => {}} />`, Tsx: true},
			{Code: `<div role={"button" satisfies string} onClick={() => {}} />`, Tsx: true},

			// ---- Parens around the role expression — ESTree flattens parens
			//      at parse time, tsgo preserves them. LiteralPropStringValue
			//      strips them so both AST shapes yield the same outcome. This
			//      is an inherent-focus path (element already focusable), so
			//      the case must stay valid regardless of paren count.
			{Code: `<div role={("button")} onClick={() => {}} tabIndex="0" />`, Tsx: true},
			{Code: `<div role={(("button"))} onClick={() => {}} tabIndex="0" />`, Tsx: true},

			// ---- tabIndex shapes that upstream getTabIndex resolves to
			//      `!== undefined` and therefore satisfy `hasTabindex=true`.
			//      Locks HasUpstreamTabIndexValue against false negatives.
			//
			//   * `Math.random() > 0` is a BinaryExpression — upstream step-2
			//     getPropValue's BinaryExpression extractor synthesizes a
			//     non-empty string `${left} ${op} ${right}`. Truthy, !== undefined.
			//   * `Infinity` is a special Identifier — staticEval resolves to
			//     +Infinity number; jsValueToNumber → !== undefined.
			//
			//   NOTE: `tabIndex={null}` and `tabIndex={1.5}` are NOT here —
			//   upstream getTabIndex's step-1 literal arm explicitly returns
			//   undefined for both (null → "null" → NaN; 1.5 → !Number.isInteger).
			//   They live in the invalid block below.
			{Code: `<div role="button" onClick={() => {}} tabIndex={Math.random() > 0} />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} tabIndex={Infinity} />`, Tsx: true},

			// ---- Real-world wrappers — JSX inside Array.map, conditional
			//      expression, fragment. Listener should still fire on each
			//      JsxOpeningElement regardless of enclosing context.
			{Code: `const list = items.map(x => <div role="button" tabIndex="0" onClick={() => {}}>{x}</div>);`, Tsx: true},
			{Code: `const x = cond ? <div role="button" tabIndex="0" onClick={() => {}} /> : null;`, Tsx: true},
			{Code: `const x = <><div role="button" tabIndex="0" onClick={() => {}} /></>;`, Tsx: true},

			// ---- Inherent non-interactive elements (per
			//      strictNonInteractiveElementRoleSchemas) with interactive
			//      role + handler — rule bails at isNonInteractiveElement, no
			//      report. Locks the data-table.
			{Code: `<article role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<footer role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<li role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<p role="button" onClick={() => {}} />`, Tsx: true},

			// ---- `<header>` short-circuit — isNonInteractiveElement's
			//      `tagName === 'header'` early-returns false, so the rule
			//      treats `<header>` as "indeterminate" and DOES report when
			//      it carries an interactive role + handler. Listed in the
			//      invalid block below; here we just lock the bare
			//      `<header onClick={…} />` (no role) staying valid by virtue
			//      of IsInteractiveRole being false (no role attribute).
			{Code: `<header onClick={() => {}} />`, Tsx: true},

			// ---- `<td>` quirk — upstream `tagName !== 'td'` short-circuits
			//      both role-schema matchers, falls through to the AX schemas.
			//      `<td role="button" onClick={…} />` is therefore
			//      inherently-interactive via the elementAXObjects path, NOT
			//      non-interactive. Rule bails at isInteractiveElement.
			{Code: `<td role="button" onClick={() => {}} />`, Tsx: true},

			// ---- `aria-disabled="false"` is NOT disabled — upstream's
			//      LITERAL_TYPES.Literal coerces "false" → boolean false →
			//      `=== true` is false. Rule must NOT bail at IsDisabledElement.
			//      Locks the case where disabled=false combined with focusable
			//      tabIndex keeps the case valid.
			{Code: `<div role="button" onClick={() => {}} aria-disabled="false" tabIndex="0" />`, Tsx: true},

			// ---- HTML entity decoding on direct attribute strings. tsgo
			//      keeps the raw `&…;` source in StringLiteral.Text, but the
			//      @typescript-eslint / @babel JSX parser decodes entities
			//      before exposing the value to upstream. jsxa11yutil applies
			//      `jsxtransforms.DecodeEntities` on the direct-attribute
			//      branch of [LiteralStringValue] / [LiteralPropStringValue] /
			//      IsHiddenFromScreenReader to realign. Cases below are valid
			//      because the decoded role / type / aria-hidden / etc. routes
			//      the rule to a bail-out branch.
			//
			//   * `<input type="&#104;idden">` decodes to `type="hidden"` →
			//     hidden from screen reader → bail.
			//   * `<div aria-hidden="&#116;rue">` decodes to `aria-hidden="true"`
			//     → hidden → bail.
			//   * `<div role="&#112;resentation">` decodes to `role="presentation"`
			//     → presentation role → bail.
			//   * `<a role="button" href="&#35;">` decodes `href="#"` → still
			//     bails by being inherently interactive (`a[href]`).
			{Code: `<input type="&#104;idden" role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<div aria-hidden="&#116;rue" role="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="&#112;resentation" onClick={() => {}} />`, Tsx: true},
			{Code: `<a role="button" href="&#35;" onClick={() => {}} />`, Tsx: true},

			// ---- aria-disabled with entity-encoded "true" — must be treated
			//      as disabled. Locks [LiteralPropIsExactlyTrue]'s direct-
			//      attribute entity-decode path.
			{Code: `<div role="button" onClick={() => {}} aria-disabled="&#116;rue" />`, Tsx: true},

			// ---- polymorphic `as` with entity-encoded interactive element —
			//      `<Foo as="&#98;utton">` decodes to `as="button"`, which
			//      resolves the JSX component to a native <button> (inherently
			//      interactive → bail). Locks [polymorphicPropValue]'s direct-
			//      attribute entity-decode path.
			{
				Code:     `<Foo as="&#98;utton" onClick={() => {}} />`,
				Tsx:      true,
				Settings: extrasPolymorphicSettings,
			},

			// ---- Entity-encoded tabIndex direct attribute — tsgo keeps the
			//      raw `&…;` text on the StringLiteral, GetTabIndexEx /
			//      HasUpstreamTabIndexValue's direct-attribute branch already
			//      runs `jsxtransforms.DecodeEntities` so `tabIndex="&#48;"`
			//      decodes to "0" and `tabIndex="&#45;1"` to "-1", both
			//      counting as `!== undefined` → hasTabindex=true → bail.
			//      Differential-verified vs upstream.
			{Code: `<div role="button" onClick={() => {}} tabIndex="&#48;" />`, Tsx: true},
			{Code: `<div role="button" onClick={() => {}} tabIndex="&#45;1" />`, Tsx: true},

			// ---- Entity-encoded HTML5 `disabled` — upstream `getPropValue`
			//      decodes to "false", jsxAstUtilsLiteralCoerce → bool false,
			//      and `false !== undefined` so isHTML5Disabled=true → bail.
			//      Our [IsDisabledElement] mirrors via `!AttributeIsExplicitUndefined`
			//      — the entity-string isn't the JS undefined identifier, so
			//      we also classify as disabled. Differential-verified.
			{Code: `<div role="button" onClick={() => {}} disabled="&#102;alse" />`, Tsx: true},

			// ---- JsxExpression-wrapped string with raw entity text. Crucial
			//      negative case: JS string literals do NOT decode HTML entities
			//      (that's a JSX parser feature only), so `role={"&#98;utton"}`
			//      carries the literal string `&#98;utton`, which is not in the
			//      rolesMap → IsInteractiveRole = false → bail. Locks the
			//      directAttributeStringValue branch from over-reaching into
			//      JsxExpression wrappers. Differential-verified vs upstream.
			{Code: `<div role={"&#98;utton"} onClick={() => {}} />`, Tsx: true},

			// ---- Options edge: empty tabbable list → same as default → all
			//      diagnostics go down the focusable arm. The case below is
			//      valid because tabIndex="0" satisfies focus regardless of
			//      arm choice.
			{
				Code:    `<div role="button" onClick={() => {}} tabIndex="0" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tabbable": []interface{}{}},
			},
			// ---- Options edge: tabbable absent / nil / non-array — all
			//      collapse to the default (empty) list under GetOptionsMap +
			//      StringSliceOption nil fallback.
			{
				Code:    `<div role="button" onClick={() => {}} tabIndex="0" />`,
				Tsx:     true,
				Options: map[string]interface{}{},
			},
			{
				Code:    `<div role="button" onClick={() => {}} tabIndex="0" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tabbable": nil},
			},
			{
				Code:    `<div role="button" onClick={() => {}} tabIndex="0" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tabbable": "not-an-array"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Plain failure — fixed columns + suggestion output assertion.
			{
				Code: `<div role="button" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "focusable",
					Message:   extrasFocusableMessage("button"),
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "tabIndexZero", Output: `<div tabIndex={0} role="button" onClick={() => void 0} />`},
						{MessageId: "tabIndexNegOne", Output: `<div tabIndex={-1} role="button" onClick={() => void 0} />`},
					},
				}},
			},
			// ---- Locks the `tabbable` option branch — same code with the
			//      role in `tabbable` produces ONE suggestion and the
			//      "must be tabbable" message.
			{
				Code:    `<div role="button" onClick={() => void 0} />`,
				Tsx:     true,
				Options: map[string]interface{}{"tabbable": []interface{}{"button"}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "tabbable",
					Message:   extrasTabbableMessage("button"),
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "tabIndexZero", Output: `<div tabIndex={0} role="button" onClick={() => void 0} />`},
					},
				}},
			},
			// ---- Paired form (rule listens on JsxOpeningElement, but report
			//      anchors the opening element node — assert position points
			//      at the opening element, not the closing tag).
			{
				Code: `<div role="button" onClick={() => void 0}>label</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button" onClick={() => void 0}>label</div>`, `<div tabIndex={-1} role="button" onClick={() => void 0}>label</div>`),
				}},
			},
			// ---- Polymorphic `as` resolves to a div; role="button" + onClick
			//      should still trip.
			{
				Code:     `<Foo as="div" role="button" onClick={() => void 0} />`,
				Tsx:      true,
				Settings: extrasPolymorphicSettings,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<Foo tabIndex={0} as="div" role="button" onClick={() => void 0} />`, `<Foo tabIndex={-1} as="div" role="button" onClick={() => void 0} />`),
				}},
			},
			// ---- role given as NoSubstitutionTemplateLiteral — tsgo exposes
			//      `role={` button `}` as KindNoSubstitutionTemplateLiteral.
			//      LiteralStringValue must accept it for the interactive-role
			//      gate to fire.
			{
				Code: "<div role={`button`} onClick={() => void 0} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo("<div tabIndex={0} role={`button`} onClick={() => void 0} />", "<div tabIndex={-1} role={`button`} onClick={() => void 0} />"),
				}},
			},
			// ---- role expression is `{"button"}` — wrapped StringLiteral
			//      inside JsxExpression. Same outcome as the bare attribute
			//      string form.
			{
				Code: `<div role={"button"} onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role={"button"} onClick={() => void 0} />`, `<div tabIndex={-1} role={"button"} onClick={() => void 0} />`),
				}},
			},
			// ---- Multiple roles, first valid is interactive (button).
			{
				Code: `<div role="button heading" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button heading"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button heading" onClick={() => void 0} />`, `<div tabIndex={-1} role="button heading" onClick={() => void 0} />`),
				}},
			},
			// ---- Multi-line — diagnostic anchors at the opening element start.
			{
				Code: "<div\n  role=\"button\"\n  onClick={() => void 0}\n/>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo("<div tabIndex={0}\n  role=\"button\"\n  onClick={() => void 0}\n/>", "<div tabIndex={-1}\n  role=\"button\"\n  onClick={() => void 0}\n/>"),
				}},
			},
			// ---- onKeyDown is also a triggering handler — lock branch.
			{
				Code: `<div role="checkbox" onKeyDown={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("checkbox"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="checkbox" onKeyDown={() => void 0} />`, `<div tabIndex={-1} role="checkbox" onKeyDown={() => void 0} />`),
				}},
			},
			// ---- Element type comes from `components` mapping — Div → div, then div + role="button" + onClick trips.
			{
				Code:     `<Div role="button" onClick={() => void 0} />`,
				Tsx:      true,
				Settings: extrasComponentsSettings,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "focusable",
					Message:   extrasFocusableMessage("button"),
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "tabIndexZero", Output: `<Div tabIndex={0} role="button" onClick={() => void 0} />`},
						{MessageId: "tabIndexNegOne", Output: `<Div tabIndex={-1} role="button" onClick={() => void 0} />`},
					},
				}},
			},
			// ---- case-insensitive role lookup — upstream lower-cases before
			//      rolesMap.has. `<div role="BUTTON" onClick={…} />` should
			//      report (no tabIndex here).
			{
				Code: `<div role="BUTTON" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("BUTTON"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="BUTTON" onClick={() => void 0} />`, `<div tabIndex={-1} role="BUTTON" onClick={() => void 0} />`),
				}},
			},
			// ---- role via literal-spread, no tabIndex — upstream walks
			//      literal spread in `getProp`, so role resolves to "button".
			//      But hasAnyProp (spreadStrict: true) still sees `onClick` as
			//      a direct prop, so the rule fires.
			//
			//      NOTE: upstream `getLiteralPropValue` extraction over the
			//      spread-resolved PropertyAssignment also returns "button",
			//      so the diagnostic carries role="button" verbatim.
			{
				Code: `<div {...{role: "button"}} onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} {...{role: "button"}} onClick={() => void 0} />`, `<div tabIndex={-1} {...{role: "button"}} onClick={() => void 0} />`),
				}},
			},
			// ---- Template-literal `role` with substitutions — upstream's
			//      `getLiteralPropValue` (LITERAL_TYPES.TemplateExpression)
			//      synthesizes a concatenated string from quasi + substitution
			//      LITERAL_TYPES extractions. `` role={`button${""}`} `` should
			//      resolve to "button" and trip the rule. Locks the
			//      LiteralPropStringValue routing in [IsInteractiveRole].
			{
				Code: "<div role={`button${\"\"}`} onClick={() => void 0} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo("<div tabIndex={0} role={`button${\"\"}`} onClick={() => void 0} />", "<div tabIndex={-1} role={`button${\"\"}`} onClick={() => void 0} />"),
				}},
			},
			// ---- Three levels of nested JSX — outer / middle / inner each
			//      qualifying. Three diagnostics, each anchored to its own
			//      opening element. Locks no-bleed across the listener.
			{
				Code: "<div role=\"button\" onClick={() => void 0}>\n  <span role=\"menuitem\" onClick={() => void 0}>\n    <a role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusable",
						Message:   extrasFocusableMessage("button"),
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "tabIndexZero", Output: "<div tabIndex={0} role=\"button\" onClick={() => void 0}>\n  <span role=\"menuitem\" onClick={() => void 0}>\n    <a role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
							{MessageId: "tabIndexNegOne", Output: "<div tabIndex={-1} role=\"button\" onClick={() => void 0}>\n  <span role=\"menuitem\" onClick={() => void 0}>\n    <a role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
						},
					},
					{
						MessageId: "focusable",
						Message:   extrasFocusableMessage("menuitem"),
						Line:      2,
						Column:    3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "tabIndexZero", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span tabIndex={0} role=\"menuitem\" onClick={() => void 0}>\n    <a role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
							{MessageId: "tabIndexNegOne", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span tabIndex={-1} role=\"menuitem\" onClick={() => void 0}>\n    <a role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
						},
					},
					{
						MessageId: "focusable",
						Message:   extrasFocusableMessage("tab"),
						Line:      3,
						Column:    5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "tabIndexZero", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span role=\"menuitem\" onClick={() => void 0}>\n    <a tabIndex={0} role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
							{MessageId: "tabIndexNegOne", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span role=\"menuitem\" onClick={() => void 0}>\n    <a tabIndex={-1} role=\"tab\" onClick={() => void 0} />\n  </span>\n</div>"},
						},
					},
				},
			},
			// ---- JSX inside Array.map callback — listener fires for the
			//      inner element. Real-world pattern.
			{
				Code: `const list = items.map(x => <div role="button" onClick={() => void 0} />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      29,
					Suggestions: extrasFocusableSuggestionsTwo(`const list = items.map(x => <div tabIndex={0} role="button" onClick={() => void 0} />);`, `const list = items.map(x => <div tabIndex={-1} role="button" onClick={() => void 0} />);`),
				}},
			},
			// ---- JSX child of a Fragment — same outcome.
			{
				Code: `const x = <><div role="button" onClick={() => void 0} /></>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      13,
					Suggestions: extrasFocusableSuggestionsTwo(`const x = <><div tabIndex={0} role="button" onClick={() => void 0} /></>;`, `const x = <><div tabIndex={-1} role="button" onClick={() => void 0} /></>;`),
				}},
			},
			// ---- `tabIndex={null}` — upstream getLiteralPropValue → "null"
			//      string → Number("null") = NaN → step-1 undefined.
			//      hasTabindex=false → rule reports.
			{
				Code: `<div role="button" onClick={() => void 0} tabIndex={null} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button" onClick={() => void 0} tabIndex={null} />`, `<div tabIndex={-1} role="button" onClick={() => void 0} tabIndex={null} />`),
				}},
			},
			// ---- `tabIndex={1.5}` — upstream step-1 `Number.isInteger(1.5)`
			//      false → undefined. hasTabindex=false → reports.
			{
				Code: `<div role="button" onClick={() => void 0} tabIndex={1.5} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button" onClick={() => void 0} tabIndex={1.5} />`, `<div tabIndex={-1} role="button" onClick={() => void 0} tabIndex={1.5} />`),
				}},
			},
			// ---- `<header role="button" onClick={…} />` — upstream's
			//      isNonInteractiveElement returns false early for `header`
			//      (treated as indeterminate), so the rule's
			//      `!isNonInteractiveElement` arm is true and the case
			//      passes the bail-out checks → reports.
			{
				Code: `<header role="button" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<header tabIndex={0} role="button" onClick={() => void 0} />`, `<header tabIndex={-1} role="button" onClick={() => void 0} />`),
				}},
			},
			// ---- `aria-disabled="false"` is NOT disabled — `getLiteralPropValue`
			//      coerces "false" → boolean false → `=== true` fails. Without
			//      tabIndex the rule reports. Differential-verified vs upstream.
			{
				Code: `<div role="button" onClick={() => void 0} aria-disabled="false" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button" onClick={() => void 0} aria-disabled="false" />`, `<div tabIndex={-1} role="button" onClick={() => void 0} aria-disabled="false" />`),
				}},
			},
			// ---- `disabled={undefined}` — the ONLY shape that upstream's
			//      `getPropValue(disabledAttr) !== undefined` rejects. Rule
			//      doesn't bail at IsDisabledElement → reports. Counter-
			//      intuitive but differential-verified.
			{
				Code: `<div role="button" onClick={() => void 0} disabled={undefined} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button" onClick={() => void 0} disabled={undefined} />`, `<div tabIndex={-1} role="button" onClick={() => void 0} disabled={undefined} />`),
				}},
			},
			// ---- Template literal with non-Identifier non-Expression
			//      substitution: `` `button${1}` ``. upstream
			//      `extractValueFromTemplateLiteral` emits `""` for the
			//      numeric Literal substitution (only Identifier / *Expression
			//      kinds produce non-empty placeholders), so the role string
			//      ends up as "button" — first valid role "button" → reports.
			//      Differential-verified.
			{
				Code: "<div role={`button${1}`} onClick={() => void 0} />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo("<div tabIndex={0} role={`button${1}`} onClick={() => void 0} />", "<div tabIndex={-1} role={`button${1}`} onClick={() => void 0} />"),
				}},
			},
			// ---- Leading/trailing whitespace in role. upstream's report
			//      message uses the RAW role string (no trim), so " button "
			//      appears verbatim in the diagnostic. first valid role after
			//      split/filter is still "button" → reports.
			{
				Code: `<div role=" button " onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage(" button "),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role=" button " onClick={() => void 0} />`, `<div tabIndex={-1} role=" button " onClick={() => void 0} />`),
				}},
			},
			// ---- Double-space inside role — upstream split(' ') produces an
			//      empty token between, filter removes it via rolesMap.has,
			//      first valid stays "button" → reports. Message keeps raw text.
			{
				Code: `<div role="button  link" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button  link"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="button  link" onClick={() => void 0} />`, `<div tabIndex={-1} role="button  link" onClick={() => void 0} />`),
				}},
			},
			// ---- HTML entity in the role direct-attribute string. tsgo keeps
			//      raw `&…;` in StringLiteral.Text; LiteralPropStringValue's
			//      `directAttributeStringValue` branch runs
			//      `jsxtransforms.DecodeEntities` so the rule sees
			//      `role="button"` (decoded) and reports. Locks
			//      [LiteralPropStringValue] / [LiteralStringValue] entity-
			//      decoding alignment.
			//
			//      Note the diagnostic message carries the DECODED role string
			//      ("button"), matching upstream's behaviour where the parser
			//      decodes before the rule receives the value. The suggestion
			//      `Output` preserves the original raw source so the autofix
			//      doesn't reformat the user's entity choice.
			{
				Code: `<div role="&#98;utton" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="&#98;utton" onClick={() => void 0} />`, `<div tabIndex={-1} role="&#98;utton" onClick={() => void 0} />`),
				}},
			},
			// ---- Multi-role with entity encoding on each segment. Upstream's
			//      diagnostic message reflects the DECODED role string
			//      ("button  link") — entity decoding happens before the
			//      message template fills in `role`. Differential-verified.
			{
				Code: `<div role="&#98;utton  &#108;ink" onClick={() => void 0} />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId:   "focusable",
					Message:     extrasFocusableMessage("button  link"),
					Line:        1,
					Column:      1,
					Suggestions: extrasFocusableSuggestionsTwo(`<div tabIndex={0} role="&#98;utton  &#108;ink" onClick={() => void 0} />`, `<div tabIndex={-1} role="&#98;utton  &#108;ink" onClick={() => void 0} />`),
				}},
			},
			// ---- Nested JSX — outer AND inner both qualify. The listener
			//      fires once per JsxOpeningElement, so we see two diagnostics.
			//      Column points at each respective opening element. Each
			//      diagnostic carries its own pair of suggestions (independent
			//      fixer locations on the outer vs inner tag-name node).
			{
				Code: "<div role=\"button\" onClick={() => void 0}>\n  <span role=\"checkbox\" onClick={() => void 0} />\n</div>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "focusable",
						Message:   extrasFocusableMessage("button"),
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "tabIndexZero", Output: "<div tabIndex={0} role=\"button\" onClick={() => void 0}>\n  <span role=\"checkbox\" onClick={() => void 0} />\n</div>"},
							{MessageId: "tabIndexNegOne", Output: "<div tabIndex={-1} role=\"button\" onClick={() => void 0}>\n  <span role=\"checkbox\" onClick={() => void 0} />\n</div>"},
						},
					},
					{
						MessageId: "focusable",
						Message:   extrasFocusableMessage("checkbox"),
						Line:      2,
						Column:    3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "tabIndexZero", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span tabIndex={0} role=\"checkbox\" onClick={() => void 0} />\n</div>"},
							{MessageId: "tabIndexNegOne", Output: "<div role=\"button\" onClick={() => void 0}>\n  <span tabIndex={-1} role=\"checkbox\" onClick={() => void 0} />\n</div>"},
						},
					},
				},
			},
		},
	)
}

