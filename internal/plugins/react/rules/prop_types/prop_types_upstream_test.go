package prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPropTypesRuleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PropTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `function Hello({ name }) { return <div>{name}</div>; } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `var Hello = createReactClass({ propTypes: { name: PropTypes.string }, render: function() { return <div>{this.props.name}</div>; } });`, Tsx: true},
			{Code: `function Hello() { return <div>Hello</div>; }`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			{Code: `function Hello({ name }) { return <div>{name}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
			{Code: `var Hello = createReactClass({ render: function() { return <div>{this.props.name}</div>; } });`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
		})
}
