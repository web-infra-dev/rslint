package no_multi_comp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMultiCompAssignmentName(t *testing.T) {
	// Checked against eslint-plugin-react v7.37.5: an identifier assignment
	// uses the binding's name even when its function expression has a name.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMultiCompRule, []rule_tester.ValidTestCase{
		{Code: `
        var helper;
        helper = function NamedComp() { return <div /> }
        class App extends React.Component { render() { return <div /> } }
      `, Tsx: true},
	}, nil)
}
