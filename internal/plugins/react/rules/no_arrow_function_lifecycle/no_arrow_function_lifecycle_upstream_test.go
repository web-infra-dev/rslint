package no_arrow_function_lifecycle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from eslint-plugin-react v7.37.5 without changing the upstream
// code, output, or error-message strings.
func TestNoArrowFunctionLifecycleRuleUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getDerivedStateFromProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps = () => { return {}; } // not a lifecycle method
          static getDerivedStateFromProps() {}
          render() { return <div />; }
        }
      `, Tsx: true},
		{Code: `
        var Hello = createReactClass({
          getDerivedStateFromProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true},
		{Code: `
        class MyComponent extends React.Component {
          onChange: () => void;
        }
      `, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		{Code: `
        var Hello = createReactClass({
          render: () => { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          getDefaultProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getDefaultProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          getInitialState: () => { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getInitialState is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          getChildContext: () => { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getChildContext is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentDidMount: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          shouldComponentUpdate: () => { return true; },
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `shouldComponentUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: () => { return {}; },
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getSnapshotBeforeUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentDidCatch: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidCatch is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        var Hello = createReactClass({
          componentWillUnmount: () => {},
          render: function() { return <div />; }
        });
      `, Tsx: true, Output: []string{`
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillUnmount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getDefaultProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState = () => { return {}; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getInitialState is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext = () => { return {}; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getChildContext is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getDerivedStateFromProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate = () => { return true; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `shouldComponentUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `UNSAFE_componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate = () => { return {}; }
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getSnapshotBeforeUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate = (prevProps) => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate(prevProps) {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentDidCatch is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount = () => {}
          render = () => { return <div />; }
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `componentWillUnmount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}, {MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          render = () => <div />
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          render() { return <div />; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        class Hello extends React.Component {
          render = () => /*first*/<div />/*second*/
        }
      `, Tsx: true, Output: []string{`
        class Hello extends React.Component {
          render() { return /*first*/<div />/*second*/; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
		{Code: `
        export default class Root extends Component {
          getInitialState = () => ({
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          });
        }
      `, Tsx: true, Output: []string{`
        export default class Root extends Component {
          getInitialState() { return {
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          }; }
        }
      `}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lifecycle", Message: `getInitialState is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`}}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoArrowFunctionLifecycleRule, valid, invalid)
}
