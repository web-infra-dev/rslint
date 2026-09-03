// This file contains rslint-specific edge-shape, regression, and branch
// lock-in cases. The upstream mirror is in no_arrow_function_lifecycle_upstream_test.go.
package no_arrow_function_lifecycle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const upstreamLifecycleMessage = " is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead."

func TestNoArrowFunctionLifecycleRuleExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 1: class expression ----
		{Code: "const Hello = class extends React.Component { render() { return null; } };", Tsx: true},
		// ---- Dimension 1: async arrow and empty class body ----
		{Code: "class Hello extends React.Component { handle = async () => {}; }", Tsx: true},
		{Code: "class Hello extends React.Component {}", Tsx: true},
		// ---- Dimension 1: type-only field has no initializer ----
		{Code: "class Hello extends React.Component { onChange: () => void; }", Tsx: true},
		// ---- Dimension 2: non-React nested class must not be treated as a component ----
		{Code: "class Outer extends React.Component { render() { return class Inner { render = () => null; }; } }", Tsx: true},
		// ---- Dimension 2: an unrelated class remains outside the component set ----
		{Code: "class Hello { render = () => null; }", Tsx: true},
		// ---- Dimension 2: parenthesized class heritage is transparent ----
		{Code: "class Hello extends (React.Component) { render() { return null; } }", Tsx: true},
		// ---- Dimension 2: configured pragma and createClass names ----
		{Code: "class Hello extends React.Component { render = () => null; }", Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},
		{Code: "const Hello = createReactClass({ render: () => null });", Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "makeClass"}}},
		// ---- Dimension 4: string and computed keys are not `key.name` ----
		{Code: "class Hello extends React.Component { ['render'] = () => null; }", Tsx: true},
		{Code: "var Hello = createReactClass({ ['render']: () => null });", Tsx: true},
		{Code: "var Hello = createReactClass({ [name]: () => null });", Tsx: true},
		// ---- Dimension 4: spread members do not mask sibling properties ----
		{Code: "var Hello = createReactClass({ ...defaults, render() { return null; } });", Tsx: true},
		// ---- Dimension 4: overload-like / body-absent members are ignored ----
		{Code: "abstract class Hello extends React.Component { abstract render(): unknown; }", Tsx: true},
		// ---- Branch lock-in: a non-createReactClass object is not a component ----
		{Code: "const options = { render: () => null };", Tsx: true},
		// ---- Branch lock-in: object method values are not arrow functions ----
		{Code: "var Hello = createReactClass({ render() { return null; } });", Tsx: true},
		// ---- Regression: TypeScript field without an initializer ----
		{Code: "import * as React from 'react'; class ParameterEditor extends React.Component { onChange: () => void; render() { return null; } }", Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 3: expression body ----
		{
			Code:   "class Hello extends React.Component { render = (() => <div />) }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { render() { return <div />; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage, Line: 1, Column: 39, EndLine: 1, EndColumn: 63}},
		},
		// ---- Regression: parenthesized object expression body ----
		{
			Code:   "export default class Root extends Component {\n  getInitialState = () => ({\n    errorImporting: null,\n    file: null,\n  });\n}",
			Tsx:    true,
			Output: []string{"export default class Root extends Component {\n  getInitialState() { return {\n    errorImporting: null,\n    file: null,\n  }; }\n}"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "getInitialState" + upstreamLifecycleMessage}},
		},
		// ---- Dimension 4: private identifier names behave like ESTree names ----
		{
			Code:   "class Hello extends React.Component { #render = () => null; }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { #render() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "class Hello extends React.Component { [render] = () => null; }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { [render]() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "createReactClass({ [render]: () => null });",
			Tsx:    true,
			Output: []string{"createReactClass({ [render]: function() { return null; } });"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		// ---- Branch lock-in: static members use the separate static lifecycle list ----
		{
			Code:   "class Hello extends React.Component { static getDerivedStateFromProps = () => null; }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { static getDerivedStateFromProps() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "getDerivedStateFromProps" + upstreamLifecycleMessage, Line: 1, Column: 39, EndLine: 1, EndColumn: 84}},
		},
		{
			Code:   "class Hello extends React.Component { static [getDerivedStateFromProps] = () => null; }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { static [getDerivedStateFromProps]() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "getDerivedStateFromProps" + upstreamLifecycleMessage}},
		},
		{
			Code:   "createReactClass(factory(), { render: () => null });",
			Tsx:    true,
			Output: []string{"createReactClass(factory(), { render: function() { return null; } });"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "class Hello extends React.Component { render = (value = 1, ...rest) => value; }",
			Tsx:    true,
			Output: []string{"class Hello extends React.Component { render(value = 1, ...rest) { return value; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "var Hello = createReactClass({ render: ({ value }, [item]) => value });",
			Tsx:    true,
			Output: []string{"var Hello = createReactClass({ render: function({ value }, [item]) { return value; } });"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "/* @jsx Preact.h */ class Hello extends Preact.Component { render = () => null; }",
			Tsx:    true,
			Output: []string{"/* @jsx Preact.h */ class Hello extends Preact.Component { render() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "/** @extends React.Component */ class Hello { render = () => null; }",
			Tsx:    true,
			Output: []string{"/** @extends React.Component */ class Hello { render() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
		{
			Code:   "/** @augments React.PureComponent */ class Hello { render = () => null; }",
			Tsx:    true,
			Output: []string{"/** @augments React.PureComponent */ class Hello { render() { return null; } }"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: "render" + upstreamLifecycleMessage}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrowFunctionLifecycleRule, valid, invalid)
}
