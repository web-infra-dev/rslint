package prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPropTypesRuleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PropTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = { user: PropTypes.shape({ name: PropTypes.string }) };`, Tsx: true},
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = { user: PropTypes.shape({ name: PropTypes.string }).isRequired };`, Tsx: true},
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = { user: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({ name: PropTypes.string })]) };`, Tsx: true},
			{Code: `function Hello({ ...rest }) { return <div>{rest.name}</div>; }`, Tsx: true},
			{Code: `function Hello({ name }) { return <div>{name}</div>; } Hello.propTypes = { ...externalPropTypes };`, Tsx: true},
			{Code: `function Hello({ name, children }) { return <div>{name}{children}</div>; }`, Options: map[string]interface{}{"ignore": []interface{}{"name", "children"}}, Tsx: true},
			{Code: `function Hello({ name }) { return <div>{name}</div>; }`, Options: map[string]interface{}{"skipUndeclared": true}, Tsx: true},
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = externalPropTypes;`, Tsx: true},
			{Code: `const Hello = props => <div>{props.name}</div>; Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.name.firstname}</div>; } } Hello.propTypes = { name: PropTypes.object.isRequired };`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.a.b}</div>; } Hello.propTypes = {}; Hello.propTypes.a = PropTypes.shape({ b: PropTypes.string });`, Tsx: true},
			{Code: `class Hello extends React.Component { static get propTypes() { return { name: PropTypes.string }; } render() { return <div>{this.props.name}</div>; } }`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			{Code: `function Hello(props) { return <div>{props[0]}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType", Message: `'0' is missing in props validation`}}},
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = { user: PropTypes.shape({}) };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
			{Code: `function Hello(props) { return <div>{props.missing}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
			{Code: `function Hello(props) { const { name } = props; return <div>{name}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType", Message: `'name' is missing in props validation`}}},
		})
}
