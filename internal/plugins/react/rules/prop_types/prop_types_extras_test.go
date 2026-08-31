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
		}, []rule_tester.InvalidTestCase{
			{Code: `function Hello(props) { return <div>{props[0]}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType", Message: `'0' is missing in props validation`}}},
			{Code: `function Hello({ user }) { return <div>{user.name}</div>; } Hello.propTypes = { user: PropTypes.shape({}) };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
			{Code: `function Hello(props) { return <div>{props.missing}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingPropType"}}},
		})
}
