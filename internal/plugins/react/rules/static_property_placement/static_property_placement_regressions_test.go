package static_property_placement

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStaticPropertyPlacementRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StaticPropertyPlacementRule,
		[]rule_tester.ValidTestCase{
			{Code: `abstract class C extends React.Component { abstract propTypes: Props; abstract context: Context; }`, Tsx: true},
			{Code: `let Box = { C: {} }; Box.C.propTypes = {}; Box.C = class extends React.Component {};`, Tsx: true},
			{Code: `let Box = {}; Box.C = {}; Box.C = class extends React.Component {}; Box.C.propTypes = {};`, Tsx: true},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C["displayName"] = {};`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C['displayName'] = {};`, Tsx: true},
			{Code: `class C extends React.Component { static ["propTypes"] = {}; }`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: "class C extends React.Component { static [`displayName`] = {}; }", Options: []interface{}{staticGetter}, Tsx: true},
			{Code: `class C extends React.Component { static #displayName = {}; }`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: `class C extends React.Component { static get #displayName() { return {}; } }`, Options: []interface{}{propertyAssignment}, Tsx: true},
			{Code: `class C extends React.Component {} C["propTypes"] = {};`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: "class C extends React.Component {} C[`propTypes`] = {};", Tsx: true},
			{Code: "class C extends React.Component {} C[`displayName`] = {};", Tsx: true},
			{Code: `class C extends React.Component {} C.#displayName = {};`, Options: []interface{}{staticGetter}, Tsx: true},
			{Code: `class C extends React.Component {} function f() { class C {} C.propTypes = {}; }`, Tsx: true},
			{Code: `class C extends React.Component {} function f() { let C; C.propTypes = {}; }`, Tsx: true},
			{Code: `class C extends React.Component {} { class C {} C.propTypes = {}; }`, Tsx: true},
			{Code: `const C = createReactClass({ render() { class Helper { static get propTypes() { return {}; } } return null; } });`, Options: []interface{}{propertyAssignment}, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `let Box = {}; Box.C = class extends React.Component {}; Box.C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `/** @extends React.Component */ class C { propTypes = {}; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `/** @augments React.Component */ class C {} C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `class C extends React.Component { static #propTypes = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component { #propTypes = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C.foo.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box[C].propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C[foo].propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C["foo"].propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box["C"].propTypes = {};`, Tsx: true},
			{Code: `class C extends React.Component {} C.#helper.propTypes + value;`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C.#helper.propTypes + value;`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component { propTypes = {}; }`, Options: []interface{}{propertyAssignment}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass"}}},
			{Code: `class C extends React.Component { static ["displayName"] = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc"}}},
			{Code: `class C extends React.Component {} C["displayName"] = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp"}}},
			{Code: `class C extends React.Component {} C[(propTypes)] = {};`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component {} C[("displayName")] = {};`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'displayName' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component { static getDefaultProps = {}; }`, Options: []interface{}{propertyAssignment}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass", Message: "'defaultProps' should be declared outside the class body."}}},
			{Code: `class C extends React.Component {} C.getDefaultProps = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'defaultProps' should be declared as a static class property."}}},
			{Code: `class C extends React.Component { static [propTypes]?: Props; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc"}}},
			{Code: `class C extends React.Component { static #props: Props = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component { static #context: Context = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'contextTypes' should be declared as a static getter class function."}}},
			{Code: `class C extends React.Component { #props: Props = {}; }`, Options: []interface{}{propertyAssignment}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass", Message: "'propTypes' should be declared outside the class body."}}},
			{Code: `const Box = { C: class extends React.Component {} }; Box.C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp"}}},
			{Code: `const C = createReactClass({ render() { class Helper extends React.Component { static get propTypes() { return {}; } } return null; } });`, Options: []interface{}{propertyAssignment}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass", Message: "'propTypes' should be declared outside the class body."}}},
			{Code: `const C = class extends React.Component { method() { C.propTypes = {}; } };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp"}}},
			{Code: `class C extends React.Component { static props: Props = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc"}}},
			{Code: `class C extends React.Component { static context: Context = {}; }`, Options: []interface{}{staticGetter}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc"}}},
		},
	)
}
