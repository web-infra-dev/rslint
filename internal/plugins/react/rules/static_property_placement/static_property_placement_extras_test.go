// TestStaticPropertyPlacementExtras covers tsgo edge shapes, real-user
// patterns, and branch lock-ins not present in the upstream fixture suite.
// Upstream-mirror cases live in static_property_placement_upstream_test.go.
package static_property_placement

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStaticPropertyPlacementExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StaticPropertyPlacementRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream MemberExpression boundary behavior ----
			{Code: `class MyComponent extends React.Component {} MyComponent?.propTypes = {};`, Tsx: true},
			{Code: `class MyComponent extends React.Component {} MyComponent?.[displayName] = {};`, Tsx: true},
			{Code: `class MyComponent extends React.Component {} MyComponent.propTypes, value;`, Tsx: true},
			// ---- Dimension 4: receiver wrappers on a component assignment ----
			{Code: `class MyComponent extends React.Component {} (MyComponent).propTypes = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component {} ((MyComponent)).propTypes = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component {} MyComponent!.propTypes = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component {} (MyComponent as any).propTypes = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component {} MyComponent['propTypes'] = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component {} MyComponent[
  'propTypes'
] = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			// ---- Dimension 4: access/key forms that are not watched ----
			{Code: `class MyComponent extends React.Component {} MyComponent[dynamic] = {};`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class MyComponent extends React.Component { static ['other'] = {}; }`, Tsx: true},
			{Code: `class MyComponent extends React.Component { static #propTypes = {}; }`, Tsx: true},
			{Code: `class MyComponent extends React.Component { static [dynamic] = {}; }`, Tsx: true},
			{Code: `class MyComponent extends React.Component { static accessor propTypes = {}; }`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: `class MyComponent extends React.Component { accessor displayName = {}; }`, Options: []interface{}{propertyAssignment}, Tsx: true},
			// ---- Dimension 4: declaration/container forms ----
			// A class expression is detected as a component, so assignment mode
			// correctly rejects its static field (the invalid counterpart below
			// locks the report contract).
			{Code: `class MyComponent extends Other.Component { static propTypes = {}; }`, Tsx: true},
			{Code: `class MyComponent { static propTypes = {}; }`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `abstract class MyComponent extends React.Component { static propTypes: unknown; }`, Tsx: true},
			{Code: `declare class MyComponent extends React.Component { static propTypes: unknown; }`, Tsx: true},
			// ---- Dimension 4: nested traversal boundary ----
			{Code: `class Outer extends React.Component { method() { class Helper { static propTypes = {}; } } }`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class Outer extends React.Component { method() { class Inner extends React.Component { static defaultProps = {}; } } }`, Tsx: true},
			// ---- Dimension 4: graceful degradation ----
			{Code: `class MyComponent extends React.Component { static propTypes = {...props}; static defaultProps = {}; }`, Tsx: true},
			{Code: `class MyComponent extends React.Component {}`, Options: []interface{}{propertyAssignment}, Tsx: true},
			// ---- Branch lock-in: assignments inside a class scope are skipped ----
			{Code: `class MyComponent extends React.Component { method() { MyComponent.propTypes = {}; } }`, Options: []interface{}{propertyAssignment}, Tsx: true},
			// ---- Real-user: class expression used for a component binding ----
			{Code: `const Button = class extends React.PureComponent { static defaultProps = {}; };`, Options: []interface{}{staticPublicField}, Tsx: true},
			// ---- Real-user: per-property migration from getter to public fields ----
			{Code: `class Form extends React.Component { static get propTypes() { return {}; } static defaultProps = {}; }`, Options: []interface{}{staticPublicField, map[string]interface{}{"propTypes": staticGetter}}, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			// The upstream MemberExpression listener reports non-comma binary
			// parents, not only assignment expressions.
			{Code: `class MyComponent extends React.Component {} MyComponent.propTypes + value;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp"}}},
			// ---- Dimension 4: class field mismatch reports the whole member ----
			{Code: "class MyComponent extends React.Component {\n  static propTypes = {};\n}", Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function.", Line: 2, Column: 3}}},
			// ---- Dimension 4: getter mismatch reports the whole accessor ----
			{Code: "class MyComponent extends React.Component {\n  static get displayName() { return 'x'; }\n}", Options: []interface{}{staticPublicField}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'displayName' should be declared as a static class property.", Line: 2, Column: 3}}},
			// ---- Dimension 4: assignment reports the member expression ----
			{Code: "class MyComponent extends React.Component {}\nMyComponent.propTypes = {};", Options: []interface{}{staticPublicField}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 2, Column: 1}}},
			// ---- Branch lock-in: a non-component class never participates ----
			{Code: `class Widget { static propTypes = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: nil},
			// ---- Branch lock-in: all property names use the configured override ----
			{Code: `class MyComponent extends React.Component { static propTypes = {}; static defaultProps = {}; }`, Options: []interface{}{propertyAssignment, map[string]interface{}{"propTypes": staticPublicField}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass"}}},
			// ---- Branch lock-in: compound assignment is still an assignment ----
			{Code: `class MyComponent extends React.Component {} MyComponent.propTypes += value;`, Options: []interface{}{staticPublicField}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp"}}},
		},
	)
}
