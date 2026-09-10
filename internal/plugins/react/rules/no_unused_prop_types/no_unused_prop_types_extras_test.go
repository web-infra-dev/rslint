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
			// Locks in upstream isPropUsed()'s nested-name comparison arm.
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.user.name}</div>; } } Hello.propTypes = { user: PropTypes.shape({ first: PropTypes.string, last: PropTypes.string }) };`, Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'user.first' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'user.last' PropType is defined but prop is never used"}}},
			// Locks in the custom validator option branch.
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = { unused: Custom.string };`, Options: map[string]interface{}{"customValidators": []interface{}{"Custom"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Locks in the direct propTypes member-assignment branch.
			{Code: `class Hello extends React.Component { render() { return <div />; } } Hello.propTypes = {}; Hello.propTypes.unused = PropTypes.string;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
		},
	)
}
