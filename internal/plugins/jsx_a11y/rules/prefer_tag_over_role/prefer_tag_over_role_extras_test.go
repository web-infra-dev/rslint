// cspell:ignore heckbox eading

package prefer_tag_over_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// This file holds every test case that is NOT a 1:1 mirror of upstream's own
// test file. Upstream-parity cases live in `prefer_tag_over_role_upstream_test.go`
// so it stays trivially comparable against future upstream updates via diff.
//
// Cases are grouped by the Dimension 1–4 / semantic-walk axis they cover; the
// inline comments name the specific upstream branch / AST edge they protect.

// polymorphicSettings exercises the `polymorphicPropName` setting — the rule
// must consult the polymorphic prop (here `as`) to derive the effective HTML
// element name before consulting roleElements.tagNames for membership.
// Upstream's own test file doesn't cover this path.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// componentsSettings exercises the `components` settings map — `<MyButton>`
// resolves to nodeType "button" via the user-supplied component → tag map.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyButton": "button",
			"MyHeader": "header",
			"NotAnImg": "div",
		},
	},
}

// helper for invalid cases that aren't at line 1, column 1 (multi-line code,
// declaration-prefixed JSX).
func expectedErrAt(role, tag string, line, col int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preferTagOverRole",
		Message:   errorMessage(tag, role),
		Line:      line,
		Column:    col,
	}
}

// TestPreferTagOverRoleExtras locks in branches that upstream's test file
// doesn't exercise but are reachable through the rule's listener gate.
func TestPreferTagOverRoleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferTagOverRoleRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 1: AST shapes that DON'T resolve to a literal-typed
			//              string role → rule short-circuits before any
			//              roleElements lookup.
			// ============================================================

			// ---- No role attribute → rule never fires.
			//      Locks in upstream `if (!role) return;` arm 1. ----
			{Code: `<div />`, Tsx: true},
			{Code: `<span />`, Tsx: true},

			// ---- Boolean form `<div role />` — PropStaticStringValue sees a
			//      JsxAttribute with no Initializer → ("", false) → skip.
			//      Upstream would crash on `(true).lastIndexOf(' ')` for the
			//      boolean-attr extracted value; we are strictly safer. ----
			{Code: `<div role />`, Tsx: true},

			// ---- Empty string role — passes the !ok check (ok=true, str="")
			//      and trips the `rawRole == ""` short-circuit. Mirrors
			//      upstream `if (!propValue) return;`. ----
			{Code: `<div role="" />`, Tsx: true},
			{Code: `<div role={""} />`, Tsx: true},

			// ---- Trailing space — getLastPropValue returns "" via
			//      lastIndexOf+substring → roleElements.get("") undefined →
			//      skip. Locks in the boundary at i+1 == len(s). ----
			{Code: `<div role="checkbox " />`, Tsx: true},

			// ---- Identifier value (`role={x}`) — PropStaticStringValue
			//      synthesizes the identifier NAME as a string ("x"); "x"
			//      isn't a real ARIA role, so roleElements.get fails → skip.
			//      Locks the helper's "identifier-to-name-string" behavior
			//      against future refactors that might silently flip it. ----
			{Code: `<div role={someRole} />`, Tsx: true},
			{Code: `<button role={x} />`, Tsx: true},

			// ---- LogicalExpression `r || "checkbox"` — staticEval(r) returns
			//      the identifier name "r" (truthy string), so the OR short-
			//      circuits to "r"; "r" isn't a role → skip. Locks in
			//      staticEval's Logical short-circuit and the Identifier-as-
			//      name-string synthesis. ----
			{Code: `<div role={r || "checkbox"} />`, Tsx: true},
			// ---- CallExpression / MemberExpression role — staticEval emits
			//      stable placeholders ("(call)" / "(member)") that are not
			//      ARIA roles → skip. Locks in the synthesized-string shape;
			//      a future change that lets these resolve to a real role
			//      would unexpectedly start reporting. ----
			{Code: `<div role={getRole()} />`, Tsx: true},
			{Code: `<div role={config.role} />`, Tsx: true},

			// ---- Non-string literal under a JsxExpression — boolean, null,
			//      undefined, number all fall outside the jvString classification
			//      and are silently skipped. Upstream's getPropValue would
			//      return these and then `.lastIndexOf` would throw; we are
			//      strictly safer.
			//      ----
			{Code: `<div role={true} />`, Tsx: true},
			{Code: `<div role={false} />`, Tsx: true},
			{Code: `<div role={null} />`, Tsx: true},
			{Code: `<div role={undefined} />`, Tsx: true},
			{Code: `<div role={42} />`, Tsx: true},

			// ---- Role names that aren't in roleElements.
			//      Locks upstream's `if (!matchedTagsSet) return;` arm. ----
			{Code: `<div role="unknown" />`, Tsx: true},
			{Code: `<div role="not-a-real-role" />`, Tsx: true},
			// Multi-token where the LAST token isn't a real role.
			{Code: `<div role="checkbox unknown" />`, Tsx: true},
			// Abstract roles are in aria-query's rolesMap but NOT in
			// roleElementMap (their concept lists are empty after the
			// HTML-only filter). roleElements.get("command") → undefined → skip.
			{Code: `<div role="command" />`, Tsx: true},
			{Code: `<div role="widget" />`, Tsx: true},
			{Code: `<div role="composite" />`, Tsx: true},

			// ============================================================
			// Dimension 4 / Semantic walk: element already IS the role's
			//              semantic tag → upstream's `.some()` returns true
			//              → skip.
			// ============================================================

			// ---- heading: every h1–h6 should pass (locks in the multi-tag
			//      `.some()` walk, not just the first entry). ----
			{Code: `<h1 role="heading" />`, Tsx: true},
			{Code: `<h2 role="heading" />`, Tsx: true},
			{Code: `<h3 role="heading" />`, Tsx: true},
			{Code: `<h4 role="heading" />`, Tsx: true},
			{Code: `<h5 role="heading" />`, Tsx: true},
			{Code: `<h6 role="heading" />`, Tsx: true},

			// ---- rowgroup: tbody / tfoot / thead all valid. ----
			{Code: `<tbody role="rowgroup" />`, Tsx: true},
			{Code: `<tfoot role="rowgroup" />`, Tsx: true},
			{Code: `<thead role="rowgroup" />`, Tsx: true},

			// ---- link: a / area both valid. The match is on NAME only —
			//      upstream's matchedTags.some(tag.name === elementType)
			//      ignores the role concept's attribute requirements (the
			//      `href` constraint is purely a formatTag detail). So
			//      `<a role="link" />` is valid even WITHOUT href; locks
			//      in that asymmetric behavior. ----
			{Code: `<a role="link" />`, Tsx: true},
			{Code: `<area role="link" />`, Tsx: true},

			// ---- banner / header parity: header is the only tag in the
			//      banner role's mapping. ----
			{Code: `<header role="banner" />`, Tsx: true},

			// ---- generic role: one entry per name — pick a few spread across
			//      the list. ----
			{Code: `<div role="generic" />`, Tsx: true},
			{Code: `<span role="generic" />`, Tsx: true},
			{Code: `<body role="generic" />`, Tsx: true},
			{Code: `<section role="generic" />`, Tsx: true},

			// ---- list role: menu / ol / ul all valid; named match,
			//      attribute-free. ----
			{Code: `<menu role="list" />`, Tsx: true},
			{Code: `<ol role="list" />`, Tsx: true},
			{Code: `<ul role="list" />`, Tsx: true},

			// ---- term role: dfn / dt both valid. ----
			{Code: `<dfn role="term" />`, Tsx: true},
			{Code: `<dt role="term" />`, Tsx: true},

			// ---- listbox: select OR datalist. ----
			{Code: `<select role="listbox" />`, Tsx: true},
			{Code: `<datalist role="listbox" />`, Tsx: true},

			// ---- textbox: input OR textarea — locks in the OR-arm. ----
			{Code: `<input role="textbox" />`, Tsx: true},
			{Code: `<textarea role="textbox" />`, Tsx: true},

			// ---- combobox: input OR select — locks in the OR-arm. ----
			{Code: `<input role="combobox" />`, Tsx: true},
			{Code: `<select role="combobox" />`, Tsx: true},

			// ---- group: details / fieldset / optgroup / address — locks in
			//      the 4-arm OR walk for completeness. ----
			{Code: `<details role="group" />`, Tsx: true},
			{Code: `<fieldset role="group" />`, Tsx: true},
			{Code: `<optgroup role="group" />`, Tsx: true},
			{Code: `<address role="group" />`, Tsx: true},

			// ============================================================
			// Dimension 4 / Semantic walk: polymorphicPropName + components.
			// ============================================================

			// ---- `<Box as="input" role="checkbox" />` — elementType resolves
			//      to "input" via polymorphicPropName=as → matches → valid.
			//      Locks in that GetElementType runs BEFORE the tagNames
			//      check. ----
			{Code: `<Box as="input" role="checkbox" />`, Tsx: true, Settings: polymorphicSettings},

			// ---- components map: `<MyHeader role="banner" />` resolves to
			//      tag "header" → header is in banner.tagNames → valid. ----
			{Code: `<MyHeader role="banner" />`, Tsx: true, Settings: componentsSettings},
			// ---- components map: `<MyButton role="button" />` resolves to
			//      tag "button" → button.tagNames includes "button" → valid.
			//      Without the settings the same code would REPORT (PascalCase
			//      element-name fails the tagNames check); the inverse-case
			//      `<MyComponent role="link" />` lives in invalid below. ----
			{Code: `<MyButton role="button" />`, Tsx: true, Settings: componentsSettings},

			// ============================================================
			// Dimension 1: case-sensitivity on the role-name comparison.
			// ============================================================

			// ---- roleElements lookup is case-SENSITIVE (matches upstream
			//      `roleElements.get(role)` — JS Map keys are case-sensitive).
			//      `<div role="CHECKBOX" />` therefore fails the lookup →
			//      skip. Upstream's own test file doesn't cover this. ----
			{Code: `<div role="CHECKBOX" />`, Tsx: true},
			{Code: `<div role="Heading" />`, Tsx: true},

			// ============================================================
			// Dimension 4: empty role / whitespace-only role.
			// ============================================================

			// ---- Whitespace-only role string — getLastPropValue:
			//      lastIndexOf(' ') = len-1 → substring(len) = "" → skip. ----
			{Code: `<div role=" " />`, Tsx: true},
			{Code: `<div role="   " />`, Tsx: true},

			// ============================================================
			// Listener gate: PAIRED vs self-closing — both should be reached.
			// (Self-closing is covered throughout; verify paired-no-trigger
			// for an element on the role's tagNames list.)
			// ============================================================

			// ---- Paired-form `<header role="banner"></header>` — matches
			//      header → valid. Lock the gate for the paired
			//      KindJsxOpeningElement listener arm. ----
			{Code: `<header role="banner"></header>`, Tsx: true},

			// ============================================================
			// Dimension 4: explicit role on the element's own semantic tag
			// across all single-tag mappings (regression cordon — locks in
			// every single-tag role's tagNames against future table edits).
			// ============================================================
			{Code: `<article role="article" />`, Tsx: true},
			{Code: `<blockquote role="blockquote" />`, Tsx: true},
			{Code: `<caption role="caption" />`, Tsx: true},
			{Code: `<td role="cell" />`, Tsx: true},
			{Code: `<code role="code" />`, Tsx: true},
			{Code: `<th role="columnheader" />`, Tsx: true},
			{Code: `<aside role="complementary" />`, Tsx: true},
			{Code: `<footer role="contentinfo" />`, Tsx: true},
			{Code: `<dd role="definition" />`, Tsx: true},
			{Code: `<del role="deletion" />`, Tsx: true},
			{Code: `<dialog role="dialog" />`, Tsx: true},
			{Code: `<html role="document" />`, Tsx: true},
			{Code: `<em role="emphasis" />`, Tsx: true},
			{Code: `<figure role="figure" />`, Tsx: true},
			{Code: `<form role="form" />`, Tsx: true},
			{Code: `<td role="gridcell" />`, Tsx: true},
			{Code: `<ins role="insertion" />`, Tsx: true},
			{Code: `<li role="listitem" />`, Tsx: true},
			{Code: `<main role="main" />`, Tsx: true},
			{Code: `<mark role="mark" />`, Tsx: true},
			{Code: `<math role="math" />`, Tsx: true},
			{Code: `<meter role="meter" />`, Tsx: true},
			{Code: `<nav role="navigation" />`, Tsx: true},
			{Code: `<option role="option" />`, Tsx: true},
			{Code: `<p role="paragraph" />`, Tsx: true},
			{Code: `<img role="presentation" />`, Tsx: true},
			{Code: `<progress role="progressbar" />`, Tsx: true},
			{Code: `<input role="radio" />`, Tsx: true},
			{Code: `<section role="region" />`, Tsx: true},
			{Code: `<tr role="row" />`, Tsx: true},
			{Code: `<th role="rowheader" />`, Tsx: true},
			{Code: `<input role="searchbox" />`, Tsx: true},
			{Code: `<hr role="separator" />`, Tsx: true},
			{Code: `<input role="slider" />`, Tsx: true},
			{Code: `<input role="spinbutton" />`, Tsx: true},
			{Code: `<output role="status" />`, Tsx: true},
			{Code: `<strong role="strong" />`, Tsx: true},
			{Code: `<sub role="subscript" />`, Tsx: true},
			{Code: `<sup role="superscript" />`, Tsx: true},
			{Code: `<table role="table" />`, Tsx: true},
			{Code: `<time role="time" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Dimension 4: role values inside `{…}` JsxExpression — direct
			//              literal extraction.
			// ============================================================

			// ---- `<div role={"checkbox"} />` — JsxExpression wrapping a
			//      StringLiteral. PropStaticStringValue must reach into the
			//      expression and return "checkbox"; upstream extracts via
			//      `extractValue` for the JSXExpressionContainer arm. ----
			{
				Code:   `<div role={"checkbox"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ---- NoSubstitutionTemplateLiteral form — same extraction path.
			//      Lock in to confirm the helper doesn't restrict to
			//      KindStringLiteral. ----
			{
				Code:   "<div role={`checkbox`} />",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ---- BinaryExpression resolvable to a literal — `"check" + "box"`
			//      → "checkbox" via staticEval's PlusToken arm. Upstream's
			//      getPropValue routes through the same string concatenation
			//      coercion. ----
			{
				Code:   `<div role={"check" + "box"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Dimension 4: case-insensitive role-ATTRIBUTE-name lookup.
			// ============================================================

			// ---- `<div ROLE="checkbox" />` — FindAttributeByName uses
			//      EqualFold, mirroring jsx-ast-utils' default `ignoreCase:
			//      true`. Locks in case-insensitive attribute-name matching. ----
			{
				Code:   `<div ROLE="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Dimension 4: Spread-attribute literal-spread path.
			// ============================================================

			// ---- Literal-object spread carries the role attribute.
			//      FindAttributeByName's literal-spread arm walks the inner
			//      ObjectLiteralExpression and matches `role`. ----
			{
				Code:   `<div {...{role:"checkbox"}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Listener gate: PAIRED form `<X role="...">…</X>` — tsgo splits
			//                into KindJsxOpeningElement + children, separate
			//                from the self-closing case. Both arms must fire.
			// ============================================================

			// ---- Paired form on a div → report on the OPENING element. ----
			{
				Code:   `<div role="checkbox"></div>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ---- Paired form on a custom component → report. ----
			{
				Code:   `<MyComponent role="heading">child</MyComponent>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},

			// ============================================================
			// Dimension 4: polymorphicPropName + components — element-name
			//              resolution must run BEFORE the tagNames check.
			// ============================================================

			// ---- `<Box as="div" role="checkbox" />` → elementType = "div"
			//      via polymorphicPropName=as → not in checkbox.tagNames →
			//      report. ----
			{
				Code:     `<Box as="div" role="checkbox" />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ---- components map: `<NotAnImg role="img" />` resolves to "div"
			//      → img.tagNames is just ["img"] → report. Locks
			//      the components map being consulted by GetElementType. ----
			{
				Code:     `<NotAnImg role="img" />`,
				Tsx:      true,
				Settings: componentsSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedErr("img", `<img alt=...>, or <img alt=...>`)},
			},

			// ---- PascalCase custom element (no settings) — elementType =
			//      "MyComponent" → not in link.tagNames → report. ----
			{
				Code:   `<MyComponent role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("link", `<a href=...>, or <area href=...>`)},
			},

			// ---- Namespaced JSX (`<svg:foo>`) — elementType = "svg:foo" →
			//      not in any tagNames → report. Locks in the GetElementType
			//      passthrough for namespaced names. ----
			{
				Code:   `<svg:foo role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Dimension 4: multi-token role — LAST-token semantics across
			//              several shapes.
			// ============================================================

			// ---- 3 tokens — verify lastIndexOf-based extraction (NOT
			//      first-match). Upstream `getLastPropValue` only consults
			//      the last token. ----
			{
				Code:   `<div role="alert button checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Double-space between tokens — still tokenized on single
			//      ASCII space; lastIndexOf finds the LAST space, so the
			//      empty inner-token is irrelevant. Last token = "checkbox". ----
			{
				Code:   `<div role="button  checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Leading space — last token is still the unique role token. ----
			{
				Code:   `<div role=" checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Dimension 4: roles whose tagNames list has exactly one entry —
			//              regression cordon against table edits. Pick a
			//              representative subset of single-tag roles not
			//              covered by upstream's invalid table.
			// ============================================================
			{
				Code:   `<div role="article" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("article", `<article>`)},
			},
			{
				Code:   `<div role="paragraph" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("paragraph", `<p>`)},
			},
			{
				Code:   `<div role="navigation" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("navigation", `<nav>`)},
			},
			{
				Code:   `<div role="separator" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("separator", `<hr>`)},
			},
			{
				Code:   `<div role="time" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("time", `<time>`)},
			},

			// ============================================================
			// Dimension 4: roles with 2 tagNames — locks in the `, or `
			//              join shape and exact ordering.
			// ============================================================
			{
				Code:   `<div role="term" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("term", `<dfn>, or <dt>`)},
			},

			// ============================================================
			// Dimension 4: roles where formatTag emits a `value` for the
			//              first attribute — locks in the
			//              `"\"${value}\""` branch vs the `...` fallback.
			// ============================================================
			{
				Code:   `<div role="radio" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("radio", `<input type="radio">`)},
			},
			{
				Code:   `<div role="slider" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("slider", `<input type="range">`)},
			},
			{
				Code:   `<div role="spinbutton" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("spinbutton", `<input type="number">`)},
			},

			// ============================================================
			// Dimension 4: roles where the FIRST attribute is value-less
			//              (formatTag emits `...`). Locks in the long
			//              upstream concat for textbox / combobox / etc.
			// ============================================================
			{
				Code:   `<div role="textbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("textbox", `<input type=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <textarea>`)},
			},
			{
				Code:   `<div role="combobox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("combobox", `<input list=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, <input list=...>, or <select multiple=...>`)},
			},
			{
				Code:   `<div role="listbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("listbox", `<select size=...>, <select multiple=...>, or <datalist>`)},
			},
			{
				Code:   `<div role="searchbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("searchbox", `<input list=...>`)},
			},
			// ---- generic role: 19 single-name tags → the longest concat. ----
			{
				Code:   `<table role="generic" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("generic", `<a>, <area>, <aside>, <b>, <bdo>, <body>, <data>, <div>, <footer>, <header>, <hgroup>, <i>, <pre>, <q>, <samp>, <section>, <small>, <span>, or <u>`)},
			},

			// ============================================================
			// tsgo-specific AST shapes (no ESTree analog): TS assertion
			// wrappers, parenthesized values, NonNull `!` — staticEval's
			// skipTransparent strips `( )` + `as` + `<T>x` but PRESERVES
			// `!` (NonNullExpression) and `satisfies`, matching upstream
			// jsx-ast-utils' getPropValue.
			// ============================================================

			// ---- `as` wrapper — stripped by staticEval.skipTransparent →
			//      inner string literal is extracted → report. tsgo-only
			//      shape (ESTree's parser elides `as`-expressions at parse
			//      time so upstream never sees them). ----
			{
				Code:   `<div role={"checkbox" as any} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role={"checkbox" as const} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Paren-wrapped value — staticEval strips parens. ----
			{
				Code:   `<div role={("checkbox")} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role={(("checkbox"))} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Paren + `as` combination. ----
			{
				Code:   `<div role={("checkbox" as string)} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// JSX namespace / member-access tag names.
			// ============================================================

			// ---- `<Foo.Bar role="link" />` — elementType resolves to
			//      "Foo.Bar" via reactutil.GetJsxElementTypeString's
			//      PropertyAccessExpression arm → never in link.tagNames →
			//      report. ----
			{
				Code:   `<Foo.Bar role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("link", `<a href=...>, or <area href=...>`)},
			},
			// ---- Three-level member access. ----
			{
				Code:   `<Foo.Bar.Baz role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},

			// ============================================================
			// Element IS in some OTHER role's tagNames but NOT in this
			// role's — locks in that tagNames lookup is role-specific.
			// ============================================================

			// ---- `<header role="heading" />` — header is in banner's
			//      tagNames but NOT heading's → report. ----
			{
				Code:   `<header role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},
			// ---- `<a role="checkbox" />` — a is in link's tagNames but NOT
			//      checkbox's → report. ----
			{
				Code:   `<a role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Listener boundary: nested JSX inside conditional, fragment,
			// and array-children — listener fires on every JsxOpeningElement
			// / JsxSelfClosingElement encountered.
			// ============================================================

			// ---- JSX child of a JSX expression container. ----
			{
				Code: `<div>{cond ? <span role="checkbox" /> : null}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("checkbox", `<input type="checkbox">`, 1, 14),
				},
			},
			// ---- JSX child of a Fragment. ----
			{
				Code: `<><div role="banner" /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("banner", `<header>`, 1, 3),
				},
			},
			// ---- JSX inside an array literal. ----
			{
				Code: `<div>{[<span role="link" key={1} />, <span role="heading" key={2} />]}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("link", `<a href=...>, or <area href=...>`, 1, 8),
					expectedErrAt("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`, 1, 38),
				},
			},
			// ---- Triple-nested JSX with mixed role / non-role siblings —
			//      each report comes from the same listener entry, not
			//      from a residual visitor state. ----
			{
				Code: `<main>
  <section role="banner">
    <div role="rowgroup">
      <p>ok</p>
    </div>
  </section>
</main>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("banner", `<header>`, 2, 3),
					expectedErrAt("rowgroup", `<tbody>, <tfoot>, or <thead>`, 3, 5),
				},
			},

			// ============================================================
			// LogicalExpression where the LEFT operand statically resolves
			// to a falsy literal: short-circuits to the RIGHT operand. The
			// right-hand role string is what reaches roleElements lookup.
			// Lock in staticEval's `||` truthy-short-circuit arm.
			// ============================================================
			{
				Code:   `<div role={"" || "checkbox"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Nullish coalescing `??` — same path. ----
			{
				Code:   `<div role={null ?? "checkbox"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Dimension 4: line/column assertions on multi-line code.
			// ============================================================

			// ---- JSX on line 2 — locks in the report position. ----
			{
				Code: `
<div role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 2, 1)},
			},
			// ---- Indented JSX inside a return — locks column. ----
			{
				Code: `function App() {
  return <div role="link" />;
}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("link", `<a href=...>, or <area href=...>`, 2, 10)},
			},

			// ============================================================
			// Dimension 4: nested JSX — listener must fire on every match,
			//              not just outer-most.
			// ============================================================
			{
				Code: `<section>
  <div role="banner" />
</section>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("banner", `<header>`, 2, 3)},
			},
			{
				Code: `<div role="heading">
  <span role="checkbox" />
</div>`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`, 1, 1),
					expectedErrAt("checkbox", `<input type="checkbox">`, 2, 3),
				},
			},

			// ============================================================
			// Dimension 4: button-on-element-that's-not-button — the rule
			//              should not auto-pair a `<button>` to the `link`
			//              role (they live in separate roleElements entries).
			// ============================================================
			{
				Code:   `<button role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("link", `<a href=...>, or <area href=...>`)},
			},
			// ---- Same direction inverted: `<a role="button" />` — a is NOT
			//      a button tag; button.tagNames is just [input, button].
			//      Report. Locks the unidirectional mapping. ----
			{
				Code:   `<a role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("button", `<input type="button">, <input type="image">, <input type="reset">, <input type="submit">, or <button>`)},
			},

			// ============================================================
			// Dimension 4: column position for declaration-prefixed JSX
			//              variants — exercise the same shape with arrow
			//              and class wrappers, both single-line.
			// ============================================================
			{
				Code:   `const x = <div role="rowgroup" />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("rowgroup", `<tbody>, <tfoot>, or <thead>`, 1, 11)},
			},

			// ============================================================
			// HTML entity decoding on the direct `<X role="...">` form —
			// extractRoleString applies jsxtransforms.DecodeEntities so
			// `role="&#99;heckbox"` resolves to "checkbox" (same as ESTree's
			// JSX parser would expose). The `{…}`-wrapped form carries a
			// JS string literal (not JSX attribute text) and is NOT decoded,
			// matching jsx-ast-utils.
			// ============================================================

			// ---- &#99; → "c"; full decoded value "checkbox". ----
			{
				Code:   `<div role="&#99;heckbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- &#32; → ASCII space; decoding yields a multi-token role
			//      whose LAST token "checkbox" trips the rule. Locks in
			//      the entity-decode-before-tokenize ordering. ----
			{
				Code:   `<div role="button&#32;checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Hex entity. ----
			{
				Code:   `<div role="&#x68;eading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},

			// ============================================================
			// Identifier whose NAME is a real role string — upstream
			// `getPropValue(Identifier)` synthesizes the identifier name
			// as a string, so this REPORTS. Locks in PropStaticStringValue's
			// identifier-name-string behavior; without it, real codebase
			// patterns like
			//
			//	const checkbox = 'checkbox';
			//	<div role={checkbox} />
			//
			// would silently slip past the rule.
			// ============================================================
			{
				Code:   `<div role={checkbox} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			{
				Code:   `<div role={heading} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},
			{
				Code:   `<div role={banner} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("banner", `<header>`)},
			},
			// ---- Logical short-circuit to identifier: real-world
			//      "default role" pattern `role={x || checkbox}` when x is
			//      empty. staticEval(LogicalExpression) → right operand
			//      "checkbox" → report. ----
			{
				Code:   `<div role={"" || checkbox} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Real-world component patterns — the rule must fire identically
			// regardless of the enclosing declaration.
			// ============================================================

			// ---- Function-declaration component. ----
			{
				Code: `function Card() {
  return <div role="checkbox" />;
}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 2, 10)},
			},
			// ---- Arrow-function component (implicit return). ----
			{
				Code:   `const Card = () => <div role="checkbox" />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 1, 20)},
			},
			// ---- Class-component render method. ----
			{
				Code: `class Card extends React.Component {
  render() {
    return <div role="heading" />;
  }
}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`, 3, 12)},
			},
			// ---- React.forwardRef wrap. ----
			{
				Code: `const Card = React.forwardRef((props, ref) => (
  <div ref={ref} role="banner" />
));`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("banner", `<header>`, 2, 3)},
			},
			// ---- React.memo wrap. ----
			{
				Code:   `const Card = React.memo(() => <span role="checkbox" />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 1, 31)},
			},
			// ---- Custom HOC. ----
			{
				Code:   `const Card = withTheme(props => <span role="link" />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("link", `<a href=...>, or <area href=...>`, 1, 33)},
			},
			// ---- List rendering via .map — JSX inside the callback fires. ----
			{
				Code: `function List({ items }) {
  return items.map(item => <div key={item.id} role="checkbox" />);
}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 2, 28)},
			},
			// ---- Render-prop pattern. ----
			{
				Code:   `<Provider>{() => <div role="rowgroup" />}</Provider>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("rowgroup", `<tbody>, <tfoot>, or <thead>`, 1, 18)},
			},
			// ---- Custom hook returning JSX (uncommon but legal). ----
			{
				Code: `function useLabel() {
  return <span role="checkbox" />;
}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 2, 10)},
			},
			// ---- Composed providers + a deep child. ----
			{
				Code: `<ThemeProvider>
  <Layout>
    <div role="checkbox" />
  </Layout>
</ThemeProvider>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErrAt("checkbox", `<input type="checkbox">`, 3, 5)},
			},

			// ============================================================
			// Multiple offending elements in one file — each fires
			// independently with the correct line/column.
			// ============================================================
			{
				Code: `function App() {
  return (
    <div>
      <span role="checkbox" />
      <span role="heading" />
      <span role="banner" />
    </div>
  );
}`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					expectedErrAt("checkbox", `<input type="checkbox">`, 4, 7),
					expectedErrAt("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`, 5, 7),
					expectedErrAt("banner", `<header>`, 6, 7),
				},
			},

			// ============================================================
			// Duplicate-role attribute — JSX allows two attributes with
			// the same name; jsx-ast-utils' getProp returns the FIRST.
			// FindAttributeByName mirrors that, so the diagnostic
			// corresponds to the FIRST role value.
			// ============================================================
			{
				Code:   `<div role="checkbox" role="banner" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},

			// ============================================================
			// Spread-vs-explicit ordering. FindAttributeByName scans in
			// source order; the first JsxAttribute or literal-object spread
			// containing the prop wins. Non-literal spread (`{...props}`)
			// is opaque under upstream's strict default — does NOT shadow
			// later explicit attributes.
			// ============================================================

			// ---- Literal spread comes FIRST and contains the role —
			//      FindAttributeByName returns the spread's PropertyAssignment,
			//      not the later explicit attribute. ----
			{
				Code:   `<div {...{role:"checkbox"}} role="link" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Explicit role first, literal spread after — explicit wins. ----
			{
				Code:   `<div role="checkbox" {...{role:"link"}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Non-literal spread before explicit role — spread is
			//      opaque, explicit role wins. ----
			{
				Code:   `<div {...props} role="checkbox" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("checkbox", `<input type="checkbox">`)},
			},
			// ---- Non-literal spread AFTER explicit role — explicit wins. ----
			{
				Code:   `<span role="heading" {...rest} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},

			// ============================================================
			// Design-system "Box as=…" real-world pattern via
			// polymorphicPropName resolution to a non-semantic element.
			// ============================================================
			{
				Code:     `<Box as="span" role="banner" />`,
				Tsx:      true,
				Settings: polymorphicSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedErr("banner", `<header>`)},
			},

			// ============================================================
			// PascalCase tag vs lowercase native tag — Header is a custom
			// component, NOT the HTML <header>, so the heading role on it
			// still reports.
			// ============================================================
			{
				Code:   `<Header role="heading" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},

			// ============================================================
			// Multi-role attribute where ONLY the last token names an
			// invalid role — the rule reports the entire role-token even
			// if earlier tokens look "fine". Locks in upstream
			// getLastPropValue's "ignore all but the last" semantics.
			// ============================================================
			{
				Code:   `<div role="presentation banner" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("banner", `<header>`)},
			},

			// ============================================================
			// Element resolves through BOTH polymorphicPropName AND the
			// components map — the polymorphic prop wins (it's consulted
			// first in GetElementType), so the components map is only used
			// when the polymorphic prop is absent.
			// ============================================================
			{
				Code: `<Box as="span" role="heading" />`,
				Tsx:  true,
				Settings: map[string]interface{}{
					"jsx-a11y": map[string]interface{}{
						"polymorphicPropName": "as",
						"components": map[string]interface{}{
							"Box": "h1",
						},
					},
				},
				// "as" overrides "components", so elementType = "span",
				// NOT "h1". span ∉ heading.tagNames → report.
				Errors: []rule_tester.InvalidTestCaseError{expectedErr("heading", `<h1>, <h2>, <h3>, <h4>, <h5>, or <h6>`)},
			},
		},
	)
}
