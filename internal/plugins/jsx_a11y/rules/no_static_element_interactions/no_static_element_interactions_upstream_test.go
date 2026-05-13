package no_static_element_interactions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError is the single shape every invalid case in this rule
// produces. Shared with the extras suite via package scope.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noStaticElementInteractions",
	Message:   errorMessage,
}

// componentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Button: 'button', TestComponent: 'div' } } }
//
// `<Button>` resolves to `button` (interactive, exempt) and `<TestComponent>`
// resolves to `div` (static, no role → reported).
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Button":        "button",
			"TestComponent": "div",
		},
	},
}

// recommendedOptions mirrors upstream's
//
//	configs.recommended.rules['jsx-a11y/no-static-element-interactions'][1]
//
// — the option object passed by the recommended preset. The recommended
// `handlers` list is a narrower subset of the default (focus + keyboard +
// mouse), and `allowExpressionValues: true` opts into the non-literal-role
// skip.
//
// Source: eslint-plugin-jsx-a11y/src/index.js.
var recommendedOptions = []interface{}{
	map[string]interface{}{
		"handlers": []interface{}{
			"onClick",
			"onMouseDown",
			"onMouseUp",
			"onKeyPress",
			"onKeyDown",
			"onKeyUp",
		},
		"allowExpressionValues": true,
	},
}

// allowExpressionValuesTrueOptions mirrors upstream's
// `[{ allowExpressionValues: true }]` — a per-test option override on top of
// the default `handlers` list (no recommended-narrowed handler subset).
var allowExpressionValuesTrueOptions = []interface{}{
	map[string]interface{}{
		"allowExpressionValues": true,
	},
}

// allowExpressionValuesFalseOptions mirrors upstream's
// `[{ allowExpressionValues: false }]`.
var allowExpressionValuesFalseOptions = []interface{}{
	map[string]interface{}{
		"allowExpressionValues": false,
	},
}

// TestNoStaticElementInteractionsUpstream covers the full valid / invalid
// suite migrated 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/no-static-element-interactions-test.js`. Upstream
// runs two configurations of the same suite:
//
//   - `:recommended` — recommendedOptions (narrowed `handlers`,
//     `allowExpressionValues: true`)
//   - `:strict`      — no options (defaults: full focus+keyboard+mouse
//     handler list, allowExpressionValues falsy)
//
// Both configurations are run as one Go test, with each upstream case
// duplicated under both option shapes when the upstream suite did so. Order
// inside each group mirrors the upstream file so a future audit can grep
// across both side-by-side.
//
// Anything NOT in upstream's test file — TS wrappers, position assertions,
// listener-boundary repeats, options JSON-path matrix, etc. — lives in
// no_static_element_interactions_extras_test.go.
func TestNoStaticElementInteractionsUpstream(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule, []rule_tester.ValidTestCase{
		// ============================================================
		// alwaysValid (run under both :recommended and :strict)
		// ============================================================
		// Custom component — TestComponent has no setting here, treated
		// as a higher-level component (not in dom). Exempt by step 1.
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true},
		// `Button` is a JSX-component name, not in dom set → exempt step 1.
		{Code: `<Button onClick={doFoo} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<Button onClick={doFoo} />`, Tsx: true},
		// Same, with components setting → resolves to "button" (inherently interactive).
		{Code: `<Button onClick={doFoo} />`, Tsx: true, Options: recommendedOptions, Settings: componentsSettings},
		{Code: `<Button onClick={doFoo} />`, Tsx: true, Settings: componentsSettings},
		// No interactive handlers attached.
		{Code: `<div />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div />;`, Tsx: true},
		{Code: `<div className="foo" />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div className="foo" />;`, Tsx: true},
		// Spread is opaque (hasProp/getProp skip spreads), no direct
		// handler → no interactive props → exempt.
		{Code: `<div className="foo" {...props} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div className="foo" {...props} />;`, Tsx: true},
		// `aria-hidden` short-circuits via isHiddenFromScreenReader.
		{Code: `<div onClick={() => void 0} aria-hidden />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onClick={() => void 0} aria-hidden />;`, Tsx: true},
		{Code: `<div onClick={() => void 0} aria-hidden={true} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onClick={() => void 0} aria-hidden={true} />;`, Tsx: true},
		// onClick={null} → getPropValue → null → `!= null` is false → no
		// interactive prop → exempt.
		{Code: `<div onClick={null} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onClick={null} />;`, Tsx: true},

		// All flavors of input — inherently interactive (matches
		// interactiveElementRoleSchemas / interactiveElementAXSchemas).
		{Code: `<input onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="button" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="button" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="checkbox" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="checkbox" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="color" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="color" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="date" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="date" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="datetime" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="datetime" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="datetime-local" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="datetime-local" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="email" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="email" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="file" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="file" onClick={() => void 0} />`, Tsx: true},
		// `<input type="hidden">` — IsHiddenFromScreenReader returns true.
		{Code: `<input type="hidden" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="hidden" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="image" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="image" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="month" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="month" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="number" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="number" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="password" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="password" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="radio" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="radio" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="range" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="range" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="reset" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="reset" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="search" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="search" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="submit" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="submit" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="tel" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="tel" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="text" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="text" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="time" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="time" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="url" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="url" onClick={() => void 0} />`, Tsx: true},
		{Code: `<input type="week" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<input type="week" onClick={() => void 0} />`, Tsx: true},

		// Inherently interactive elements with handlers.
		{Code: `<button onClick={() => void 0} className="foo" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<button onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<datalist onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<datalist onClick={() => {}} />;`, Tsx: true},
		{Code: `<menuitem onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<menuitem onClick={() => {}} />;`, Tsx: true},
		{Code: `<option onClick={() => void 0} className="foo" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<option onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<select onClick={() => void 0} className="foo" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<select onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<textarea onClick={() => void 0} className="foo" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<textarea onClick={() => void 0} className="foo" />`, Tsx: true},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" />`, Tsx: true},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />`, Tsx: true},
		// `<audio>` etc. — non-interactive HTML elements (matched by
		// nonInteractiveElementAXSchemas or strictNonInteractive role
		// schemas).
		{Code: `<audio onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<audio onClick={() => {}} />;`, Tsx: true},
		{Code: `<form onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<form onClick={() => {}} />;`, Tsx: true},
		{Code: `<form onSubmit={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<form onSubmit={() => {}} />;`, Tsx: true},

		// Interactive role attribute on <div>.
		{Code: `<div role="button" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="button" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="checkbox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="checkbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="columnheader" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="columnheader" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="combobox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="combobox" onClick={() => {}} />;`, Tsx: true},
		// `form` is a non-interactive role (presence in the non-interactive
		// set → IsNonInteractiveRole returns true → exempt).
		{Code: `<div role="form" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="form" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="gridcell" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="gridcell" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="link" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="link" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitem" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="menuitem" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitemcheckbox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="menuitemcheckbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menuitemradio" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="menuitemradio" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="option" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="option" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="radio" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="radio" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="rowheader" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="rowheader" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="searchbox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="searchbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="slider" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="slider" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="spinbutton" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="spinbutton" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="switch" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="switch" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="tab" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="tab" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="textbox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="textbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="treeitem" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="treeitem" onClick={() => {}} />;`, Tsx: true},

		// Presentation roles — IsPresentationRole short-circuit.
		{Code: `<div role="presentation" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="presentation" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="presentation" onKeyDown={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="presentation" onKeyDown={() => {}} />;`, Tsx: true},

		// HTML elements with inherent non-interactive role.
		{Code: `<address onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<address onClick={() => {}} />;`, Tsx: true},
		{Code: `<article onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<article onClick={() => {}} />;`, Tsx: true},
		{Code: `<article onDblClick={() => void 0} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<article onDblClick={() => void 0} />;`, Tsx: true},
		{Code: `<aside onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<aside onClick={() => {}} />;`, Tsx: true},
		{Code: `<blockquote onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<blockquote onClick={() => {}} />;`, Tsx: true},
		{Code: `<br onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<br onClick={() => {}} />;`, Tsx: true},
		{Code: `<canvas onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<canvas onClick={() => {}} />;`, Tsx: true},
		{Code: `<caption onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<caption onClick={() => {}} />;`, Tsx: true},
		{Code: `<code onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<code onClick={() => {}} />;`, Tsx: true},
		{Code: `<dd onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<dd onClick={() => {}} />;`, Tsx: true},
		{Code: `<del onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<del onClick={() => {}} />;`, Tsx: true},
		{Code: `<details onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<details onClick={() => {}} />;`, Tsx: true},
		{Code: `<dfn onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<dfn onClick={() => {}} />;`, Tsx: true},
		{Code: `<dir onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<dir onClick={() => {}} />;`, Tsx: true},
		{Code: `<dl onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<dl onClick={() => {}} />;`, Tsx: true},
		{Code: `<dt onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<dt onClick={() => {}} />;`, Tsx: true},
		{Code: `<em onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<em onClick={() => {}} />;`, Tsx: true},
		{Code: `<embed onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<embed onClick={() => {}} />;`, Tsx: true},
		{Code: `<fieldset onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<fieldset onClick={() => {}} />;`, Tsx: true},
		{Code: `<figcaption onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<figcaption onClick={() => {}} />;`, Tsx: true},
		{Code: `<figure onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<figure onClick={() => {}} />;`, Tsx: true},
		{Code: `<footer onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<footer onClick={() => {}} />;`, Tsx: true},
		{Code: `<h1 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h1 onClick={() => {}} />;`, Tsx: true},
		{Code: `<h2 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h2 onClick={() => {}} />;`, Tsx: true},
		{Code: `<h3 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h3 onClick={() => {}} />;`, Tsx: true},
		{Code: `<h4 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h4 onClick={() => {}} />;`, Tsx: true},
		{Code: `<h5 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h5 onClick={() => {}} />;`, Tsx: true},
		{Code: `<h6 onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<h6 onClick={() => {}} />;`, Tsx: true},
		{Code: `<hr onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<hr onClick={() => {}} />;`, Tsx: true},
		{Code: `<html onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<html onClick={() => {}} />;`, Tsx: true},
		{Code: `<iframe onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<iframe onClick={() => {}} />;`, Tsx: true},
		{Code: `<img onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<img onClick={() => {}} />;`, Tsx: true},
		{Code: `<ins onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<ins onClick={() => {}} />;`, Tsx: true},
		{Code: `<label onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<label onClick={() => {}} />;`, Tsx: true},
		{Code: `<legend onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<legend onClick={() => {}} />;`, Tsx: true},
		{Code: `<li onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<li onClick={() => {}} />;`, Tsx: true},
		{Code: `<main onClick={() => void 0} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<main onClick={() => void 0} />;`, Tsx: true},
		{Code: `<mark onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<mark onClick={() => {}} />;`, Tsx: true},
		{Code: `<marquee onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<marquee onClick={() => {}} />;`, Tsx: true},
		{Code: `<menu onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<menu onClick={() => {}} />;`, Tsx: true},
		{Code: `<meter onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<meter onClick={() => {}} />;`, Tsx: true},
		{Code: `<nav onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<nav onClick={() => {}} />;`, Tsx: true},
		{Code: `<ol onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<ol onClick={() => {}} />;`, Tsx: true},
		{Code: `<optgroup onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<optgroup onClick={() => {}} />;`, Tsx: true},
		{Code: `<output onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<output onClick={() => {}} />;`, Tsx: true},
		{Code: `<p onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<p onClick={() => {}} />;`, Tsx: true},
		{Code: `<pre onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<pre onClick={() => {}} />;`, Tsx: true},
		{Code: `<progress onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<progress onClick={() => {}} />;`, Tsx: true},
		{Code: `<ruby onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<ruby onClick={() => {}} />;`, Tsx: true},
		// `<section>` is non-interactive ONLY when paired with
		// aria-label / aria-labelledby (per strictNonInteractive schema).
		{Code: `<section onClick={() => {}} aria-label="Aa" />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<section onClick={() => {}} aria-label="Aa" />;`, Tsx: true},
		{Code: `<section onClick={() => {}} aria-labelledby="js_1" />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<section onClick={() => {}} aria-labelledby="js_1" />;`, Tsx: true},
		{Code: `<strong onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<strong onClick={() => {}} />;`, Tsx: true},
		{Code: `<sub onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<sub onClick={() => {}} />;`, Tsx: true},
		{Code: `<summary onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<summary onClick={() => {}} />;`, Tsx: true},
		{Code: `<sup onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<sup onClick={() => {}} />;`, Tsx: true},
		{Code: `<table onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<table onClick={() => {}} />;`, Tsx: true},
		{Code: `<tbody onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<tbody onClick={() => {}} />;`, Tsx: true},
		{Code: `<tfoot onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<tfoot onClick={() => {}} />;`, Tsx: true},
		// `<th>` is inherently interactive (gridcell). `<thead>` non-interactive AX.
		{Code: `<th onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<th onClick={() => {}} />;`, Tsx: true},
		{Code: `<thead onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<thead onClick={() => {}} />;`, Tsx: true},
		{Code: `<time onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<time onClick={() => {}} />;`, Tsx: true},
		// `<tr>` is inherently interactive via interactiveElementRoleSchemas.
		{Code: `<tr onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<tr onClick={() => {}} />;`, Tsx: true},
		{Code: `<video onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<video onClick={() => {}} />;`, Tsx: true},
		{Code: `<ul onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<ul onClick={() => {}} />;`, Tsx: true},

		// Abstract roles — IsAbstractRole short-circuit.
		{Code: `<div role="command" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="command" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="composite" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="composite" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="input" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="input" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="landmark" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="landmark" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="range" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="range" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="roletype" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="roletype" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="sectionhead" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="sectionhead" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="select" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="select" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="structure" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="structure" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="widget" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="widget" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="window" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="window" onClick={() => {}} />;`, Tsx: true},

		// Non-interactive role attributes on <div>.
		{Code: `<div role="alert" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="alert" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="alertdialog" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="alertdialog" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="application" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="application" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="article" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="article" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="banner" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="banner" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="cell" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="cell" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="complementary" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="complementary" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="contentinfo" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="contentinfo" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="definition" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="definition" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="dialog" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="dialog" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="directory" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="directory" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="document" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="document" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="feed" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="feed" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="figure" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="figure" onClick={() => {}} />;`, Tsx: true},
		// `grid` is interactive (a widget descendant in interactiveRolesSet).
		{Code: `<div role="grid" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="grid" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="group" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="group" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="heading" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="heading" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="img" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="img" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="list" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="list" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="listbox" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="listbox" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="listitem" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="listitem" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="log" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="log" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="main" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="main" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="marquee" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="marquee" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="math" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="math" onClick={() => {}} />;`, Tsx: true},
		// `menu` / `menubar` are interactive (widget descendants).
		{Code: `<div role="menu" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="menu" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="menubar" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="menubar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="navigation" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="navigation" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="note" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="note" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="progressbar" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="progressbar" onClick={() => {}} />;`, Tsx: true},
		// `radiogroup` is interactive (widget descendant).
		{Code: `<div role="radiogroup" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="radiogroup" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="region" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="region" onClick={() => {}} />;`, Tsx: true},
		// `row` is interactive (widget descendant).
		{Code: `<div role="row" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="row" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="rowgroup" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="rowgroup" onClick={() => {}} />;`, Tsx: true},
		// `section` is in upstream's abstractRoles (yes — it overlaps).
		// Resolved via IsAbstractRole first, since abstract roles are
		// checked in the OR with the four interactive/non-interactive
		// classifications — observable result: exempt.
		{Code: `<div role="section" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="section" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="search" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="search" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="separator" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="separator" onClick={() => {}} />;`, Tsx: true},
		// `scrollbar` is interactive (widget descendant).
		{Code: `<div role="scrollbar" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="scrollbar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="status" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="status" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="table" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="table" onClick={() => {}} />;`, Tsx: true},
		// `tablist` is interactive (widget descendant).
		{Code: `<div role="tablist" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="tablist" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="tabpanel" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="tabpanel" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="term" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="term" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="timer" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="timer" onClick={() => {}} />;`, Tsx: true},
		// `toolbar` is interactive (widget descendant, plus axobject).
		{Code: `<div role="toolbar" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="toolbar" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="tooltip" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="tooltip" onClick={() => {}} />;`, Tsx: true},
		// `tree` / `treegrid` are interactive (widget descendants).
		{Code: `<div role="tree" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="tree" onClick={() => {}} />;`, Tsx: true},
		{Code: `<div role="treegrid" onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="treegrid" onClick={() => {}} />;`, Tsx: true},

		// All the possible handlers — none of them is in either the
		// recommended `handlers` list NOR the default focus+keyboard+mouse
		// list (these are clipboard / composition / change / form / touch /
		// scroll / wheel / media / animation handlers). hasInteractiveProps
		// → false → exempt.
		{Code: `<div onCopy={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCopy={() => {}} />;`, Tsx: true},
		{Code: `<div onCut={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCut={() => {}} />;`, Tsx: true},
		{Code: `<div onPaste={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onPaste={() => {}} />;`, Tsx: true},
		{Code: `<div onCompositionEnd={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCompositionEnd={() => {}} />;`, Tsx: true},
		{Code: `<div onCompositionStart={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCompositionStart={() => {}} />;`, Tsx: true},
		{Code: `<div onCompositionUpdate={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCompositionUpdate={() => {}} />;`, Tsx: true},
		{Code: `<div onChange={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onChange={() => {}} />;`, Tsx: true},
		{Code: `<div onInput={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onInput={() => {}} />;`, Tsx: true},
		{Code: `<div onSubmit={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onSubmit={() => {}} />;`, Tsx: true},
		{Code: `<div onSelect={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onSelect={() => {}} />;`, Tsx: true},
		{Code: `<div onTouchCancel={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTouchCancel={() => {}} />;`, Tsx: true},
		{Code: `<div onTouchEnd={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTouchEnd={() => {}} />;`, Tsx: true},
		{Code: `<div onTouchMove={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTouchMove={() => {}} />;`, Tsx: true},
		{Code: `<div onTouchStart={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTouchStart={() => {}} />;`, Tsx: true},
		{Code: `<div onScroll={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onScroll={() => {}} />;`, Tsx: true},
		{Code: `<div onWheel={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onWheel={() => {}} />;`, Tsx: true},
		{Code: `<div onAbort={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onAbort={() => {}} />;`, Tsx: true},
		{Code: `<div onCanPlay={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCanPlay={() => {}} />;`, Tsx: true},
		{Code: `<div onCanPlayThrough={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onCanPlayThrough={() => {}} />;`, Tsx: true},
		{Code: `<div onDurationChange={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDurationChange={() => {}} />;`, Tsx: true},
		{Code: `<div onEmptied={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onEmptied={() => {}} />;`, Tsx: true},
		{Code: `<div onEncrypted={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onEncrypted={() => {}} />;`, Tsx: true},
		{Code: `<div onEnded={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onEnded={() => {}} />;`, Tsx: true},
		{Code: `<div onError={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onError={() => {}} />;`, Tsx: true},
		{Code: `<div onLoadedData={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onLoadedData={() => {}} />;`, Tsx: true},
		{Code: `<div onLoadedMetadata={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onLoadedMetadata={() => {}} />;`, Tsx: true},
		{Code: `<div onLoadStart={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onLoadStart={() => {}} />;`, Tsx: true},
		{Code: `<div onPause={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onPause={() => {}} />;`, Tsx: true},
		{Code: `<div onPlay={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onPlay={() => {}} />;`, Tsx: true},
		{Code: `<div onPlaying={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onPlaying={() => {}} />;`, Tsx: true},
		{Code: `<div onProgress={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onProgress={() => {}} />;`, Tsx: true},
		{Code: `<div onRateChange={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onRateChange={() => {}} />;`, Tsx: true},
		{Code: `<div onSeeked={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onSeeked={() => {}} />;`, Tsx: true},
		{Code: `<div onSeeking={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onSeeking={() => {}} />;`, Tsx: true},
		{Code: `<div onStalled={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onStalled={() => {}} />;`, Tsx: true},
		{Code: `<div onSuspend={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onSuspend={() => {}} />;`, Tsx: true},
		{Code: `<div onTimeUpdate={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTimeUpdate={() => {}} />;`, Tsx: true},
		{Code: `<div onVolumeChange={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onVolumeChange={() => {}} />;`, Tsx: true},
		{Code: `<div onWaiting={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onWaiting={() => {}} />;`, Tsx: true},
		{Code: `<div onLoad={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onLoad={() => {}} />;`, Tsx: true},
		{Code: `<div onAnimationStart={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onAnimationStart={() => {}} />;`, Tsx: true},
		{Code: `<div onAnimationEnd={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onAnimationEnd={() => {}} />;`, Tsx: true},
		{Code: `<div onAnimationIteration={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onAnimationIteration={() => {}} />;`, Tsx: true},
		{Code: `<div onTransitionEnd={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onTransitionEnd={() => {}} />;`, Tsx: true},

		// ============================================================
		// recommended-only valid (under recommendedOptions: narrowed
		// handlers + allowExpressionValues=true)
		// ============================================================
		// Handlers that exist in the DEFAULT focus+keyboard+mouse list but
		// NOT in recommendedOptions.handlers — these are reported under
		// :strict (see invalid section) but exempt here.
		{Code: `<div onFocus={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onBlur={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onContextMenu={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDblClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDoubleClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDrag={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragEnd={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragEnter={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragExit={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragLeave={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragOver={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDragStart={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onDrop={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onMouseEnter={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onMouseLeave={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onMouseMove={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onMouseOut={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div onMouseOver={() => {}} />;`, Tsx: true, Options: recommendedOptions},

		// Expression-typed role under :recommended — allowExpressionValues=true
		// → IsNonLiteralProperty skip.
		{Code: `<div role={ROLE_BUTTON} onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions},
		{Code: `<div  {...this.props} role={this.props.role} onKeyPress={e => this.handleKeyPress(e)}>{this.props.children}</div>`, Tsx: true, Options: recommendedOptions},
		// allowExpressionValues=true override + default handlers.
		{Code: `<div role={BUTTON} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Ternary with literal arms under allowExpressionValues=true:
		// IsNonLiteralProperty returns true → skip (the "two-literal-arms"
		// branch upstream returns unconditionally, so observable behavior is
		// the same).
		{Code: `<div role={isButton ? "button" : "link"} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Upstream's test file lists these next two cases under VALID with a
		// stray `errors` field that the upstream RuleTester ignores on valid
		// cases. Mirror as valid.
		{Code: `<div role={isButton ? "button" : LINK} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		{Code: `<div role={isButton ? BUTTON : LINK} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// neverValid (run under both :recommended and :strict)
		// ============================================================
		// `<div onClick={...} />` — static, no role, hasInteractiveProps → REPORT.
		{Code: `<div onClick={() => void 0} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `role={undefined}` — IsNonLiteralProperty returns false ("undefined"
		// identifier matches the upstream JSXExpressionContainer escape arm),
		// so allowExpressionValues doesn't kick in even under :recommended.
		// IsInteractiveRole on `undefined` → false. Fall through → REPORT.
		{Code: `<div onClick={() => void 0} role={undefined} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onClick={() => void 0} role={undefined} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Spread is opaque to hasProp / getProp under default
		// spreadStrict=true; the direct `onClick` is what matches.
		{Code: `<div onClick={() => void 0} {...props} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onClick={() => void 0} {...props} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `aria-hidden={false}` — IsHiddenFromScreenReader returns false
		// (only `=== true` exempts). onKeyUp is in both handler lists.
		{Code: `<div onKeyUp={() => void 0} aria-hidden={false} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyUp={() => void 0} aria-hidden={false} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Static elements with no inherent role.
		{Code: `<a onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a tabIndex="0" onClick={() => void 0} />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<a tabIndex="0" onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<acronym onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<acronym onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<applet onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<applet onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<area onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<area onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<b onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<b onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<base onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<base onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<bdi onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<bdi onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<bdo onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<bdo onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<big onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<big onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<blink onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<blink onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<body onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<body onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<center onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<center onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<cite onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<cite onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<col onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<col onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<colgroup onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<colgroup onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<content onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<content onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<data onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<data onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<font onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<font onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<frame onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<frame onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<frameset onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<frameset onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<head onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<head onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<header>` always returns false from IsNonInteractiveElement
		// (upstream's banner-landmark-context note). Falls through to REPORT.
		{Code: `<header onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<header onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<hgroup onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<hgroup onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<i onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<i onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<kbd onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<kbd onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<keygen onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<keygen onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<map onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<map onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<meta onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<meta onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<noembed onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<noembed onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<noscript onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<noscript onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<object onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<object onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<param onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<param onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<picture onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<picture onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<q onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<q onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rp onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rp onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rt onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rt onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rtc onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<rtc onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<s onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<s onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<samp onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<samp onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<script onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<script onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Bare `<section>` (no aria-label / aria-labelledby) — does not match
		// the non-interactive section schema → REPORT.
		{Code: `<section onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<section onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<small onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<small onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<source onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<source onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<spacer onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<spacer onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<span onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<span onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<strike onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<strike onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<style onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<style onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<title onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<title onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<track onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<track onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tt onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<tt onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<u onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<u onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<var onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<var onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<wbr onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<wbr onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<xmp onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<xmp onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Keyboard / mouse handler variants present in both handler lists.
		{Code: `<div onKeyDown={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyDown={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyPress={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyPress={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyUp={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onKeyUp={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onClick={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseDown={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseDown={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseUp={() => {}} />;`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseUp={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Custom component remapped to <div> via settings.
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true, Options: recommendedOptions, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<TestComponent onClick={doFoo} />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// strict-only invalid (no options → defaults: full focus +
		// keyboard + mouse handler list, allowExpressionValues falsy)
		// ============================================================
		// Handlers from the default focus+keyboard+mouse list that are NOT
		// in recommendedOptions.handlers → reported only under :strict.
		{Code: `<div onContextMenu={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDblClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDoubleClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDrag={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragEnd={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragEnter={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragExit={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragLeave={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragOver={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDragStart={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onDrop={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseEnter={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseLeave={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseMove={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseOut={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div onMouseOver={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// Without allowExpressionValues=true, non-literal `role` is no
		// longer a skip. IsInteractiveRole on a non-literal returns false →
		// REPORT.
		{Code: `<div role={ROLE_BUTTON} onClick={() => {}} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div  {...this.props} role={this.props.role} onKeyPress={e => this.handleKeyPress(e)}>{this.props.children}</div>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Explicit allowExpressionValues=false — same outcome as defaults.
		{Code: `<div role={BUTTON} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesFalseOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Ternary with literal arms but allowExpressionValues=false → IsInteractiveRole
		// on a ConditionalExpression returns false → REPORT.
		{Code: `<div role={isButton ? "button" : "link"} onClick={() => {}} />;`, Tsx: true, Options: allowExpressionValuesFalseOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
