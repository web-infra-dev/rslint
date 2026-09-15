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

// These cases were checked against eslint-plugin-react 7.37.5 using ESLint
// 9.39.5 and @typescript-eslint/parser 8.69.0, including diagnostic ranges.
func TestStaticPropertyPlacementComponentLookup(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StaticPropertyPlacementRule,
		[]rule_tester.ValidTestCase{
			// An earlier write overrides the initializer.
			{Code: `const Box = { C: class extends React.Component {} }; Box.C = {}; Box.C.propTypes = {};`, Tsx: true},
			// Only an exact source reference matches.
			{Code: `let Box = {}; Box . C = class extends React.Component {}; Box.C.propTypes = {};`, Tsx: true},
			// Parentheses inside an earlier reference remain significant.
			{Code: `let Box = {}; (Box).C = class extends React.Component {}; Box.C.propTypes = {};`, Tsx: true},
			// Preserve TypeScript assertion wrappers.
			{Code: `const Box = { C: class extends React.Component {} } as const; Box.C.propTypes = {};`, Tsx: true},
			// Preserve satisfies wrappers.
			{Code: `const Box = { C: class extends React.Component {} } satisfies Record<string,unknown>; Box.C.propTypes = {};`, Tsx: true},
			// The first matching initializer is authoritative.
			{Code: `var C = {}; var C = class extends React.Component {}; C.propTypes = {};`, Tsx: true},
			// Do not search an unrelated second child scope.
			{Code: `function f() {} function g() { class C extends React.Component {} } C.propTypes = {};`, Tsx: true},
			// Parameters shadow outer components.
			{Code: `class C extends React.Component {} function f(C) { C.propTypes = {}; }`, Tsx: true},
			// Braced names are not explicit component markers.
			{Code: `/** @extends {React.Component} */ class C { propTypes = {}; }`, Tsx: true},
			// An outer class marker does not mark a nested class.
			{Code: `/** @extends React.Component */ class Outer { C = class { propTypes = {}; }; }`, Tsx: true},
			// The file pragma changes the recognized superclass.
			{Code: `/* @jsx R.h */ class C extends React.Component { propTypes = {}; } C.defaultProps = {};`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			// A computed reference must not hide the declaration.
			{Code: `const Box = { C: class extends React.Component {} }; Box[C] = {}; Box.C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 67, EndLine: 1, EndColumn: 82}}},
			// Skip ordinary receiver parentheses.
			{Code: `const Box = { C: class extends React.Component {} }; (Box).C.propTypes = {};`, Tsx: true, Options: []any{staticGetter}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function.", Line: 1, Column: 54, EndLine: 1, EndColumn: 71}}},
			// Keep the full nested path for a literal property.
			{Code: `const Box = { nested: { C: class extends React.Component {} } }; Box.nested.C["displayName"] = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'displayName' should be declared as a static class property.", Line: 1, Column: 66, EndLine: 1, EndColumn: 93}}},
			// An initializer on a repeated declaration is a reference.
			{Code: `var C; var C = class extends React.Component {}; C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 50, EndLine: 1, EndColumn: 61}}},
			// Search the first child scope.
			{Code: `function f() { class C extends React.Component {} } C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 53, EndLine: 1, EndColumn: 64}}},
			// Search the first grandchild scope.
			{Code: `function f() { function g() { class C extends React.Component {} } } C.propTypes = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 70, EndLine: 1, EndColumn: 81}}},
			// Class expression comments belong to their declaration.
			{Code: `/** @extends React.Component */ const C = class { propTypes = {}; };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 51, EndLine: 1, EndColumn: 66}}},
			// Same-line markers on object methods are supported.
			{Code: `const Box = { /** @extends React.Component */ C() {} }; Box.C.propTypes = {};`, Tsx: true, Options: []any{staticGetter}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function.", Line: 1, Column: 57, EndLine: 1, EndColumn: 72}}},
			// Same-line markers on assignments are supported.
			{Code: `const Box = {}; /** @extends React.Component */ Box.C = class { propTypes = {}; }; Box.C.defaultProps = {};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notStaticClassProp", Message: "'propTypes' should be declared as a static class property.", Line: 1, Column: 65, EndLine: 1, EndColumn: 80}, {MessageId: "notStaticClassProp", Message: "'defaultProps' should be declared as a static class property.", Line: 1, Column: 84, EndLine: 1, EndColumn: 102}}},
			// Follow the file pragma.
			{Code: `/* @jsx R */ class C extends R.Component { propTypes = {}; } C.defaultProps = {};`, Tsx: true, Options: []any{staticGetter}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGetterClassFunc", Message: "'propTypes' should be declared as a static getter class function.", Line: 1, Column: 44, EndLine: 1, EndColumn: 59}, {MessageId: "notGetterClassFunc", Message: "'defaultProps' should be declared as a static getter class function.", Line: 1, Column: 62, EndLine: 1, EndColumn: 76}}},
			// Decorated fields retain the diagnostic range.
			{Code: `class C extends React.Component { @decorator propTypes = {}; }`, Tsx: true, Options: []any{propertyAssignment}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "declareOutsideClass", Message: "'propTypes' should be declared outside the class body.", Line: 1, Column: 35, EndLine: 1, EndColumn: 61}}},
		},
	)
}
