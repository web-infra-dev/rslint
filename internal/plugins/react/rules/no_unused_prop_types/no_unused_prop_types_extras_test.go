// TestNoUnusedPropTypesExtras locks in tsgo edge shapes and branches that the
// upstream suite does not fully exercise. The upstream-mirror cases live in
// no_unused_prop_types_upstream_test.go.
package no_unused_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPropTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedPropTypesRule,
		[]rule_tester.ValidTestCase{
			// Typed helpers are not components merely because their parameter has a type.
			{Code: `function helper(props: { used: string; unused: string }) { return props.used.length; } helper.propTypes = { used: PropTypes.string, unused: PropTypes.string };`, Tsx: true},
			// Same-name type aliases resolve from the nearest lexical scope.
			{Code: `{ type Props = { first: string }; const Foo = (props: Props) => <div>{props.first}</div>; } { type Props = { second: string }; const Foo = (props: Props) => <div>{props.second}</div>; }`, Tsx: true},
			// Configured custom validators remain opaque, including shape-like method names.
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: CustomValidator.shape({ inner: CustomValidator.string }) };`, Options: map[string]interface{}{"customValidators": []interface{}{"CustomValidator"}, "skipShapeProps": false}, Tsx: true},
			// ---- Dimension 4: parenthesized and TypeScript expression wrappers ----
			{Code: `class Hello extends React.Component { render() { return <div>{(this.props).name}</div>; } } Hello.propTypes = ({ name: PropTypes.string });`, Tsx: true},
			{Code: `const Hello = ((props: Props) => <div>{props.name}</div>) as any; type Props = { name: string }; Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			// ---- Dimension 4: computed/static property keys ----
			{Code: `class Hello extends React.Component { render() { return <div>{this.props["name"]}</div>; } } Hello.propTypes = { ["name"]: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this.props[0]}</div>; } } Hello.propTypes = { 0: PropTypes.string };`, Tsx: true},
			// ---- Dimension 4: optional-chain links ----
			{Code: `class Hello extends React.Component { static propTypes = { name: PropTypes.string }; render() { return <div>{this.props?.name}</div>; } }`, Tsx: true},
			// ---- Real-user: nested helper reads ----
			{Code: `function Hello(props) { function renderLabel() { return <span>{props.label}</span>; } return <div>{renderLabel()}</div>; } Hello.propTypes = { label: PropTypes.string };`, Tsx: true},
			// ---- Real-user: rest props make the accessed set open ----
			{Code: `function Hello({ name, ...rest }) { return <div>{rest.other}</div>; } Hello.propTypes = { name: PropTypes.string, other: PropTypes.string };`, Tsx: true},
			// Locks in upstream reportUnusedPropType()'s `props === true` arm.
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string, ...external };`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			// Shadowed components must receive their own propTypes assignment.
			{Code: `{ function Foo(props) { return <div>{props.second}</div>; } Foo.propTypes = { first: PropTypes.string, second: PropTypes.string }; } { function Foo(props) { return <div>{props.first}</div>; } Foo.propTypes = { first: PropTypes.string, second: PropTypes.string }; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'first' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'second' PropType is defined but prop is never used"}}},
			// Only callbacks with a props-derived receiver may destructure prop paths.
			{Code: `function Foo(props) { [1].map(({ used }) => used); return <div />; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// User-configured component wrappers participate in component discovery.
			{Code: `const Foo = observer((props) => <div />); Foo.propTypes = { unused: PropTypes.string };`, Settings: map[string]interface{}{"componentWrapperFunctions": []interface{}{"observer"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// UNSAFE lifecycle aliases are unavailable before React 16.3.
			{Code: `class Foo extends React.Component { static propTypes = { value: PropTypes.string }; UNSAFE_componentWillReceiveProps(nextProps) { return nextProps.value; } render() { return <div />; } }`, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.2.0"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'value' PropType is defined but prop is never used"}}},
			// TypeScript method signatures are declared props too.
			{Code: `interface Props { used: string; onClick(): void } function Foo(props: Props) { return <div>{props.used}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'onClick' PropType is defined but prop is never used"}}},
			// Synthetic array children report at their validator argument, not the outer key.
			{Code: "class Comp1 extends React.Component {\n  render() { return <span>{this.props.prop1}</span>; }\n}\nComp1.propTypes = { prop1: PropTypes.number };\nclass Comp2 extends React.Component {\n  static propTypes = {\n    prop2: PropTypes.arrayOf(\n    Comp1.propTypes.prop1\n    )\n  };\n  render() { return <span>{this.props.prop2}</span>; }\n}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'prop2.*' PropType is defined but prop is never used", Line: 8, Column: 5, EndLine: 8, EndColumn: 26}}},
			// Every synthesized nested declaration retains the complete assignment range.
			{Code: "class Foo extends React.Component {\n  render() { return <div />; }\n}\nFoo.propTypes = {};\nFoo.propTypes.a.b.c = PropTypes.number;", Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'a' PropType is defined but prop is never used", Line: 5, Column: 1, EndLine: 5, EndColumn: 39}, {MessageId: "unusedPropType", Message: "'a.b' PropType is defined but prop is never used", Line: 5, Column: 1, EndLine: 5, EndColumn: 39}, {MessageId: "unusedPropType", Message: "'a.b.c' PropType is defined but prop is never used", Line: 5, Column: 1, EndLine: 5, EndColumn: 39}}},
			// Locks in upstream isPropUsed()'s nested-name comparison arm.
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.user.name}</div>; } } Hello.propTypes = { user: PropTypes.shape({ first: PropTypes.string, last: PropTypes.string }) };`, Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'user.first' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'user.last' PropType is defined but prop is never used"}}},
			// Locks in the custom validator option branch.
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = { unused: Custom.string };`, Options: map[string]interface{}{"customValidators": []interface{}{"Custom"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Locks in the direct propTypes member-assignment branch.
			{Code: `class Hello extends React.Component { render() { return <div />; } } Hello.propTypes = {}; Hello.propTypes.unused = PropTypes.string;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
		},
	)
}
