package click_events_have_key_events

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This file holds every test case that is NOT a 1:1 mirror of upstream's
// own test file. Upstream-parity cases live in
// `click_events_have_key_events_upstream_test.go` so it stays trivially
// comparable against future upstream updates via diff.
//
// Every case here is independently verified against upstream
// eslint-plugin-jsx-a11y@6.10.2 via differential validation, so behavior
// is byte-for-byte aligned — not just internally consistent.
//
// Coverage axes:
//
//   - Dimension 1 (AST node shape): tsgo's parenthesized / `as` /
//     `satisfies` wrappers on attribute values, JsxSelfClosingElement vs
//     paired JsxElement boundary, literal-spread vs non-literal-spread,
//     PropertyAssignment / ShorthandPropertyAssignment inside spreads.
//   - Dimension 2 (scoping / nesting): JSX inside HOC wrapper / forwardRef
//     / memo / hooks / class render / map / conditional / fragment /
//     async / generator / IIFE. Every JSX opening element is classified
//     independently.
//   - Dimension 4 (universal edge shapes): attribute existence forms
//     (boolean, `={null}`, `={undefined}`), case-insensitivity of
//     attribute names, role values via literal-extract paths (StringLiteral,
//     JsxExpression{StringLiteral}, NoSubstitutionTemplateLiteral,
//     TemplateExpression with substitutions = non-literal).
//   - Settings paths: components / polymorphicPropName /
//     polymorphicAllowList including edge cases upstream's tests don't
//     cover (polymorphic prop absent, empty string, multi-key map).
//   - Real-world component patterns: forwardRef / memo / HOC / hooks /
//     class / map / fragment / async / generator / IIFE. Locks the
//     listener gate against silent regressions when the rule or its
//     shared helpers are refactored.

// expectedErrorAnyPos is the standard expected diagnostic without a fixed
// line/column. Use when the JSX opening element is not at column 1 — e.g.
// inside a `function X() { return <div /> }` wrapper or `arr.map(...)`.
// The Line/Column == 0 fields are skipped by the rule_tester (see
// internal/rule_tester/rule_tester.go assertions).
var expectedErrorAnyPos = rule_tester.InvalidTestCaseError{
	MessageId: "clickEventsHaveKeyEvents",
	Message:   errorMessage,
}

// polymorphicSettings exercises `settings['jsx-a11y'].polymorphicPropName`.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// allowListSettings exercises polymorphicAllowList.
var allowListSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName":  "as",
		"polymorphicAllowList": []interface{}{"Foo"},
	},
}

// componentsMapSettings exercises multi-entry `components` map. Covers
// the "listed and remapped" branch and the "map present, key absent →
// rawType unchanged" branch.
var componentsMapSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyButton":  "button",
			"MyFooter":  "footer",
			"MyArticle": "article",
		},
	},
}

// TestClickEventsHaveKeyEventsExtras locks in branches that upstream's
// test file doesn't exercise but are reachable through the rule's
// listener gate. Each case is differential-validated against upstream
// eslint-plugin-jsx-a11y@6.10.2.
func TestClickEventsHaveKeyEventsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ClickEventsHaveKeyEventsRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: attribute-existence short-circuit edge shapes.
			// ============================================================

			// ---- onClick absent → no diagnostic, regardless of element. ----
			{Code: `<div />;`, Tsx: true},
			{Code: `<section />;`, Tsx: true},
			{Code: `<article aria-label="x" />;`, Tsx: true},
			// ---- onClick + onKeyPress only + spread combo. ----
			{Code: `<div onClick={() => void 0} onKeyPress={foo} {...props} />;`, Tsx: true},

			// ============================================================
			// Dimension 1 (tsgo AST): JsxSelfClosingElement vs paired
			// JsxElement listener boundary — both kinds must fire.
			// ============================================================

			// ---- Paired form <button>...</button>: tsgo emits KindJsxOpeningElement. ----
			{Code: `<button onClick={() => void 0}>label</button>`, Tsx: true},
			// ---- Self-closing form <button />: tsgo emits KindJsxSelfClosingElement. ----
			{Code: `<button onClick={() => void 0} />`, Tsx: true},
			// ---- Deeply nested mix of inherently-interactive elements. ----
			{
				Code: `<section aria-label="x">` +
					`<button onClick={() => void 0}>a</button>` +
					`<a href="x" onClick={() => void 0}>b</a>` +
					`<input type="text" onClick={() => void 0} />` +
					`</section>`,
				Tsx: true,
			},

			// ============================================================
			// Dimension 1 (tsgo AST): attribute-value wrappers on
			// aria-hidden / type. ESTree flattens parens at parse time;
			// tsgo preserves them. staticEval unwraps parens + the
			// TS-`as`-assertion form that jsx-ast-utils' `extract()`
			// while-loop also unwraps.
			// ============================================================

			// ---- aria-hidden with paren wrapping (single + multi level). ----
			{Code: `<div onClick={() => void 0} aria-hidden={(true)} />`, Tsx: true},
			{Code: `<div onClick={() => void 0} aria-hidden={((true))} />`, Tsx: true},
			// ---- aria-hidden with TS `as` cast (matches upstream's
			//      TSAsExpression unwrap in `extract()`). ----
			{Code: `<div onClick={() => void 0} aria-hidden={true as boolean} />`, Tsx: true},
			// ---- aria-hidden as JsxExpression containing string "true"
			//      (jsxAstUtilsLiteralCoerce: "true" → bool true). ----
			{Code: `<div onClick={() => void 0} aria-hidden={"true"} />`, Tsx: true},

			// ---- role as JsxExpression containing string literal. ----
			{Code: `<div onClick={() => void 0} role={"presentation"} />`, Tsx: true},
			// ---- role as no-substitution template literal. ----
			{Code: "<div onClick={() => void 0} role={`none`} />", Tsx: true},
			// ---- role with paren wrapping. ----
			{Code: `<div onClick={() => void 0} role={("presentation")} />`, Tsx: true},
			// ---- TemplateExpression whose quasis + extracted substitution
			//      values concatenate to "presentation". Upstream's
			//      LITERAL_TYPES.TemplateLiteral extractor joins each quasi
			//      with the extracted value of each substitution. Locks in
			//      the IsPresentationRole → LiteralPropStringValue routing
			//      (was LiteralStringValue, which silently rejected every
			//      TemplateExpression even when statically resolvable). ----
			{Code: "<div onClick={() => void 0} role={`presentation${''}`} />", Tsx: true},
			{Code: "<div onClick={() => void 0} role={`none${''}`} />", Tsx: true},
			// ---- role={null} — LITERAL_TYPES.Literal special-cases the
			//      `null` literal to the string "null"; "null" ≠
			//      "presentation"/"none" → IsPresentationRole returns false,
			//      so this falls through to the non-interactive `div` check
			//      and the rule fires. Covered in invalid section below. ----

			// ---- input[type=hidden] with JsxExpression containing string. ----
			{Code: `<input onClick={() => void 0} type={"hidden"} />`, Tsx: true},
			// ---- input[type=hidden] with case-insensitive value match
			//      (literalPropValue → "hidden" comparison via EqualFold). ----
			{Code: `<input onClick={() => void 0} type="HIDDEN" />`, Tsx: true},
			{Code: `<input onClick={() => void 0} type="Hidden" />`, Tsx: true},

			// ============================================================
			// Dimension 1: case-insensitive attribute name matching
			// (jsx-ast-utils `getProp` default `ignoreCase: true`).
			// ============================================================

			// ---- All-lowercase onclick / onkeydown. Upstream matches
			//      these against the canonical names. ----
			{Code: `<div onclick={() => void 0} onkeydown={foo} />`, Tsx: true},
			// ---- ALL-CAPS attribute names. ----
			{Code: `<div ONCLICK={() => void 0} ONKEYPRESS={foo} />`, Tsx: true},
			// ---- Mixed-case onClick + lowercase onkeypress. ----
			{Code: `<div onClick={() => void 0} onkeypress={foo} />`, Tsx: true},

			// ============================================================
			// Dimension 1: onClick value variants on interactive elements.
			// The rule gates on attribute existence, not value — these
			// would all report on a non-interactive element. Here they
			// pair with interactive tags so they land in the valid bucket.
			// ============================================================

			// ---- `<button onClick />` boolean form on interactive element. ----
			{Code: `<button onClick />`, Tsx: true},
			// ---- onClick={null} + onKeyDown on non-interactive element
			//      — null value irrelevant; keyboard listener saves it. ----
			{Code: `<div onClick={null} onKeyDown={foo} />`, Tsx: true},
			// ---- onClick={undefined} + onKeyDown on non-interactive element. ----
			{Code: `<div onClick={undefined} onKeyDown={foo} />`, Tsx: true},

			// ============================================================
			// Inherent-interactive element survey (upstream's tests cover
			// a small sample; these lock in the rest of the elementRoles /
			// elementAXObjects schema).
			// ============================================================
			{Code: `<input type="button" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="submit" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="reset" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="image" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="checkbox" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="radio" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="range" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="number" onClick={() => void 0} />`, Tsx: true},
			// input[type=search] matches the catch-all `{Name: "input"}`
			// entry in interactiveElementAXSchemas → interactive.
			{Code: `<input type="search" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input type="email" onClick={() => void 0} />`, Tsx: true},
			{Code: `<th onClick={() => void 0} />`, Tsx: true},
			{Code: `<td onClick={() => void 0} />`, Tsx: true},
			{Code: `<tr onClick={() => void 0} />`, Tsx: true},
			{Code: `<datalist onClick={() => void 0} />`, Tsx: true},
			{Code: `<menuitem onClick={() => void 0} />`, Tsx: true},
			{Code: `<summary onClick={() => void 0} />`, Tsx: true},

			// ============================================================
			// Settings: polymorphicPropName / polymorphicAllowList / components.
			// ============================================================

			// ---- polymorphicPropName: <Foo as="button"> → interactive. ----
			{Code: `<Foo as="button" onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: <Foo as="a" href="x"> → interactive <a href>. ----
			{Code: `<Foo as="a" href="x" onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: <Foo as="input" type="text"> → interactive. ----
			{Code: `<Foo as="input" type="text" onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: non-literal `as` falls back to rawType. ----
			{Code: `<Foo as={someType} onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: empty-string `as` falsy → fallback rawType. ----
			{Code: `<Foo as="" onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicPropName: `as` absent → fallback rawType (custom). ----
			{Code: `<Foo onClick={() => void 0} />`, Tsx: true, Settings: polymorphicSettings},
			// ---- polymorphicAllowList: <Bar as="div"> stays as Bar (custom). ----
			{Code: `<Bar as="div" onClick={() => void 0} />`, Tsx: true, Settings: allowListSettings},
			// ---- polymorphicAllowList: <Foo as="button"> swap allowed → interactive. ----
			{Code: `<Foo as="button" onClick={() => void 0} />`, Tsx: true, Settings: allowListSettings},

			// ---- components map: <MyButton> → button (interactive). ----
			{Code: `<MyButton onClick={() => void 0} />`, Tsx: true, Settings: componentsMapSettings},
			// ---- components map present but element not listed → custom. ----
			{Code: `<Unmapped onClick={() => void 0} />`, Tsx: true, Settings: componentsMapSettings},

			// ============================================================
			// Spread shapes — direct listener satisfies hasAnyProp.
			// ============================================================

			// ---- Direct onKeyDown + literal-spread of onKeyUp. ----
			{Code: `<div onClick={() => void 0} onKeyDown={foo} {...{onKeyUp: bar}} />`, Tsx: true},
			// ---- Spread before direct onKeyDown — order doesn't matter. ----
			{Code: `<div {...props} onClick={() => void 0} onKeyDown={foo} />`, Tsx: true},
			// ---- Multiple spreads + direct onClick + direct keyboard. ----
			{Code: `<div {...a} {...b} onClick={() => void 0} onKeyDown={foo} />`, Tsx: true},
			// ---- Reverse listener order — onKeyDown before onClick. ----
			{Code: `<div onKeyDown={foo} onClick={() => void 0} />`, Tsx: true},
			// ---- Multiple onClick attributes — first match suffices;
			//      keyboard listener present. ----
			{Code: `<div onClick={fn1} onClick={fn2} onKeyDown={foo} />`, Tsx: true},

			// ============================================================
			// Nested JSX: listener doesn't bleed past the boundary. Every
			// element classified independently.
			// ============================================================

			// ---- Outer div valid (has keyboard), inner button valid (interactive). ----
			{Code: `<div onClick={() => void 0} onKeyDown={foo}><button onClick={() => void 0}>x</button></div>`, Tsx: true},
			// ---- Three-level nesting, all interactive. ----
			{
				Code: `<fieldset><button onClick={() => void 0}>` +
					`<a href="x" onClick={() => void 0}>x</a>` +
					`</button></fieldset>`,
				Tsx: true,
			},

			// ============================================================
			// Real-world component patterns (no-report).
			// ============================================================

			// ---- React.forwardRef wrapping interactive button. ----
			{
				Code: `const Btn = React.forwardRef((props, ref) => <button ref={ref} onClick={props.onClick} {...props} />);`,
				Tsx:  true,
			},
			// ---- React.memo wrapping interactive button. ----
			{
				Code: `const Btn = React.memo(({ id, onClick }) => <button id={id} onClick={onClick}>{id}</button>);`,
				Tsx:  true,
			},
			// ---- Array.map producing interactive children only. ----
			{
				Code: `const list = items.map(item => <button key={item.id} onClick={() => choose(item)}>{item.label}</button>);`,
				Tsx:  true,
			},
			// ---- Custom hook returning JSX. ----
			{
				Code: `function useFancy() { return <button onClick={() => void 0} />; }`,
				Tsx:  true,
			},
			// ---- Generator yielding interactive elements. ----
			{
				Code: `function* render() { yield <button onClick={() => void 0} />; yield <a href="x" onClick={() => void 0} />; }`,
				Tsx:  true,
			},
			// ---- Async function returning interactive element. ----
			{
				Code: `async function render() { return <button onClick={() => void 0} />; }`,
				Tsx:  true,
			},
			// ---- Class component render() with interactive root. ----
			{
				Code: `class Form extends React.Component { render() { return <button onClick={() => void 0}>x</button>; } }`,
				Tsx:  true,
			},
			// ---- IIFE returning interactive element. ----
			{Code: `const x = (() => <button onClick={() => void 0} />)();`, Tsx: true},
			// ---- Fragment with mixed valid children. ----
			{
				Code: `const x = (<><button onClick={fn} /><a href="x" onClick={fn} /><div onClick={fn} onKeyDown={fn} /></>);`,
				Tsx:  true,
			},
			// ---- Conditional rendering — both branches interactive. ----
			{
				Code: `const x = cond ? <button onClick={fn} /> : <a href="x" onClick={fn} />;`,
				Tsx:  true,
			},
			// ---- `&&` short-circuit. ----
			{Code: `const x = cond && <button onClick={fn} />;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 1: onClick existence forms — boolean, null,
			// undefined values all count as "onClick present".
			// ============================================================

			// ---- Boolean form `<div onClick />` on non-interactive. ----
			{Code: `<div onClick />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- onClick={null} on non-interactive — value irrelevant. ----
			{Code: `<div onClick={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- onClick={undefined} on non-interactive. ----
			{Code: `<div onClick={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- All-lowercase `onclick` (case-insensitive) on non-interactive. ----
			{Code: `<div onclick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- ALL-CAPS `ONCLICK` on non-interactive. ----
			{Code: `<div ONCLICK={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Dimension 1: role / aria-hidden value shapes that don't
			// classify as the magic exempting literal.
			// ============================================================

			// ---- role="NONE" — case-sensitive match against
			//      "presentation" / "none"; "NONE" does NOT match. ----
			{Code: `<div onClick={() => void 0} role="NONE" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role="PRESENTATION" — same case-sensitivity. ----
			{Code: `<div onClick={() => void 0} role="PRESENTATION" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role as TemplateExpression with substitution (non-literal). ----
			{Code: "<div onClick={() => void 0} role={`presentation-${x}`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role as Identifier expression (non-literal). ----
			{Code: `<div onClick={() => void 0} role={someRole} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role as call expression (non-literal). ----
			{Code: `<div onClick={() => void 0} role={getRole()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role as conditional (non-literal under literalPropValue). ----
			{Code: `<div onClick={() => void 0} role={cond ? "presentation" : "main"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden as non-static call → not boolean true. ----
			{Code: `<div onClick={() => void 0} aria-hidden={isHidden()} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden as Identifier (non-static) → not boolean true. ----
			{Code: `<div onClick={() => void 0} aria-hidden={hiddenVar} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden={true!} (TS non-null assertion on a boolean
			//      literal). Upstream's `TSNonNullExpression` extractor
			//      stringifies the inner value and appends "!" → "true!"
			//      (a non-empty string, NOT bool true). aria-hidden !== true
			//      → element NOT hidden → rule fires. Aligned with upstream
			//      after the OEKNonNullAssertions strip was removed from
			//      `skipTransparent`; see the `case ast.KindNonNullExpression`
			//      arm in static_eval.go. ----
			{Code: `<div onClick={() => void 0} aria-hidden={true!} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Non-interactive DOM element survey (upstream tests cover
			// section/main/article/header/footer; these add the rest).
			// ============================================================
			{Code: `<span onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<aside onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<nav onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<p onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<h1 onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<h6 onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<ul onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<ol onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<li onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// `<form>` is in aria-query's `dom` map but matches NO
			// interactive nor non-interactive role schema (unless
			// aria-label / aria-labelledby / name is set, which still
			// resolves to non-interactive). IsInteractiveElement → false
			// → reported. Same for bare aria-label form.
			{Code: `<form onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<form aria-label="x" onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

			// ============================================================
			// Position assertions: nested JSX, multi-line containers.
			// ============================================================

			// ---- Nested non-interactive child fires on its own opening
			//      element. Locks in that the listener visits children. ----
			{
				Code: `<button onClick={() => void 0}>` + "\n" +
					`  <div onClick={() => void 0} />` + "\n" +
					`</button>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "clickEventsHaveKeyEvents",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},
			// ---- Multi-line opening element — Line/Column anchor to `<`. ----
			{
				Code: `<div` + "\n" +
					`  onClick={() => void 0}` + "\n" +
					`/>;`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "clickEventsHaveKeyEvents",
					Message:   errorMessage,
					Line:      1,
					Column:    1,
				}},
			},
			// ---- Two same-line non-interactive elements both report. ----
			{
				Code: `<><div onClick={fn} /><span onClick={fn} /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "clickEventsHaveKeyEvents", Message: errorMessage, Line: 1, Column: 3},
					{MessageId: "clickEventsHaveKeyEvents", Message: errorMessage, Line: 1, Column: 23},
				},
			},
			// ---- Three-level mix: outer valid (button), middle invalid
			//      (div), inner valid (a). ----
			{
				Code: `<button onClick={fn}>` + "\n" +
					`  <div onClick={fn}>` + "\n" +
					`    <a href="x" onClick={fn} />` + "\n" +
					`  </div>` + "\n" +
					`</button>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "clickEventsHaveKeyEvents",
					Message:   errorMessage,
					Line:      2,
					Column:    3,
				}},
			},

			// ============================================================
			// Settings: polymorphicPropName / components — DOM-resolved
			// element loses its custom-component exemption.
			// ============================================================

			// ---- polymorphicPropName: <Foo as="div"> → div → reported. ----
			{
				Code:     `<Foo as="div" onClick={() => void 0} />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- polymorphicPropName: <Foo as="section"> → non-interactive. ----
			{
				Code:     `<Foo as="section" onClick={() => void 0} />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- polymorphicAllowList: <Foo as="div"> — Foo allowed → swap → reported. ----
			{
				Code:     `<Foo as="div" onClick={() => void 0} />`,
				Tsx:      true,
				Settings: allowListSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- components map: <MyFooter> → footer (non-interactive). ----
			{
				Code:     `<MyFooter onClick={() => void 0} />`,
				Tsx:      true,
				Settings: componentsMapSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- components map: <MyArticle> → article (non-interactive). ----
			{
				Code:     `<MyArticle onClick={() => void 0} />`,
				Tsx:      true,
				Settings: componentsMapSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Spread shapes — literal-spread of onKeyDown is opaque under
			// hasAnyProp's spreadStrict default; spread cannot save the rule.
			// ============================================================

			// ---- Literal-spread `{...{onKeyDown: foo}}` does NOT satisfy
			//      hasAnyProp under default `spreadStrict: true`. ----
			{
				Code:   `<div onClick={() => void 0} {...{onKeyDown: foo}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- Spread before onClick: order doesn't change opacity. ----
			{
				Code:   `<div {...props} onClick={() => void 0} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- Multiple spreads + literal-spread of keyboard + direct onClick. ----
			{
				Code:   `<div {...a} {...b} {...{onKeyUp: fn}} onClick={() => void 0} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- Spread of literal containing onClick — getProp walks
			//      the literal, so the rule sees the onClick. No keyboard
			//      listener present → reported. ----
			{
				Code:   `<div {...{onClick: () => void 0}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- ShorthandPropertyAssignment in literal spread:
			//      `{...{onClick}}` — getProp matches the shorthand. ----
			{
				Code:   `<div {...{onClick}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// `<a>` / `<area>` interactive promotion gates on `href` ONLY.
			// ============================================================

			// ---- `<a>` with tabIndex={0} but NO href — non-interactive. ----
			{
				Code:   `<a onClick={() => void 0} tabIndex={0} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- `<a>` with role="button" but NO href. isInteractiveRole
			//      is NOT called by this rule, so the role doesn't promote. ----
			{
				Code:   `<a onClick={() => void 0} role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},
			// ---- `<area>` without href — non-interactive. ----
			{
				Code:   `<area onClick={() => void 0} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError},
			},

			// ============================================================
			// Real-world component patterns (reports). Use expectedErrorAnyPos
			// — the JSX opening element isn't at column 1 in these wrappers.
			// ============================================================

			// ---- React.forwardRef wrapping non-interactive div. ----
			{
				Code:   `const Card = React.forwardRef((props, ref) => <div ref={ref} onClick={props.onClick} {...props} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- React.memo wrapping non-interactive section. ----
			{
				Code:   `const Pane = React.memo(({ id, onClick }) => <section id={id} onClick={onClick} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- HOC wrapper producing non-interactive root. ----
			{
				Code:   `const Enhanced = withTracking(({ value, onClick }) => <div data-value={value} onClick={onClick} />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Array.map producing non-interactive children. ----
			{
				Code:   `const list = items.map(item => <li key={item.id} onClick={() => choose(item)}>{item.label}</li>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Class component render returning non-interactive root. ----
			{
				Code:   `class Form extends React.Component { render() { return <article onClick={this.onClick}>x</article>; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Generator yielding non-interactive elements. ----
			{
				Code:   `function* render() { yield <div onClick={fn} />; yield <article onClick={fn} />; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos, expectedErrorAnyPos},
			},
			// ---- Async function returning non-interactive root. ----
			{
				Code:   `async function render() { return <div onClick={fn} />; }`,
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
				Code:   `const x = cond ? <div onClick={fn} /> : null;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos},
			},
			// ---- Fragment with mixed interactive / non-interactive children
			//      — only non-interactive children report. ----
			{
				Code:   `const x = (<><button onClick={fn} /><div onClick={fn} /><a href="x" onClick={fn} /><section onClick={fn} /></>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrorAnyPos, expectedErrorAnyPos},
			},
			// ---- Two-component file: A reports at line 1, B reports at line 2. ----
			{
				Code: `function A() { return <div onClick={fn} />; }` + "\n" +
					`function B() { return <article onClick={fn} />; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "clickEventsHaveKeyEvents", Message: errorMessage, Line: 1},
					{MessageId: "clickEventsHaveKeyEvents", Message: errorMessage, Line: 2},
				},
			},
		},
	)
}
