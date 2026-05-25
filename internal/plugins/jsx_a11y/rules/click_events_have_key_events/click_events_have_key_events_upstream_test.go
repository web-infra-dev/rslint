package click_events_have_key_events

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's `expectedError` constant — the single
// shape every invalid case in the rule produces. Centralized so a future
// error-text tweak touches one place. Shared with the extras test via
// package scope.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "clickEventsHaveKeyEvents",
	Message:   errorMessage,
	Line:      1,
	Column:    1,
}

// footerComponentSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Footer: 'footer' } } }
//
// so `<Footer>` resolves to `footer` and trips the non-interactive check.
var footerComponentSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Footer": "footer",
		},
	},
}

// TestClickEventsHaveKeyEventsUpstream covers the full valid/invalid suite
// migrated 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/click-events-have-key-events-test.js`. Order inside
// each group mirrors the upstream file so a future audit can grep across
// both side-by-side.
//
// Anything NOT in upstream's test file — TS wrappers, position assertions,
// extra spread shapes — lives in click_events_have_key_events_extras_test.go.
func TestClickEventsHaveKeyEventsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ClickEventsHaveKeyEventsRule,
		[]rule_tester.ValidTestCase{
			// ---- onClick + each keyboard listener variant ----
			{Code: `<div onClick={() => void 0} onKeyDown={foo}/>;`, Tsx: true},
			{Code: `<div onClick={() => void 0} onKeyUp={foo} />;`, Tsx: true},
			{Code: `<div onClick={() => void 0} onKeyPress={foo}/>;`, Tsx: true},
			{Code: `<div onClick={() => void 0} onKeyDown={foo} onKeyUp={bar} />;`, Tsx: true},
			// ---- onClick + onKeyDown then spread — direct keyboard prop wins. ----
			{Code: `<div onClick={() => void 0} onKeyDown={foo} {...props} />;`, Tsx: true},
			// ---- No onClick prop → rule short-circuits. ----
			{Code: `<div className="foo" />;`, Tsx: true},
			// ---- aria-hidden variants — element is hidden from screen readers. ----
			{Code: `<div onClick={() => void 0} aria-hidden />;`, Tsx: true},
			{Code: `<div onClick={() => void 0} aria-hidden={true} />;`, Tsx: true},
			// ---- aria-hidden={false} BUT has keyboard listener → still valid. ----
			{Code: `<div onClick={() => void 0} aria-hidden={false} onKeyDown={foo} />;`, Tsx: true},
			// ---- aria-hidden={undefined} BUT has keyboard listener → still valid. ----
			{Code: `<div onClick={() => void 0} onKeyDown={foo} aria-hidden={undefined} />;`, Tsx: true},
			// ---- Interactive elements via aria-query elementRoles / elementAXObjects. ----
			{Code: `<input type="text" onClick={() => void 0} />`, Tsx: true},
			{Code: `<input onClick={() => void 0} />`, Tsx: true},
			{Code: `<button onClick={() => void 0} className="foo" />`, Tsx: true},
			{Code: `<option onClick={() => void 0} className="foo" />`, Tsx: true},
			{Code: `<select onClick={() => void 0} className="foo" />`, Tsx: true},
			{Code: `<textarea onClick={() => void 0} className="foo" />`, Tsx: true},
			{Code: `<a onClick={() => void 0} href="http://x.y.z" />`, Tsx: true},
			{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />`, Tsx: true},
			// ---- input[type=hidden] → isHiddenFromScreenReader returns true. ----
			{Code: `<input onClick={() => void 0} type="hidden" />;`, Tsx: true},
			// ---- role=presentation / role=none → isPresentationRole. ----
			{Code: `<div onClick={() => void 0} role="presentation" />;`, Tsx: true},
			{Code: `<div onClick={() => void 0} role="none" />;`, Tsx: true},
			// ---- Custom components — element type isn't in aria-query's `dom`
			//      map → upstream short-circuits before checking interactivity. ----
			{Code: `<TestComponent onClick={doFoo} />`, Tsx: true},
			{Code: `<Button onClick={doFoo} />`, Tsx: true},
			{Code: `<Footer onClick={doFoo} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Plain <div> with onClick, no keyboard listener. ----
			{Code: `<div onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- role={undefined} — not "presentation"/"none" → div remains non-interactive. ----
			{Code: `<div onClick={() => void 0} role={undefined} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- {...props} spread is opaque under hasAnyProp's spreadStrict
			//      default → cannot supply keyboard listener. ----
			{Code: `<div onClick={() => void 0} {...props} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- Non-interactive sectioning DOM elements. ----
			{Code: `<section onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<main onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<article onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<header onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<footer onClick={() => void 0} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- aria-hidden={false} → not hidden → still requires keyboard. ----
			{Code: `<div onClick={() => void 0} aria-hidden={false} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- <a> without href → not interactive (interactive schema gates
			//      on `href` attribute existence). ----
			{Code: `<a onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			{Code: `<a tabIndex="0" onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
			// ---- componentsMap resolves <Footer> → 'footer' (non-interactive DOM). ----
			{Code: `<Footer onClick={doFoo} />`, Tsx: true, Settings: footerComponentSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		},
	)
}
