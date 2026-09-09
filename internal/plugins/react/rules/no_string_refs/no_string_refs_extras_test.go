package no_string_refs

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoStringRefsRuleExtras(t *testing.T) {
	legacy := map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoStringRefsRule, []rule_tester.ValidTestCase{
		// Intentional divergence: upstream checks ESTree `property.name` without
		// checking `computed`, so these variable/private-key accesses are false
		// positives there. Neither expression necessarily accesses React's public
		// `refs` property.
		{Code: `class Hello extends React.Component { componentDidMount() { const refs = "not-react-refs"; var c = this[refs]; } }`, Tsx: true, Settings: legacy},
		{Code: `class Hello extends React.Component { #refs; componentDidMount() { var c = this.#refs; } }`, Tsx: true, Settings: legacy},
		// Intentional divergence for the same reason: `createClass` is a variable
		// key, not necessarily the configured property name.
		{Code: `const createClass = "not-the-factory"; var Hello = React[createClass]({ componentDidMount() { var c = this.refs.foo; } });`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "createClass": "createClass"}}},
		// A source annotation overrides settings in both directions.
		{Code: `/** @jsx Preact.h */ class Hello extends Other.Component { componentDidMount() { var c = this.refs.foo; } }`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "pragma": "Other"}}},
	}, []rule_tester.InvalidTestCase{
		{Code: `/** @jsx Preact.h */ class Hello extends Preact.Component { componentDidMount() { var c = this.refs.foo; } }`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "pragma": "Other"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 91}}},
		{Code: `/** @jsx Preact.h */ var Hello = Preact.createClass({ componentDidMount() { var c = this.refs.foo; } });`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "pragma": "Other", "createClass": "createClass"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 85}}},
		{Code: `class Hello extends React.Component { componentDidMount() { var c = this.refs.foo; } }`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"defaultVersion": "18.2.0"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 69}}},
		{Code: `class Hello extends React.Component { componentDidMount() { var c = this.refs.foo; } }`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "detect", "defaultVersion": "18.2.0"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 69}}},
		{Code: `class Hello extends React.Component { componentDidMount() { var c = this.refs.foo; } }`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": 17}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 69}}},
		// String refs are deprecated on custom JSX components as well as intrinsic
		// elements; the listener intentionally does not filter by tag kind.
		{Code: `<Widget ref="instance" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "stringInRefDeprecated", Line: 1, Column: 9}}},
		{Code: `<UI.Widget ref={'instance'} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "stringInRefDeprecated", Line: 1, Column: 12}}},
		// ESTree erases these parentheses while ts-go retains them.
		{Code: `var Hello = createReactClass({ componentDidMount: (function() { var c = this.refs.foo; }) });`, Tsx: true, Settings: legacy, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 73}}},
		{Code: `var Hello = createReactClass({ componentDidMount: (() => this.refs.foo) });`, Tsx: true, Settings: legacy, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 58}}},
		// CallExpression and NewExpression both expose a callee in ESTree; the
		// create-react-class factory also returns its constructor explicitly.
		{Code: `var Hello = new createReactClass({ componentDidMount() { var c = this.refs.foo; } });`, Tsx: true, Settings: legacy, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 66}}},
		{Code: `var Hello = new React.createClass({ componentDidMount() { var c = this.refs.foo; } });`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "createClass": "createClass"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thisRefsDeprecated", Line: 1, Column: 67}}},
	})
}
