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
		{Code: `class Component { static propTypes = { a: PropTypes.string, z: PropTypes.string }; }`, Tsx: true},
		{Code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.string };`, Options: map[string]any{"noSortAlphabetically": true}, Tsx: true},
		{Code: `Component.propTypes = { z: PropTypes.string, ...shared, a: PropTypes.string };`, Tsx: true},
		{Code: `const shape = { a: PropTypes.string, b: PropTypes.string }; Component.propTypes = { item: PropTypes.shape(shape) };`, Options: map[string]any{"sortShapeProp": true}, Tsx: true},
		{Code: `type Props = { z: string; a: string }; function Component(props: Props) {}`, Options: map[string]any{"checkTypes": false}, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{Code: `Component.propTypes = { z: PropTypes.string, a: PropTypes.number };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Message: propsNotSortedText, Line: 1, Column: 46}}},
		{Code: `Component.propTypes = { onChange: PropTypes.func, name: PropTypes.string };`, Options: map[string]any{"callbacksLast": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "callbackPropsLast", Message: callbackPropsLastText, Line: 1, Column: 25}}},
		{Code: `Component.propTypes = { optional: PropTypes.string, required: PropTypes.string.isRequired };`, Options: map[string]any{"requiredFirst": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst", Message: requiredPropsFirstText, Line: 1, Column: 53}}},
		{Code: `class Component { static propTypes = { z: PropTypes.string, y: PropTypes.string, a: PropTypes.string }; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}, {MessageId: "propsNotSorted"}}},
		{Code: `Component.propTypes = { Z: PropTypes.string, a: PropTypes.string };`, Options: map[string]any{"ignoreCase": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		{Code: `class Component { static propTypes = forbidExtraProps({ z: PropTypes.string, y: PropTypes.string, a: PropTypes.string }); }`, Settings: map[string]interface{}{"propWrapperFunctions": []any{"forbidExtraProps"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}, {MessageId: "propsNotSorted"}}},
		{Code: `Component.propTypes = { x: PropTypes.string, z: PropTypes.shape({ a: PropTypes.string, c: PropTypes.number.isRequired, b: PropTypes.string, ...other, f: PropTypes.string, d: PropTypes.string }) };`, Options: map[string]any{"sortShapeProp": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}, {MessageId: "propsNotSorted"}}},
		{Code: `type Props = { onChange: () => void; name: string }; const Component = (props: Props) => null;`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
	})
}
