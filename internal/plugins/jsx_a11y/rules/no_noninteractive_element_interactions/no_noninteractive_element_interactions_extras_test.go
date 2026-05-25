package no_noninteractive_element_interactions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This file holds every test case that is NOT a 1:1 mirror of upstream's
// own test file. Upstream-parity cases live in
// `no_noninteractive_element_interactions_upstream_test.go` so it stays
// trivially comparable against future upstream updates via diff.
//
// Goals here, beyond upstream coverage:
//
//   - Lock every classifier arm under inputs upstream's own tests skip:
//     attribute-value shapes (paren / TS `as` / TS `!` / member /
//     identifier / call / conditional / template literal), attribute-name
//     case-insensitivity, role / aria-hidden non-literal expressions,
//     contentEditable raw-match nuances, literal-spread vs non-literal
//     spread, multi-attribute first-match quirks.
//   - Lock per-element allow-list semantics across the FULL downstream
//     classifier chain (IsHiddenFromScreenReader, IsPresentationRole,
//     IsContentEditable, IsAbstract/Non*/Interactive*Role/Element) — a
//     bug here is invisible in upstream's tests because their preset
//     never allow-lists state attributes.
//   - Lock real-world component-wrapper patterns (forwardRef / memo /
//     HOC / hooks / class render / map / fragment / async / generator /
//     IIFE / conditional / `&&`) — each is an independent listener
//     invocation that must classify on its own.
//   - Lock settings paths upstream's tests don't exercise:
//     polymorphicPropName with empty / non-literal `as`, polymorphicAllowList
//     swap-and-no-swap, multi-entry components map.
//   - Position assertions on multi-line / nested / fragment so a future
//     refactor of listener registration can't silently shift columns.

// expectedErrorAnyPos mirrors expectedError with Line/Column suppressed —
// for nested / wrapped JSX where the opening element isn't at col 1.
var expectedErrorAnyPos = rule_tester.InvalidTestCaseError{
	MessageId: "noNoninteractiveElementInteractions",
	Message:   errorMessage,
}

// polymorphicSettings exercises `settings['jsx-a11y'].polymorphicPropName`.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// polymorphicAllowListSettings exercises `polymorphicAllowList` — only
// the named rawTypes get the `as`-driven swap.
var polymorphicAllowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName":  "as",
		"polymorphicAllowList": []interface{}{"Foo"},
	},
}

// componentsMapSettings exercises a multi-entry `components` map.
var componentsMapSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyArticle": "article",
			"MyButton":  "button",
			"MyDiv":     "div",
		},
	},
}

// TestNoNoninteractiveElementInteractionsExtras locks in branches that
// upstream's test file doesn't exercise but are reachable through the
// rule's listener gate.
func TestNoNoninteractiveElementInteractionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&NoNoninteractiveElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1 — attribute VALUE: nullish shapes.
			// PropValueIsNullish treats null / undefined / empty-{} as
			// "absent", so even a non-interactive element with a nullish
			// listener bails at the hasInteractiveProps gate.
			// ============================================================

			{Code: `<article onClick={null} />`, Tsx: true},
			{Code: `<article onClick={undefined} />`, Tsx: true},
			// ---- TS `as` cast unwrapped to null — extractValue while-loop
			//      strips TSAsExpression. ----
			{Code: `<article onClick={null as any} />`, Tsx: true},
			// ---- Paren-wrapped null / undefined — parens stripped before
			//      the nullish check. ----
			{Code: `<article onClick={(null)} />`, Tsx: true},
			{Code: `<article onClick={(undefined)} />`, Tsx: true},
			// ---- Multiple nullish attributes of the same name — getProp
			//      returns the FIRST match, so two nullish entries don't
			//      synthesize a non-nullish listener. ----
			{Code: `<article onClick={null} onClick={undefined} />`, Tsx: true},
			// ---- First-match quirk: upstream `getProp` returns the FIRST
			//      matching attribute; even though `onClick={fn}` follows,
			//      the FIRST onClick is null → getPropValue == null → the
			//      handler scan skips this name. Locks in the
			//      FindAttributeByName first-match semantics — a future
			//      refactor that walks "any non-null match" would silently
			//      flip this case. ----
			{Code: `<article onClick={null} onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — non-default handlers (default = focus + image
			// + keyboard + mouse). These groups are NOT in the default:
			// clipboard, composition, form, selection, touch, ui, wheel,
			// media-subset, animation, transition. Sampling a few exercises
			// the default-handler-set boundary.
			// ============================================================

			{Code: `<article onCopy={fn} />`, Tsx: true},
			{Code: `<article onCompositionStart={fn} />`, Tsx: true},
			{Code: `<article onChange={fn} />`, Tsx: true},
			{Code: `<article onSubmit={fn} />`, Tsx: true},
			{Code: `<article onTouchStart={fn} />`, Tsx: true},
			{Code: `<article onScroll={fn} />`, Tsx: true},
			{Code: `<article onWheel={fn} />`, Tsx: true},
			{Code: `<article onAnimationStart={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — contentEditable raw-match: VALID forms.
			// IsContentEditable matches `prop.value.raw === '"true"'`,
			// so only the bare `contentEditable="true"` string-literal
			// form bails.
			// ============================================================

			{Code: `<article contentEditable="true" onClick={fn} />`, Tsx: true},
			// ---- Case-insensitive attribute name (upstream `getProp`
			//      default `ignoreCase: true`). ----
			{Code: `<article contenteditable="true" onClick={fn} />`, Tsx: true},
			{Code: `<article CONTENTEDITABLE="true" onClick={fn} />`, Tsx: true},
			// ---- contentEditable on a div with non-interactive role —
			//      same bail. ----
			{Code: `<div role="article" contentEditable="true" onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — aria-hidden VALUE shapes that resolve to JS
			// boolean `true`. Locks the staticEval-based aria-hidden check.
			// ============================================================

			{Code: `<article aria-hidden onClick={fn} />`, Tsx: true},
			{Code: `<article aria-hidden={true} onClick={fn} />`, Tsx: true},
			{Code: `<article aria-hidden="true" onClick={fn} />`, Tsx: true},
			// ---- Paren-wrapped boolean — parens stripped. ----
			{Code: `<article aria-hidden={(true)} onClick={fn} />`, Tsx: true},
			{Code: `<article aria-hidden={((true))} onClick={fn} />`, Tsx: true},
			// ---- TS `as` cast — TSAsExpression unwrapped (extractValue
			//      while-loop). ----
			{Code: `<article aria-hidden={true as boolean} onClick={fn} />`, Tsx: true},
			// ---- aria-hidden as JsxExpression containing the case-insensitive
			//      string "true" — jsxAstUtilsLiteralCoerce maps to bool true. ----
			{Code: `<article aria-hidden={"true"} onClick={fn} />`, Tsx: true},
			// ---- Entity-encoded aria-hidden="&#116;rue" — direct-string
			//      path entity-decodes BEFORE coercion (jsxAstUtilsLiteralCoerce). ----
			{Code: `<article aria-hidden="&#116;rue" onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — role VALUE shapes that resolve to literal
			// "presentation" / "none". Locks LiteralPropStringValue across
			// paren / NoSubstitutionTemplate / TemplateExpression-with-literal-
			// substitution.
			// ============================================================

			{Code: `<article role="presentation" onClick={fn} />`, Tsx: true},
			{Code: `<article role="none" onClick={fn} />`, Tsx: true},
			{Code: `<article role={"presentation"} onClick={fn} />`, Tsx: true},
			// ---- Paren around the literal — stripped. ----
			{Code: `<article role={("presentation")} onClick={fn} />`, Tsx: true},
			// ---- NoSubstitutionTemplate (`role={\`none\`}`). ----
			{Code: "<article role={`none`} onClick={fn} />", Tsx: true},
			// ---- TemplateExpression whose quasis + literal substitution
			//      collapse to "presentation" — upstream LITERAL_TYPES
			//      walks each substitution as a literal extract. ----
			{Code: "<article role={`presentation${''}`} onClick={fn} />", Tsx: true},

			// ============================================================
			// Dimension 1 — IsAbstractRole arm. Locks the abstract-role
			// bail-out path on INHERENTLY non-interactive elements; the
			// `(!nonInt && !nonIntRole)` fallback returns false here
			// because `<img>` / `<p>` ARE inherently non-interactive
			// (matching the AX bare-schema). Upstream tests only the
			// `<div role="X">` shape where the no-opinion arm also bails;
			// these cases lock the IsAbstractRole branch on its own.
			// ============================================================

			{Code: `<img role="widget" onClick={fn} />`, Tsx: true},
			{Code: `<p role="window" onClick={fn} />`, Tsx: true},
			{Code: `<li role="composite" onClick={fn} />`, Tsx: true},
			{Code: `<article role="input" onClick={fn} />`, Tsx: true},
			// ---- The complete abstract-role set: command / composite /
			//      input / landmark / range / roletype / section /
			//      sectionhead / select / structure / widget / window. ----
			{Code: `<p role="command" onClick={fn} />`, Tsx: true},
			{Code: `<p role="landmark" onClick={fn} />`, Tsx: true},
			{Code: `<p role="range" onClick={fn} />`, Tsx: true},
			{Code: `<p role="roletype" onClick={fn} />`, Tsx: true},
			{Code: `<p role="section" onClick={fn} />`, Tsx: true},
			{Code: `<p role="sectionhead" onClick={fn} />`, Tsx: true},
			{Code: `<p role="select" onClick={fn} />`, Tsx: true},
			{Code: `<p role="structure" onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — spread shapes. Literal-spread is walked by
			// FindAttributeByName (matching upstream `getProp` default);
			// non-literal spread is opaque.
			// ============================================================

			// ---- Literal spread containing role="presentation" — found
			//      by IsPresentationRole. ----
			{Code: `<article {...{role: "presentation"}} onClick={fn} />`, Tsx: true},
			// ---- ShorthandPropertyAssignment containing role — the
			//      shorthand value is the bound Identifier, which the
			//      LITERAL_TYPES extractor rejects (Identifier non-undefined
			//      → null), so it doesn't match. The element is `<div>`
			//      (no opinion) so the rule bails on the no-opinion arm. ----
			{Code: `<div {...{role}} onClick={fn} />`, Tsx: true},
			// ---- Non-literal spread `{...props}` is opaque; element is
			//      `<div>` (no opinion) so the rule bails on the no-opinion
			//      arm even without seeing what props carry. ----
			{Code: `<div {...props} onClick={fn} />`, Tsx: true},
			// ---- Spread BEFORE direct listener — order doesn't matter. ----
			{Code: `<div {...props} onClick={fn} />`, Tsx: true},
			// ---- Multiple spreads + direct presentation role. ----
			{Code: `<article {...a} {...b} role="presentation" onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 1 — boolean attribute form on an INTERACTIVE
			// element (upstream `getPropValue(<X onClick/>)` resolves to
			// boolean true; `!= null` is true; but the inherent-interactive
			// gate bails). Locks that boolean form is recognized as
			// "interactive prop present".
			// ============================================================

			{Code: `<button onClick />`, Tsx: true},

			// ============================================================
			// Dimension 1 — `<input>` interactive promotion gates on the
			// bare `{Name: "input"}` AX schema, so EVERY input type bails.
			// Upstream tests the common types; this samples the rest.
			// ============================================================

			{Code: `<input type="datetime-local" onClick={fn} />`, Tsx: true},
			{Code: `<input type="week" onClick={fn} />`, Tsx: true},
			{Code: `<input type="file" onClick={fn} />`, Tsx: true},
			{Code: `<input type="password" onClick={fn} />`, Tsx: true},

			// ============================================================
			// Dimension 2 — nested JSX boundaries. Every opening element
			// is classified independently; the listener doesn't carry
			// state across the parent/child boundary.
			// ============================================================

			// ---- Outer + inner both inherently interactive. ----
			{Code: `<button onClick={fn}><a href="x" onClick={fn}>x</a></button>`, Tsx: true},
			// ---- Three-level all-interactive: form > button > a. ----
			{Code: `<fieldset><button onClick={fn}><a href="x" onClick={fn}>x</a></button></fieldset>`, Tsx: true},
			// ---- Outer non-interactive WITHOUT handler, inner interactive. ----
			{Code: `<article><button onClick={fn}>x</button></article>`, Tsx: true},

			// ============================================================
			// Dimension 2 — real-world component patterns producing
			// INTERACTIVE roots. Each pattern is its own listener fire.
			// ============================================================

			// ---- React.forwardRef wrapping interactive root. ----
			{
				Code: `const Btn = React.forwardRef((props, ref) => <button ref={ref} onClick={props.onClick} {...props} />);`,
				Tsx:  true,
			},
			// ---- React.memo wrapping interactive root. ----
			{
				Code: `const Btn = React.memo(({ id, onClick }) => <button id={id} onClick={onClick}>{id}</button>);`,
				Tsx:  true,
			},
			// ---- HOC wrapping interactive root. ----
			{
				Code: `const Enhanced = withTracking(({ value, onClick }) => <button data-value={value} onClick={onClick} />);`,
				Tsx:  true,
			},
			// ---- Custom hook returning interactive root. ----
			{Code: `function useThing() { return <button onClick={fn} />; }`, Tsx: true},
			// ---- Class render() returning interactive root. ----
			{
				Code: `class Form extends React.Component { render() { return <button onClick={fn}>x</button>; } }`,
				Tsx:  true,
			},
			// ---- Array.map producing interactive children. ----
			{
				Code: `const list = items.map(item => <button key={item.id} onClick={() => choose(item)}>{item.label}</button>);`,
				Tsx:  true,
			},
			// ---- Generator yielding interactive elements. ----
			{
				Code: `function* render() { yield <button onClick={fn} />; yield <a href="x" onClick={fn} />; }`,
				Tsx:  true,
			},
			// ---- Async function returning interactive element. ----
			{Code: `async function render() { return <button onClick={fn} />; }`, Tsx: true},
			// ---- IIFE returning interactive element. ----
			{Code: `const x = (() => <button onClick={fn} />)();`, Tsx: true},
			// ---- Fragment with all-interactive children. ----
			{
				Code: `const x = (<><button onClick={fn} /><a href="x" onClick={fn} /></>);`,
				Tsx:  true,
			},
			// ---- Conditional with both interactive branches. ----
			{
				Code: `const x = cond ? <button onClick={fn} /> : <a href="x" onClick={fn} />;`,
				Tsx:  true,
			},
			// ---- `&&` short-circuit. ----
			{Code: `const x = cond && <button onClick={fn} />;`, Tsx: true},

			// ============================================================
			// Dimension 3 — element-type resolution: settings paths.
			// ============================================================

			// ---- polymorphicPropName: `as` is a literal interactive name. ----
			{Code: `<Foo as="button" onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			{Code: `<Foo as="a" href="x" onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			{Code: `<Foo as="input" type="text" onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: empty-string `as=""` is falsy →
			//      fallback to rawType ("Foo", custom). ----
			{Code: `<Foo as="" onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: non-literal `as={x}` → fallback. ----
			{Code: `<Foo as={someType} onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: `as` absent → fallback. ----
			{Code: `<Foo onClick={fn} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicAllowList: Bar NOT in allow-list → no swap,
			//      stays as Bar (custom). ----
			{Code: `<Bar as="article" onClick={fn} />`, Tsx: true, Settings: polymorphicAllowListSettings},
			// ---- polymorphicAllowList: Foo in allow-list → swap to button. ----
			{Code: `<Foo as="button" onClick={fn} />`, Tsx: true, Settings: polymorphicAllowListSettings},

			// ---- components map: MyButton → button (inherent-interactive). ----
			{Code: `<MyButton onClick={fn} />`, Tsx: true, Settings: componentsMapSettings},
			// ---- components map present but element not listed → custom. ----
			{Code: `<Unmapped onClick={fn} />`, Tsx: true, Settings: componentsMapSettings},

			// ---- Compound element name `<Foo.Bar>` — not a DOM name → custom. ----
			{Code: `<Foo.Bar onClick={fn} />`, Tsx: true},

			// ============================================================
			// Options — `handlers` override.
			// ============================================================

			// ---- handlers=['onClick'] — onMouseDown is NOT in the override. ----
			{
				Code:    `<article onMouseDown={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{"onClick"}},
			},
			// ---- handlers=[] — empty list — NOTHING triggers. ----
			{
				Code:    `<article onClick={fn} onKeyDown={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{}},
			},
			// ---- handlers override pulling in a NON-default handler. ----
			{
				Code:    `<article onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{"onMouseDown"}},
			},
			// ---- Mixed-type handlers array: non-strings silently dropped
			//      (StringSliceOption defensive filter). Effective list
			//      is just ['onMouseDown'], so onClick doesn't trigger. ----
			{
				Code:    `<article onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{42, "onMouseDown", nil}},
			},

			// ============================================================
			// Options — per-element allow-list filters event handlers.
			// ============================================================

			// ---- `iframe: ['onLoad']` filters onLoad → no remaining
			//      interactive props → bail. ----
			{
				Code:    `<iframe onLoad={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"iframe": []interface{}{"onLoad"}},
			},
			// ---- Allow-list of a non-iframe element — `article: ['onClick']`
			//      filters onClick. ----
			{
				Code:    `<article onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"article": []interface{}{"onClick"}},
			},
			// ---- Allow-list does NOT touch a different element. The body
			//      key only affects <body>, not <article>; the rule reports
			//      <article onClick> normally — covered in invalid below. ----
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 1 — attribute VALUE: non-nullish shapes. Locks
			// every non-null branch of staticEval for the handler-value
			// scan. Each of these would `getPropValue != null`, so the
			// hasInteractiveProps gate trips.
			// ============================================================

			// ---- Boolean attribute form — extractValue null-attr-value
			//      maps to JS boolean `true`. ----
			{Code: `<article onClick />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Identifier value (non-undefined). ----
			{Code: `<article onClick={someFn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- MemberExpression value. ----
			{Code: `<article onClick={obj.method} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Optional-chain member access. ----
			{Code: `<article onClick={obj?.method} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- CallExpression value. ----
			{Code: `<article onClick={getHandler()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- ConditionalExpression value. ----
			{Code: `<article onClick={cond ? a : b} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- LogicalExpression value. ----
			{Code: `<article onClick={a || b} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- AsyncArrowFunction value. ----
			{Code: `<article onClick={async () => {}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- FunctionExpression value. ----
			{Code: `<article onClick={function() {}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Paren-wrapped function — staticEval strips parens. ----
			{Code: `<article onClick={(() => {})} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- TS `as` cast — TSAsExpression unwrapped to inner value. ----
			{Code: `<article onClick={fn as any} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- TS `!` non-null assertion. Upstream stringifies inner +
			//      `!` (e.g. `"fn!"` non-empty string) — `!= null` is true. ----
			{Code: `<article onClick={fn!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Dimension 1 — attribute NAME case-insensitivity. Upstream
			// `getProp` default `ignoreCase: true` — every case variation
			// of `onClick` matches.
			// ============================================================

			{Code: `<article onclick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<article ONCLICK={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<article onCLICK={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Dimension 1 — aria-hidden VALUE shapes that do NOT resolve
			// to boolean true → IsHiddenFromScreenReader returns false →
			// the rule continues to the report.
			// ============================================================

			// ---- aria-hidden={false}. ----
			{Code: `<article aria-hidden={false} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden="false" — jsxAstUtilsLiteralCoerce → bool false. ----
			{Code: `<article aria-hidden="false" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden={"false"}. ----
			{Code: `<article aria-hidden={"false"} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden Identifier (non-undefined) — staticEval
			//      synthesizes the identifier name as string, NOT boolean
			//      true. ----
			{Code: `<article aria-hidden={hiddenVar} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden CallExpression — not statically bool true. ----
			{Code: `<article aria-hidden={isHidden()} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden Ternary where both branches are NOT bool true —
			//      upstream's `extract` recursively evaluates the ternary
			//      (test Identifier "cond" is truthy → take consequent "0",
			//      a non-bool-true value). NOT hidden → reports. ----
			{Code: `<article aria-hidden={cond ? 0 : 1} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden TS `!` on bool literal — upstream's
			//      TSNonNullExpression arm stringifies the inner value and
			//      appends "!" → "true!" (non-empty string, NOT bool true).
			//      Locks the `OEKNonNullAssertions` exclusion from
			//      skipTransparent. ----
			{Code: `<article aria-hidden={true!} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Dimension 1 — role VALUE shapes that don't pass through any
			// bail-out arm: IsPresentationRole / IsNonInteractiveRole /
			// IsInteractiveRole / IsAbstractRole all require the role to
			// resolve to a literal string via LiteralPropStringValue.
			// ============================================================

			// ---- role Identifier — non-literal. ----
			{Code: `<article role={someRole} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role CallExpression — non-literal. ----
			{Code: `<article role={getRole()} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role ConditionalExpression — even with literal branches
			//      jsx-ast-utils' LITERAL_TYPES has no Conditional arm →
			//      null → not a literal string. ----
			{Code: `<article role={cond ? "presentation" : "article"} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role LogicalExpression — same noop → null. ----
			{Code: `<article role={a || "none"} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role TemplateExpression whose substitution doesn't
			//      collapse to a presentation/none/abstract/interactive
			//      string. ----
			{Code: "<article role={`status-${x}`} onClick={fn} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role="NONE" / role="PRESENTATION" — IsPresentationRole
			//      compares CASE-SENSITIVELY against `"presentation"` /
			//      `"none"` (upstream `=== 'presentation'`). ----
			{Code: `<article role="NONE" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<article role="PRESENTATION" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Dimension 1 — contentEditable raw-match nuances. The
			// upstream check is `prop.value.raw === '"true"'`. ANY shape
			// that doesn't produce exactly that raw string fails the
			// bail-out.
			// ============================================================

			// ---- JsxExpression form `={true}` — raw is `{true}`, not `"true"`. ----
			{Code: `<article contentEditable={true} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- JsxExpression containing the string "true" — raw is
			//      `{"true"}`, not the bare `"true"`. ----
			{Code: `<article contentEditable={"true"} onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- NoSubstitutionTemplate form — raw is a template, not
			//      the string literal. ----
			{Code: "<article contentEditable={`true`} onClick={fn} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Case-sensitive raw compare: "True" / "TRUE" don't match. ----
			{Code: `<article contentEditable="True" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<article contentEditable="TRUE" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Entity-encoded: raw text is NOT entity-decoded for the
			//      `=== '"true"'` compare. ----
			{Code: `<article contentEditable="&#116;rue" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- contentEditable="false" — explicit non-true. ----
			{Code: `<article contentEditable="false" onClick={fn} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- contentEditable in a LITERAL spread — upstream's
			//      `prop?.value?.raw` looks up the literal JsxAttribute
			//      shape; a PropertyAssignment inside a spread doesn't
			//      synthesize a `.value.raw`, so IsContentEditable returns
			//      false. ----
			{
				Code:   `<article {...{contentEditable: "true"}} onClick={fn} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Dimension 1 — spread shapes. Literal-spread of an event
			// handler IS walked by FindAttributeByName (upstream `getProp`
			// default).
			// ============================================================

			{
				Code:   `<article {...{onClick: fn}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- ShorthandPropertyAssignment `{onClick}` — getProp walks
			//      it (the value is the bound Identifier, but PROPERTY KEY
			//      `onClick` matches name). ----
			{
				Code:   `<article {...{onClick}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- Direct onClick + spread of literal aria-hidden={false} —
			//      spread reveals aria-hidden but as bool false, not true
			//      → not hidden → report. ----
			{
				Code:   `<article onClick={fn} {...{"aria-hidden": false}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Dimension 2 — same-tag nested boundary. Each opening
			// element classifies independently; the inner `<article>`
			// doesn't bleed past the outer's classification.
			// ============================================================

			// ---- Outer non-interactive WITHOUT handler → no report.
			//      Inner non-interactive WITH handler → report on the
			//      inner only. ----
			{
				Code: `<article>` + "\n" +
					`  <article onClick={fn} />` + "\n" +
					`</article>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noNoninteractiveElementInteractions",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},
			// ---- Both outer and inner non-interactive WITH handlers →
			//      two reports. ----
			{
				Code: `<article onClick={fn}>` + "\n" +
					`  <section onClick={fn} aria-label="x" />` + "\n" +
					`</article>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 1, Column: 1},
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 2, Column: 3},
				},
			},

			// ============================================================
			// Dimension 2 — position assertions: multi-line, fragment.
			// ============================================================

			// ---- Multi-line opening element — Line/Column anchor to `<`. ----
			{
				Code: `<article` + "\n" +
					`  onClick={fn}` + "\n" +
					`/>;`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noNoninteractiveElementInteractions",
					Message:   errorMessage,
					Line:      1,
					Column:    1,
				}},
			},
			// ---- Two same-line non-interactive elements both report. ----
			{
				Code: `<><article onClick={fn} /><section onClick={fn} aria-label="x" /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 1, Column: 3},
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 1, Column: 27},
				},
			},
			// ---- Nested non-interactive child inside interactive parent
			//      fires on its own opening element. ----
			{
				Code: `<button onClick={fn}>` + "\n" +
					`  <article onClick={fn} />` + "\n" +
					`</button>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noNoninteractiveElementInteractions",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},

			// ============================================================
			// Dimension 2 — real-world component patterns producing
			// NON-INTERACTIVE roots. Each is its own listener invocation.
			// ============================================================

			// ---- React.forwardRef wrapping non-interactive root. ----
			{
				Code:   `const Card = React.forwardRef((props, ref) => <article ref={ref} onClick={props.onClick} {...props} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- React.memo wrapping non-interactive root. ----
			{
				Code:   `const Pane = React.memo(({ id, onClick }) => <article id={id} onClick={onClick} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- HOC wrapping non-interactive root. ----
			{
				Code:   `const Enhanced = withTracking(({ value, onClick }) => <p data-value={value} onClick={onClick} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Custom hook returning non-interactive root. ----
			{
				Code:   `function useThing() { return <article onClick={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Class render() returning non-interactive root. ----
			{
				Code:   `class Pane extends React.Component { render() { return <article onClick={this.onClick}>x</article>; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Array.map producing non-interactive children. ----
			{
				Code:   `const list = items.map(item => <li key={item.id} onClick={fn}>{item.label}</li>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Generator yielding non-interactive elements (2 reports). ----
			{
				Code:   `function* render() { yield <article onClick={fn} />; yield <li onClick={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos, expectedErrorAnyPos},
			},
			// ---- Async function returning non-interactive root. ----
			{
				Code:   `async function render() { return <article onClick={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- IIFE returning non-interactive root. ----
			{
				Code:   `const x = (() => <article onClick={fn} />)();`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Conditional with one non-interactive branch. ----
			{
				Code:   `const x = cond ? <article onClick={fn} /> : null;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Fragment with mixed interactive / non-interactive
			//      children — only the non-interactive ones report. ----
			{
				Code:   `const x = (<><button onClick={fn} /><article onClick={fn} /><a href="x" onClick={fn} /><li onClick={fn} /></>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos, expectedErrorAnyPos},
			},
			// ---- Two-component file: A reports at line 1, B reports at line 2. ----
			{
				Code: `function A() { return <article onClick={fn} />; }` + "\n" +
					`function B() { return <section onClick={fn} aria-label="x" />; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 1},
					{MessageId: "noNoninteractiveElementInteractions", Message: errorMessage, Line: 2},
				},
			},

			// ============================================================
			// Dimension 3 — settings paths producing non-interactive roots.
			// ============================================================

			// ---- polymorphicPropName: <Foo as="article"> → article. ----
			{
				Code:     `<Foo as="article" onClick={fn} />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- polymorphicPropName: <Foo as="li">. ----
			{
				Code:     `<Foo as="li" onClick={fn} />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- polymorphicAllowList: Foo allowed → swap to "article". ----
			{
				Code:     `<Foo as="article" onClick={fn} />`,
				Tsx:      true,
				Settings: polymorphicAllowListSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- components map: <MyArticle> → article. ----
			{
				Code:     `<MyArticle onClick={fn} />`,
				Tsx:      true,
				Settings: componentsMapSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- components map: <MyDiv> → div (no opinion); BUT with a
			//      non-interactive role attribute, the role-driven
			//      classification kicks in. ----
			{
				Code:     `<MyDiv role="article" onClick={fn} />`,
				Tsx:      true,
				Settings: componentsMapSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Options — `handlers` override pulling in a non-default
			// handler.
			// ============================================================

			// ---- handlers=['onSubmit'] — onSubmit is not in default but
			//      explicit override pulls it in. ----
			{
				Code:    `<article onSubmit={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{"onSubmit"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- handlers + per-element allow-list combined. onClick is
			//      in handlers, NOT in iframe's allow-list → still reports. ----
			{
				Code: `<iframe onClick={fn} />`,
				Tsx:  true,
				Options: map[string]interface{}{
					"handlers": []interface{}{"onClick", "onLoad", "onError"},
					"iframe":   []interface{}{"onLoad", "onError"},
				},
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- Allow-list of a DIFFERENT element doesn't affect this
			//      one. `body: ['onLoad']` only applies to <body>; <article
			//      onClick> trips normally. ----
			{
				Code:    `<article onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"body": []interface{}{"onLoad"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Filter case-sensitivity quirk: upstream's `includes` uses
			// strict equality, but `getProp` is ignoreCase by default.
			// ----------------------------------------------------------
			// `<iframe ONLOAD={fn}/>` with `iframe: ['onLoad']`:
			//   - filter:  propName(attr) = "ONLOAD"; includes(['onLoad'], "ONLOAD") = false → KEPT.
			//   - has-prop: getProp(attrs, 'onLoad') ignoreCase → matches "ONLOAD" → trigger.
			// Different casing on the user's part DOES NOT escape the rule.
			// ============================================================

			{
				Code:    `<iframe ONLOAD={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"iframe": []interface{}{"onLoad"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Per-element allow-list filters EVERY downstream classifier
			// (not just the handler scan). When the allow-list strips a
			// state attribute, the corresponding classifier sees a
			// different shape and the rule's verdict changes.
			// ============================================================

			// ---- `article: ['aria-hidden']` strips aria-hidden BEFORE the
			//      IsHiddenFromScreenReader check, so the rule no longer
			//      sees the element as hidden and reports. Locks in the
			//      attrs-passthrough for IsHiddenFromScreenReader. ----
			{
				Code:    `<article aria-hidden onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"article": []interface{}{"aria-hidden"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},
			{
				Code:    `<article aria-hidden={true} onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"article": []interface{}{"aria-hidden"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- `article: ['role']` strips role BEFORE the
			//      IsPresentationRole / IsAbstractRole checks. ----
			{
				Code:    `<article role="presentation" onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"article": []interface{}{"role"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- `article: ['contentEditable']` strips contentEditable
			//      BEFORE the IsContentEditable check. ----
			{
				Code:    `<article contentEditable="true" onClick={fn} />`,
				Tsx:     true,
				Options: map[string]interface{}{"article": []interface{}{"contentEditable"}},
				Errors:  []rule_tester.InvalidTestCaseError{expectedError},
			},
		},
	)
}
