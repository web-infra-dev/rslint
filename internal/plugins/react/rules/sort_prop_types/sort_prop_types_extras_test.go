// TestSortPropTypesExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Upstream-migrated cases live in the sibling
// sort_prop_types_upstream_test.go file.
package sort_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortPropTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortPropTypesRule, []rule_tester.ValidTestCase{
		// Upstream keeps TypeScript assertions opaque in this path.
		{Code: `const x = { propTypes: ({ z: PropTypes.string, a: PropTypes.string } as any) };`, Tsx: true},
		// Identifier resolution does not apply to object-literal propTypes properties.
		{Code: `const props = { z: PropTypes.string, a: PropTypes.string }; const x = { propTypes: props };`, Tsx: true},
		// The shape listener likewise leaves TypeScript assertions opaque.
		{Code: `any.shape(({ z: PropTypes.string, a: PropTypes.string } as any));`, Options: map[string]any{"sortShapeProp": true}, Tsx: true},
		// A defaulted parameter is an ESTree AssignmentPattern without a direct annotation.
		{Code: `type Props = { z: string; a: string }; function Component(p: Props = {} as Props) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		// ---- Dimension 4: spread resets the ordering group ----
		{Code: `Component.propTypes = { z: PropTypes.string, ...shared, a: PropTypes.string };`, Tsx: true},
		// ---- Dimension 4: TS expression wrappers on the propTypes object ----
		{Code: `Component.propTypes = ({ a: PropTypes.string, z: PropTypes.string } as any);`, Tsx: true},
		// N/A: private keys and element access do not participate in upstream's static key sort.
	}, []rule_tester.InvalidTestCase{
		// Upstream falls back to the authored text of computed keys when sorting.
		{Code: `Component.propTypes = { [z]: PropTypes.string, a: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 48}}},
		// Locks in upstream checkSorted() required-first arm: a required prop after an optional prop.
		{Code: `Component.propTypes = { a: PropTypes.string, b: PropTypes.string.isRequired };`, Options: map[string]any{"requiredFirst": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst", Line: 1, Column: 46}}},
		// Locks in upstream CallExpression listener: shape properties are checked independently.
		{Code: `Component.propTypes = { item: PropTypes.shape({ z: PropTypes.string, a: PropTypes.string }) };`, Options: map[string]any{"sortShapeProp": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 70}}},
		// ---- Real-user: #2702 TypeScript alias used by a function component ----
		{Code: `type Props = { z: string; a: string }; function Component(props: Props) { return null; }`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 27}}},
		// ---- Real-user: #3578 inline TypeScript component props ----
		{Code: `const Component = (props: { z: string; a: string }) => null;`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 40}}},
	})
}
