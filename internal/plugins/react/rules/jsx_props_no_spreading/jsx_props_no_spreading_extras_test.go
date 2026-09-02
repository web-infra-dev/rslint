// TestJsxPropsNoSpreadingExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// N/A: receiver wrappers, element-access keys, class/function containers, and
// ancestor scope walks do not apply because this rule only inspects JSX spread
// attributes and their immediate tag / object-literal argument.
package jsx_props_no_spreading

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxPropsNoSpreadingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoSpreadingRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized and TypeScript-wrapped arguments ----
		{Code: `<App {...({ value: 1 })} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},

		// ---- Dimension 4: object member forms all count as Property ----
		{Code: `<App {...{ shorthand }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		{Code: `<App {...{ "quoted": 1, 0: 2, [key]: 3 }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		{Code: `<App {...{ method() {}, get value() { return 1; }, set value(next) {} }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		{Code: `<App {...{}} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		// ---- Dimension 4: SpreadAssignment does not mask sibling properties ----

		// Locks in upstream create() tag-name fallback: JSXNamespacedName has
		// no tagName, so only explicitSpread can suppress this diagnostic.
		{Code: `<svg:path {...{ value: 1 }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},

		// Locks in upstream JSXMemberExpression handling for nested components.
		{Code: `<components.Group.Inner {...props} />`, Tsx: true, Options: map[string]any{"custom": "ignore"}},
		// Upstream's JSXMemberExpression helper does not preserve the full
		// name for member chains deeper than one property.

		// ---- Real-user: explicit inline object props ----
		{Code: `<Foo {...{ prop1, prop2, prop3 }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}},
		// ---- Real-user: sub-component exceptions ----
		{Code: `<components.DropdownIndicator {...props} />`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.DropdownIndicator"}}},
		{Code: `<components.ClearIndicator {...props} />`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.ClearIndicator"}}},

		// Locks in upstream html ignore branch with no exception.
		{Code: `<div {...props} />`, Tsx: true, Options: map[string]any{"html": "ignore"}},
		// Locks in upstream html enforce branch with an exception.
		{Code: `<div {...props} />`, Tsx: true, Options: map[string]any{"html": "enforce", "exceptions": []any{"div"}}},
		// Locks in upstream custom ignore branch with no exception.
		{Code: `<App {...props} />`, Tsx: true, Options: map[string]any{"custom": "ignore"}},
		// Locks in upstream custom enforce branch with an exception.
		{Code: `<App {...props} />`, Tsx: true, Options: map[string]any{"custom": "enforce", "exceptions": []any{"App"}}},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: multiline diagnostic range ----
		{
			Code: "<App\n  {...props}\n  other\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noSpreading", Message: noSpreadingMessage,
				Line: 2, Column: 3, EndLine: 2, EndColumn: 13,
			}},
		},
		// Locks in upstream explicitSpread fallback: a non-object expression
		// remains reportable when explicitSpread is ignored.
		{Code: `<App {...props} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 16)}},
		// Locks in upstream explicitSpread all-property predicate: any spread
		// member makes every(isProperty) false.
		{Code: `<App {...props!} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 17)}},
		{Code: `type Props = {}; <App {...(props as Props)} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 23, 44)}},
		{Code: `type Props = {}; <App {...(props satisfies Props)} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 23, 51)}},
		{Code: `<App {...props?.bag} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 21)}},
		{Code: `<App {...{ value: 1, ...rest }} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 32)}},
		// ---- Real-user: conditional explicit props remain enforced ----
		{Code: `<div {...(actAsBanner && { role: "banner" })} />`, Tsx: true, Options: map[string]any{"explicitSpread": "ignore"}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 46)}},
		// A namespaced tag has no tagName and therefore reaches the report.
		{Code: `<svg:path {...props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 11, 21)}},
		// ---- Real-user: unlisted sub-component remains enforced ----
		{Code: `<components.DropdownIndicator {...props} />`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.Group"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 31, 41)}},
		{Code: `<components.Group.Inner {...props} />`, Tsx: true, Options: map[string]any{"exceptions": []any{"components.Group.Inner"}}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 25, 35)}},
		// Locks in defaults: an explicitly empty option object is equivalent to
		// omitting options entirely.
		{Code: `<App {...props} />`, Tsx: true, Options: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{noSpreadingError(1, 6, 16)}},
	})
}
