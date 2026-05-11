package no_noninteractive_tabindex

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError is the single shape every invalid case in the rule
// produces. Centralized so a future error-text tweak touches one place.
// Shared between this file and no_noninteractive_tabindex_extras_test.go
// via package scope.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noNoninteractiveTabindex",
	Message:   errorMessage,
}

// componentsSettings mirrors upstream's
//
//	settings: { 'jsx-a11y': { components: { Article: 'article', MyButton: 'button' } } }
//
// `<MyButton>` resolves to `button` (interactive, exempt) and `<Article>`
// resolves to `article` (non-interactive, reported).
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Article":  "article",
			"MyButton": "button",
		},
	},
}

// recommendedOptions mirrors upstream's
//
//	configs.recommended.rules['jsx-a11y/no-noninteractive-tabindex'][1]
//
// — the option object passed by the recommended preset. `tags: []` is an
// explicit empty list (still falsy under the upstream `tags && includes(...)`
// guard, no different from the absent default), `roles: ['tabpanel']` adds
// the tab-panel exemption, and `allowExpressionValues: true` opts into the
// non-literal-role skip.
var recommendedOptions = []interface{}{
	map[string]interface{}{
		"tags":                  []interface{}{},
		"roles":                 []interface{}{"tabpanel"},
		"allowExpressionValues": true,
	},
}

// allowExpressionValuesTrueOptions mirrors upstream's `[{ allowExpressionValues: true }]`
// — a per-test option override on top of the default config (no tags, no roles).
var allowExpressionValuesTrueOptions = []interface{}{
	map[string]interface{}{
		"allowExpressionValues": true,
	},
}

// allowExpressionValuesFalseOptions mirrors upstream's `[{ allowExpressionValues: false }]`.
var allowExpressionValuesFalseOptions = []interface{}{
	map[string]interface{}{
		"allowExpressionValues": false,
	},
}

// TestNoNoninteractiveTabindexUpstream covers the full valid/invalid suite
// migrated 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/no-noninteractive-tabindex-test.js`. Upstream runs
// two configurations of the same suite:
//
//   - `:strict`       — no options (defaults: tags & roles absent,
//                       allowExpressionValues falsy)
//   - `:recommended`  — the recommended preset's options
//
// Both configurations are run as one Go test, with each upstream case
// duplicated under both option shapes when the upstream suite did so. Order
// inside each group mirrors the upstream file so a future audit can grep
// across both side-by-side.
//
// Anything NOT in upstream's test file — TS wrappers, position assertions,
// listener-boundary repeats, full ARIA-role survey, real-world a11y
// patterns, options JSON-path matrix, etc. — lives in
// no_noninteractive_tabindex_extras_test.go.
func TestNoNoninteractiveTabindexUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveTabindexRule, []rule_tester.ValidTestCase{
		// ============================================================
		// alwaysValid (run under both :strict and :recommended)
		// ============================================================
		// Custom component (not in dom set) — listener short-circuits via
		// dom.has(type) === false.
		{Code: `<MyButton tabIndex={0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<MyButton tabIndex={0} />`, Tsx: true},
		// No tabIndex prop at all — early return on `getProp(attributes, 'tabIndex')`.
		{Code: `<button />`, Tsx: true, Options: recommendedOptions},
		{Code: `<button />`, Tsx: true},
		// `<button>` is inherently interactive (interactiveElementRoleSchemas
		// matches `{name: "button"}` with no attributes) — exempt.
		{Code: `<button tabIndex="0" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<button tabIndex="0" />`, Tsx: true},
		{Code: `<button tabIndex={0} />`, Tsx: true, Options: recommendedOptions},
		{Code: `<button tabIndex={0} />`, Tsx: true},
		// `<div />` has no tabIndex — early return.
		{Code: `<div />`, Tsx: true, Options: recommendedOptions},
		{Code: `<div />`, Tsx: true},
		// tabIndex < 0 — passes the `tabIndex >= 0` gate's negation; not reported.
		{Code: `<div tabIndex="-1" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<div tabIndex="-1" />`, Tsx: true},
		// `role="button"` is interactive — IsInteractiveRole returns true → exempt.
		{Code: `<div role="button" tabIndex="0" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="button" tabIndex="0" />`, Tsx: true},
		// tabIndex < 0 short-circuits regardless of role.
		{Code: `<div role="article" tabIndex="-1" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<div role="article" tabIndex="-1" />`, Tsx: true},
		// `<article>` is non-interactive (article is in nonInteractiveRoles),
		// but tabIndex=-1 is allowed.
		{Code: `<article tabIndex="-1" />`, Tsx: true, Options: recommendedOptions},
		{Code: `<article tabIndex="-1" />`, Tsx: true},
		// componentsSettings: `<Article>` resolves to `article`, tabIndex=-1 → exempt.
		{Code: `<Article tabIndex="-1" />`, Tsx: true, Options: recommendedOptions, Settings: componentsSettings},
		{Code: `<Article tabIndex="-1" />`, Tsx: true, Settings: componentsSettings},
		// componentsSettings: `<MyButton>` resolves to `button`, interactive → exempt.
		{Code: `<MyButton tabIndex={0} />`, Tsx: true, Options: recommendedOptions, Settings: componentsSettings},
		{Code: `<MyButton tabIndex={0} />`, Tsx: true, Settings: componentsSettings},

		// ============================================================
		// recommended-only valid
		// ============================================================
		// `tabpanel` is in the recommendedOptions roles list — exempt under recommended.
		{Code: `<div role="tabpanel" tabIndex="0" />`, Tsx: true, Options: recommendedOptions},
		// recommended.allowExpressionValues=true: `role={ROLE_BUTTON}` is
		// non-literal → skipped without inspection.
		{Code: `<div role={ROLE_BUTTON} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: recommendedOptions},
		// Equivalent to the previous case but options are spelled out rather
		// than reading from `recommended`.
		{Code: `<div role={BUTTON} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// allowExpressionValues=true + role is a ConditionalExpression with
		// two literal arms. Upstream has a "special case" branch for this
		// shape, but the surrounding if-block returns regardless — the
		// observable behavior is "any non-literal role under
		// allowExpressionValues=true is exempt", which we mirror.
		{Code: `<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Same shape, one literal arm and one Identifier arm — also exempt
		// under our (and upstream's observable) behavior. Upstream's test
		// file lists this case under valid with a stray `errors` field; the
		// errors field is ignored on valid cases by the upstream RuleTester,
		// so the case is effectively "expected valid". We mirror that.
		{Code: `<div role={isButton ? "button" : LINK} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
		// Both arms are Identifiers — also exempt under our (and upstream's
		// observable) behavior, for the same reason.
		{Code: `<div role={isButton ? BUTTON : LINK} onClick={() => {}} tabIndex="0"/>;`, Tsx: true, Options: allowExpressionValuesTrueOptions},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// neverValid (run under both :strict and :recommended)
		// ============================================================
		// `<div>` is non-interactive (in nonInteractiveElementRoleSchemas),
		// no role to override, tabIndex=0 ≥ 0 → reports.
		{Code: `<div tabIndex="0" />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div tabIndex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `role="article"` is non-interactive → IsInteractiveRole returns
		// false; element is also non-interactive; tabIndex=0 → reports.
		{Code: `<div role="article" tabIndex="0" />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<div role="article" tabIndex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// `<article>` itself is non-interactive — in non-interactive role schemas.
		{Code: `<article tabIndex="0" />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<article tabIndex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Same as above with numeric tabIndex.
		{Code: `<article tabIndex={0} />`, Tsx: true, Options: recommendedOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<article tabIndex={0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// componentsSettings remaps `Article` to `article`; the resolved tag
		// is non-interactive and tabIndex=0 → reports.
		{Code: `<Article tabIndex={0} />`, Tsx: true, Options: recommendedOptions, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Article tabIndex={0} />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ============================================================
		// strict-only invalid (no options → defaults: tags / roles absent,
		// allowExpressionValues falsy)
		// ============================================================
		// Without the recommended `roles: ['tabpanel']` exemption,
		// `<div role="tabpanel">` falls through: tabpanel is in
		// nonInteractiveRoles → IsInteractiveRole returns false → reports.
		{Code: `<div role="tabpanel" tabIndex="0" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Without allowExpressionValues=true, the non-literal-role escape
		// hatch is closed → IsNonLiteralProperty's gate is bypassed →
		// IsInteractiveRole(role={ROLE_BUTTON}) returns false (role is not a
		// literal string) → reports.
		{Code: `<div role={ROLE_BUTTON} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Explicit allowExpressionValues=false — same outcome as defaults.
		{Code: `<div role={BUTTON} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: allowExpressionValuesFalseOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// Conditional with literal arms but allowExpressionValues=false →
		// the rule still reaches IsInteractiveRole, which can't resolve a
		// literal value from a ConditionalExpression → returns false →
		// reports.
		{Code: `<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />;`, Tsx: true, Options: allowExpressionValuesFalseOptions, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
	})
}
