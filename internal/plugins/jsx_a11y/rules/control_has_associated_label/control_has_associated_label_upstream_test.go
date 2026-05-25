package control_has_associated_label

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's
//
//	const expectedError = {
//	  message: 'A control must be associated with a text label.',
//	  type: 'JSXOpeningElement',
//	};
//
// Shared with the no-config and extras tests via package scope so a future
// message tweak touches one place.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "controlHasAssociatedLabel",
	Message:   errorMessage,
}

// recommendedOptions mirrors `configs.recommended.rules[ruleName][1]` /
// `configs.strict.rules[ruleName][1]` from eslint-plugin-jsx-a11y's
// `src/index.js`. The two configs ship identical option dictionaries for
// this rule — both presets disable the rule by default and supply these
// option defaults to apply when the user enables it.
//
// Notes:
//   - `depth` is NOT in either preset; tests that need a deeper traversal
//     supply it themselves under `Options` and it is preserved by upstream's
//     `ruleOptionsMapperFactory` (concat + Object.fromEntries union — the
//     test-supplied keys are listed first, but only the preset's keys
//     overwrite on conflict and neither preset includes `depth`).
//   - `includeRoles: ['alert', 'dialog']` (also shipped in the preset) is
//     NOT in the rule's option schema and is silently ignored by upstream
//     (`generateObjSchema` does not set `additionalProperties: false`).
//     We omit it for parity with the schema, since passing an extra key
//     would be a no-op anyway.
var recommendedOptions = map[string]interface{}{
	"ignoreElements": []interface{}{
		"audio", "canvas", "embed", "input", "textarea", "tr", "video",
	},
	"ignoreRoles": []interface{}{
		"grid", "listbox", "menu", "menubar", "radiogroup",
		"row", "tablist", "toolbar", "tree", "treegrid",
	},
}

// withRecommended returns a fresh options map that merges the per-test
// overrides onto recommendedOptions. Mirrors upstream's
// `ruleOptionsMapperFactory(recommendedOptions)` semantics:
//   - test-supplied keys are listed FIRST in upstream's `concat`, so on
//     conflict the recommended-preset value wins. This rule's preset
//     ships only `ignoreElements` and `ignoreRoles`, so conflicts in
//     practice only matter when a test tries to override one of those.
//   - For non-conflicting keys (e.g. `depth`, `controlComponents`,
//     `labelAttributes`), the test value is preserved verbatim.
func withRecommended(overrides map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(recommendedOptions)+len(overrides))
	for k, v := range overrides {
		out[k] = v
	}
	for k, v := range recommendedOptions {
		out[k] = v
	}
	return out
}

// customControlSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { CustomControl: 'button' } } }
//
// — used by `<CustomControl>Save</CustomControl>` and
// `<CustomControl></CustomControl>` cases.
var customControlSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"CustomControl": "button",
		},
	},
}

// TestControlHasAssociatedLabelUpstreamRecommended mirrors upstream's
// `:recommended` AND `:strict` test runs combined. The two configs ship
// identical options for this rule, so a single Go test exercises both
// alwaysValid (expect no diagnostic) and neverValid (expect a diagnostic)
// arrays with recommendedOptions merged in.
//
// Anything outside upstream's test file lives in
// `control_has_associated_label_extras_test.go`. Upstream's `:no-config`
// run lives in `TestControlHasAssociatedLabelUpstreamNoConfig` below.
func TestControlHasAssociatedLabelUpstreamRecommended(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ControlHasAssociatedLabelRule,
		[]rule_tester.ValidTestCase{
			// ---- Custom Control Components ----
			{Code: `<CustomControl><span><span>Save</span></span></CustomControl>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{
					"depth":             float64(3),
					"controlComponents": []interface{}{"CustomControl"},
				})},
			{Code: `<CustomControl><span><span label="Save"></span></span></CustomControl>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{
					"depth":             float64(3),
					"controlComponents": []interface{}{"CustomControl"},
					"labelAttributes":   []interface{}{"label"},
				})},
			{Code: `<CustomControl>Save</CustomControl>`, Tsx: true,
				Options:  withRecommended(nil),
				Settings: customControlSettings},
			// ---- Interactive Elements ----
			{Code: `<button>Save</button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button><span>Save</span></button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button><span><span>Save</span></span></button>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{"depth": float64(3)})},
			{Code: `<button><span><span><span><span><span><span><span><span>Save</span></span></span></span></span></span></span></span></button>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{"depth": float64(9)})},
			{Code: `<button><img alt="Save" /></button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button><span aria-label="Save" /></button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button><span aria-labelledby="js_1" /></button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button>{sureWhyNot}</button>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<button><span><span label="Save"></span></span></button>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{
					"depth":           float64(3),
					"labelAttributes": []interface{}{"label"},
				})},
			{Code: `<a href="#">Save</a>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<area href="#">Save</area>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<link>Save</link>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<menuitem>Save</menuitem>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<option>Save</option>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<th>Save</th>`, Tsx: true, Options: withRecommended(nil)},
			// ---- Interactive Roles ----
			{Code: `<div role="button">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="checkbox">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="columnheader">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="combobox">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="gridcell">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="link">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitem">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemcheckbox">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemradio">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="option">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="progressbar">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="radio">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="rowheader">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="searchbox">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="slider">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="spinbutton">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="switch">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tab">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="textbox">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="treeitem">Save</div>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="button" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="checkbox" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="columnheader" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="combobox" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="gridcell" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="link" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitem" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemcheckbox" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemradio" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="option" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="progressbar" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="radio" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="rowheader" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="searchbox" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="slider" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="spinbutton" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="switch" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tab" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="textbox" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="treeitem" aria-label="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="button" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="checkbox" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="columnheader" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="combobox" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="gridcell" aria-labelledby="Save" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="link" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitem" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemcheckbox" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menuitemradio" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="option" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="progressbar" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="radio" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="rowheader" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="searchbox" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="slider" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="spinbutton" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="switch" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tab" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="textbox" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="treeitem" aria-labelledby="js_1" />`, Tsx: true, Options: withRecommended(nil)},
			// ---- Non-interactive Elements ----
			{Code: `<abbr />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<article />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<blockquote />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<br />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<caption />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dd />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<details />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dfn />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dialog />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dir />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dl />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<dt />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<fieldset />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<figcaption />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<figure />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<footer />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<form />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<frame />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h1 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h2 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h3 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h4 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h5 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<h6 />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<hr />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<iframe />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<img />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<label />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<legend />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<li />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<link />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<main />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<mark />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<marquee />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<menu />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<meter />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<nav />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<ol />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<p />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<pre />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<progress />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<ruby />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<section />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<table />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<tbody />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<tfoot />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<thead />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<time />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<ul />`, Tsx: true, Options: withRecommended(nil)},
			// ---- Non-interactive Roles ----
			{Code: `<div role="alert" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="alertdialog" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="application" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="article" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="banner" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="cell" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="complementary" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="contentinfo" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="definition" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="dialog" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="directory" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="document" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="feed" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="figure" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="form" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="group" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="heading" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="img" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="list" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="listitem" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="log" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="main" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="marquee" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="math" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="navigation" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="none" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="note" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="presentation" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="region" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="rowgroup" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="search" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="separator" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="status" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="table" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tabpanel" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="term" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="timer" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tooltip" />`, Tsx: true, Options: withRecommended(nil)},
			// ---- Via config: inputs / marginal interactive elements (skipped
			//      under recommended's ignoreElements) ----
			{Code: `<input />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="button" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="checkbox" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="color" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="date" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="datetime" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="email" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="file" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="hidden" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="hidden" name="bot-field"/>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="hidden" name="form-name" value="Contact Form"/>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="image" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="month" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="number" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="password" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="radio" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="range" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="reset" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="search" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="submit" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="tel" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="text" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<label>Foo <input type="text" /></label>`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input name={field.name} id="foo" type="text" value={field.value} disabled={isDisabled} onChange={changeText(field.onChange, field.name)} onBlur={field.onBlur} />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="time" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="url" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<input type="week" />`, Tsx: true, Options: withRecommended(nil)},
			// Marginal interactive elements
			{Code: `<audio />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<canvas />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<embed />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<textarea />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<tr />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<video />`, Tsx: true, Options: withRecommended(nil)},
			// Interactive roles to ignore
			{Code: `<div role="grid" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="listbox" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menu" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="menubar" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="radiogroup" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="row" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tablist" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="toolbar" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="tree" />`, Tsx: true, Options: withRecommended(nil)},
			{Code: `<div role="treegrid" />`, Tsx: true, Options: withRecommended(nil)},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `<button />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button><span /></button>`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button><img /></button>`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button><span title="This is not a real label" /></button>`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<button><span><span><span>Save</span></span></span></button>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{"depth": float64(3)}),
				Errors:  []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<CustomControl><span><span></span></span></CustomControl>`, Tsx: true,
				Options: withRecommended(map[string]interface{}{
					"depth":             float64(3),
					"controlComponents": []interface{}{"CustomControl"},
				}),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<CustomControl></CustomControl>`, Tsx: true,
				Options:  withRecommended(nil),
				Settings: customControlSettings,
				Errors:   []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<a href="#" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<area href="#" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<menuitem />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<option />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<th />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<td />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// Interactive Roles
			{Code: `<div role="button" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="checkbox" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="columnheader" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="combobox" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="link" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="gridcell" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="menuitem" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="menuitemcheckbox" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="menuitemradio" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="option" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="progressbar" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="radio" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="rowheader" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="scrollbar" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="searchbox" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="slider" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="spinbutton" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="switch" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="tab" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<div role="textbox" />`, Tsx: true, Options: withRecommended(nil),
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		},
	)
}

// TestControlHasAssociatedLabelUpstreamNoConfig mirrors upstream's
// `:no-config` run:
//
//	ruleTester.run(`${ruleName}:no-config`, rule, {
//	  valid: [
//	    { code: '<input type="hidden" />' },
//	    { code: '<input type="text" aria-hidden="true" />' },
//	  ],
//	  invalid: [
//	    { code: '<input type="text" />', errors: [expectedError] },
//	  ],
//	});
//
// Important contrast with the :recommended/:strict suite: without options,
// the `input` element is NOT in `newIgnoreElements` (only the hard-coded
// `link` is). The hidden-from-screen-reader short-circuit therefore decides
// whether the rule reports. `<input type="hidden">` is caught by the input
// branch of IsHiddenFromScreenReader; `<input type="text" aria-hidden="true">`
// is caught by the `aria-hidden === true` branch. A plain
// `<input type="text" />` reaches the interactive-element check (input is
// in the interactive AX schema) and reports because it has no label.
func TestControlHasAssociatedLabelUpstreamNoConfig(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ControlHasAssociatedLabelRule,
		[]rule_tester.ValidTestCase{
			{Code: `<input type="hidden" />`, Tsx: true},
			{Code: `<input type="text" aria-hidden="true" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `<input type="text" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		},
	)
}
