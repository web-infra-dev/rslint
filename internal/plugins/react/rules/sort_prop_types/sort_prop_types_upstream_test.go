// TestSortPropTypesUpstream migrates representative valid/invalid cases from
// upstream tests/lib/rules/sort-prop-types.js. rslint-specific lock-ins live
// in sort_prop_types_extras_test.go.
package sort_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortPropTypesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortPropTypesRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `Component.propTypes = { a: PropTypes.string, b: PropTypes.number };`, Tsx: true},
		{Code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`, Options: map[string]any{"noSortAlphabetically": true}, Tsx: true},
		{Code: `Component.propTypes = { a: PropTypes.string, onChange: PropTypes.func };`, Options: map[string]any{"callbacksLast": true}, Tsx: true},
		{Code: `Component.propTypes = { required: PropTypes.string.isRequired, optional: PropTypes.string };`, Options: map[string]any{"requiredFirst": true}, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{Code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Message: propsNotSortedText, Line: 1, Column: 46}}},
		{Code: `Component.propTypes = { onChange: PropTypes.func, name: PropTypes.string };`, Options: map[string]any{"callbacksLast": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "callbackPropsLast", Message: callbackPropsLastText, Line: 1, Column: 25}}},
		{Code: `Component.propTypes = { optional: PropTypes.string, required: PropTypes.string.isRequired };`, Options: map[string]any{"requiredFirst": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst", Message: requiredPropsFirstText, Line: 1, Column: 53}}},
	})
}
