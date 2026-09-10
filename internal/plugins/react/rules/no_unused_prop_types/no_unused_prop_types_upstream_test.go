// TestNoUnusedPropTypesUpstream migrates representative valid and invalid
// cases from upstream tests/lib/rules/no-unused-prop-types.js 1:1. The
// rslint-specific edge-shape and branch lock-in cases live in the sibling
// no_unused_prop_types_extras_test.go file.
package no_unused_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPropTypesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedPropTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `var Hello = createReactClass({ propTypes: { name: PropTypes.string }, render: function() { return <div>Hello {this.props.name}</div>; } });`, Tsx: true},
			{Code: `var Hello = createReactClass({ propTypes: { name: PropTypes.object.isRequired }, render: function() { return <div>Hello {this.props.name.firstname}</div>; } });`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello World</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.firstname}</div>; } } Hello.propTypes = { firstname: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props?.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props["some.value"]}</div>; } } Hello.propTypes = { "some.value": PropTypes.string };`, Tsx: true},
			{Code: `function Hello(props) { return <div>Hello {props.name}</div>; } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `function Hello({ name }) { return <div>Hello {name}</div>; } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { static get propTypes() { return { name: PropTypes.string }; } render() { return <div>Hello {this.props.name}</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { static propTypes = { name: PropTypes.string }; render() { return <div>Hello {this.props.name}</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { var props = this.props; return <div>Hello {props.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { const { name } = this.props; return <div>Hello {name}</div>; } } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.name}</div>; } } Hello.propTypes = externalPropTypes;`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.name}</div>; } } Hello.propTypes = { ...externalPropTypes };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`, Options: map[string]interface{}{"ignore": []interface{}{"name"}}, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.user.name}</div>; } } Hello.propTypes = { user: PropTypes.shape({ name: PropTypes.string }) };`, Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.user.name}</div>; } } Hello.propTypes = { user: CustomValidator.shape({ name: CustomValidator.string }) };`, Options: map[string]interface{}{"customValidators": []interface{}{"CustomValidator"}, "skipShapeProps": false}, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			{Code: `var Hello = createReactClass({ propTypes: { unused: PropTypes.string }, render: function() { return <div>Hello</div>; } });`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used", Line: 1, Column: 45}}},
			{Code: `class Hello extends React.Component { static propTypes = { name: PropTypes.string }; render() { return <div>Hello {this.props.value}</div>; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'name' PropType is defined but prop is never used"}}},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.name}</div>; } } Hello.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `function Hello(props) { return <div>Hello {props.name}</div>; } Hello.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.a.z}</div>; } } Hello.propTypes = { a: PropTypes.shape({ b: PropTypes.string }) };`, Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'a.b' PropType is defined but prop is never used"}}},
			{Code: `class Hello extends React.Component { render() { return <div>Hello {this.props.a}</div>; } } Hello.propTypes = { a: PropTypes.string, b: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'b' PropType is defined but prop is never used"}}},
			{Code: `type Props = { first: string; second: string }; class Hello extends React.Component<Props> { render() { return <div>{this.props.first}</div>; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'second' PropType is defined but prop is never used"}}},
		},
	)
}
