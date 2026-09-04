// TestSortCompUpstream migrates the eslint-plugin-react v7.37.5
// tests/lib/rules/sort-comp.js suite. Rslint-specific cases live in the
// sibling sort_comp_extras_test.go file.
package sort_comp

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func sortCompError(propA, position, propB string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unsortedProps",
		Message:   fmt.Sprintf("%s should be placed %s %s", propA, position, propB),
	}
}

func sortCompOrder(entries ...string) map[string]any {
	values := make([]any, len(entries))
	for i, entry := range entries {
		values[i] = entry
	}
	return map[string]any{"order": values}
}

func TestSortCompUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortCompRule, []rule_tester.ValidTestCase{
		{Code: `
        // Must validate a full class
        var Hello = createReactClass({
          displayName : '',
          propTypes: {},
          contextTypes: {},
          childContextTypes: {},
          mixins: [],
          statics: {},
          getDefaultProps: function() {},
          getInitialState: function() {},
          getChildContext: function() {},
          componentWillMount: function() {},
          componentDidMount: function() {},
          componentWillReceiveProps: function() {},
          shouldComponentUpdate: function() {},
          componentWillUpdate: function() {},
          componentDidUpdate: function() {},
          componentWillUnmount: function() {},
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},
		{Code: `
        // Must validate a class with missing groups
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},
		{Code: `
        // Must put a custom method in 'everything-else'
        var Hello = createReactClass({
          onClick: function() {},
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          }
        });
      `, Tsx: true},
		{Code: `
        // Must allow us to re-order the groups
        var Hello = createReactClass({
          displayName : 'Hello',
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          },
          onClick: function() {}
        });
      `, Tsx: true, Options: sortCompOrder("lifecycle", "render", "everything-else")},
		{Code: `
        // Must validate a full React 16.3 createReactClass class
        var Hello = createReactClass({
          displayName : '',
          propTypes: {},
          contextTypes: {},
          childContextTypes: {},
          mixins: [],
          statics: {},
          getDefaultProps: function() {},
          getInitialState: function() {},
          getChildContext: function() {},
          UNSAFE_componentWillMount: function() {},
          componentDidMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          shouldComponentUpdate: function() {},
          UNSAFE_componentWillUpdate: function() {},
          getSnapshotBeforeUpdate: function() {},
          componentDidUpdate: function() {},
          componentDidCatch: function() {},
          componentWillUnmount: function() {},
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},
		{Code: `
        // Must validate React 16.3 lifecycle methods with the default parser
        class Hello extends React.Component {
          constructor() {}
          static getDerivedStateFromProps() {}
          UNSAFE_componentWillMount() {}
          componentDidMount() {}
          UNSAFE_componentWillReceiveProps() {}
          shouldComponentUpdate() {}
          UNSAFE_componentWillUpdate() {}
          getSnapshotBeforeUpdate() {}
          componentDidUpdate() {}
          componentDidCatch() {}
          componentWillUnmount() {}
          testInstanceMethod() {}
          render() { return (<div>Hello</div>); }
        }
      `, Tsx: true},
		{Code: `
        // Must validate a full React 16.3 ES6 class
        class Hello extends React.Component {
          static displayName = ''
          static propTypes = {}
          static defaultProps = {}
          constructor() {}
          state = {}
          static getDerivedStateFromProps = () => {}
          UNSAFE_componentWillMount = () => {}
          componentDidMount = () => {}
          UNSAFE_componentWillReceiveProps = () => {}
          shouldComponentUpdate = () => {}
          UNSAFE_componentWillUpdate = () => {}
          getSnapshotBeforeUpdate = () => {}
          componentDidUpdate = () => {}
          componentDidCatch = () => {}
          componentWillUnmount = () => {}
          testArrowMethod = () => {}
          testInstanceMethod() {}
          render = () => (<div>Hello</div>)
        }
      `, Tsx: true},
		{Code: `
        // Must allow us to create a RegExp-based group
        class Hello extends React.Component {
          customHandler() {}
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `, Tsx: true, Options: sortCompOrder("lifecycle", "everything-else", "render", "/on.*/")},
		{Code: `
        // Must allow us to create a named group
        class Hello extends React.Component {
          customHandler() {}
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `, Tsx: true, Options: map[string]any{"order": []any{"lifecycle", "everything-else", "render", "customGroup"}, "groups": map[string]any{"customGroup": []any{"/on.*/"}}}},
		{Code: `
        // Must allow a method to be in different places if it's matches multiple patterns
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `, Tsx: true, Options: sortCompOrder("/on.*/", "render", "/.*Click/")},
		{Code: `
        // Must allow us to use 'constructor' as a method name
        class Hello extends React.Component {
          constructor() {}
          displayName() {}
          render() {
            return <div>Hello</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("constructor", "lifecycle", "everything-else", "render")},
		{Code: `
        // Must ignore stateless components
        function Hello(props) {
          return <div>Hello {props.name}</div>
        }
      `, Tsx: true},
		{Code: `
        // Must ignore stateless components (arrow function with explicit return)
        var Hello = props => (
          <div>Hello {props.name}</div>
        )
      `, Tsx: true},
		{Code: `
        // Must ignore spread operator
        var Hello = createReactClass({
          ...proto,
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},
		{Code: `
        // Type Annotations should be first
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("type-annotations", "static-methods", "lifecycle", "everything-else", "render")},
		{Code: `
        // Properties with Type Annotations should not be at the top
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          state: Object = {};
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("type-annotations", "static-methods", "lifecycle", "everything-else", "render")},
		{Code: `
        // Non-react classes should be ignored, even in expressions
        return class Hello {
          render() {
            return <div>{this.props.text}</div>;
          }
          props: { text: string };
          constructor() {}
          state: Object = {};
        }
      `, Tsx: true},
		{Code: `
        // Non-react classes should be ignored, even in expressions
        return class {
          render() {
            return <div>{this.props.text}</div>;
          }
          props: { text: string };
          constructor() {}
          state: Object = {};
        }
      `, Tsx: true},
		{Code: `
        // Getters should be at the top
        class Hello extends React.Component {
          get foo() {}
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("getters", "static-methods", "lifecycle", "everything-else", "render")},
		{Code: `
        // Setters should be at the top
        class Hello extends React.Component {
          set foo(bar) {}
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("setters", "static-methods", "lifecycle", "everything-else", "render")},
		{Code: `
        // Instance methods should be at the top
        class Hello extends React.Component {
          foo = () => {}
          constructor() {}
          classMethod() {}
          static bar = () => {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-methods", "lifecycle", "everything-else", "render")},
		{Code: `
        // Instance variables should be at the top
        class Hello extends React.Component {
          foo = 'bar'
          constructor() {}
          state = {}
          static bar = 'foo'
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-variables", "lifecycle", "everything-else", "render")},
		{Code: `
        // Methods can be grouped with any matching group (with statics)
        class Hello extends React.Component {
          static onFoo() {}
          static renderFoo() {}
          render() {
            return <div>{this.props.text}</div>;
          }
          getFoo() {}
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "render", "/^get.+$/", "/^on.+$/", "/^render.+$/")},
		{Code: `
        // Methods can be grouped with any matching group (with RegExp)
        class Hello extends React.Component {
          render() {
            return <div>{this.props.text}</div>;
          }
          getFoo() {}
          static onFoo() {}
          static renderFoo() {}
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "render", "/^get.+$/", "/^on.+$/", "/^render.+$/")},
		{Code: `
        // static lifecycle methods can be grouped (with statics)
        class Hello extends React.Component {
          static getDerivedStateFromProps() {}
          constructor() {}
        }
      `, Tsx: true},
		{Code: `
        // static lifecycle methods can be grouped (with lifecycle)
        class Hello extends React.Component {
          constructor() {}
          static getDerivedStateFromProps() {}
        }
      `, Tsx: true},
		{Code: `
        class MyComponent extends React.Component {
          state = {};
          foo;
          static propTypes;

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("state", "instance-variables", "static-methods", "lifecycle", "render", "everything-else")},
		{Code: `
        class MyComponent extends React.Component {
          static foo;
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-variables", "static-methods")},
		{Code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static foo = 'some-str';

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "static-variables")},
		{Code: `
        class MyComponent extends React.Component {
          foo = {};
          static bar = 0;
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-variables", "static-variables", "static-methods")},
		{Code: `
        class MyComponent extends React.Component {
          static bar = 1;
          foo = {};
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-variables", "instance-variables", "static-methods")},
		{Code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          render() {
            return null;
          }
          static bar;
          foo = {};
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "render", "static-variables", "instance-variables")},
		{Code: `
        class MyComponent extends React.Component {
          static foo = 1;
          bar;

          constructor() {
            super(props);

            this.state = {};
          }

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-variables", "instance-variables", "constructor", "everything-else", "render")},
		{Code: `
        class ClassName extends React.Component {
          static defaultProps = {};
          static parseDateString(date?: Date) {}
          state = {};
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-variables", "static-methods", "type-annotations", "instance-variables", "lifecycle", "everything-else", "render")},
	}, []rule_tester.InvalidTestCase{
		{Code: `
        // Must force a lifecycle method to be placed before render
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          },
          displayName : 'Hello',
        });
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		{Code: `
        // Must run rule when render uses createElement instead of JSX
        var Hello = createReactClass({
          render: function() {
            return React.createElement("div", null, "Hello");
          },
          displayName : 'Hello',
        });
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		{Code: `
        // Must force a custom method to be placed before render
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          },
          onClick: function() {},
        });
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "onClick")}},
		{Code: `
        // Must force a custom method to be placed before render, even in function
        var Hello = () => {
          return class Test extends React.Component {
            render () {
              return <div>Hello</div>;
            }
            onClick () {}
          }
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "onClick")}},
		{Code: `
        // Must force a custom method to be placed after render if no 'everything-else' group is specified
        var Hello = createReactClass({
          displayName: 'Hello',
          onClick: function() {},
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          }
        });
      `, Tsx: true, Options: sortCompOrder("lifecycle", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("onClick", "after", "render")}},
		{Code: `
        // Must validate static properties
        class Hello extends React.Component {
          render() {
            return <div></div>
          }
          static displayName = 'Hello';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "displayName")}},
		{Code: `
        // Type Annotations should not be at the top by default
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          state: Object = {};
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{sortCompError("props", "after", "state")}},
		{Code: `
        // Type Annotations should be first
        class Hello extends React.Component {
          constructor() {}
          props: { text: string };
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("type-annotations", "static-methods", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("constructor", "after", "props")}},
		{Code: `
        // Properties with Type Annotations should not be at the top
        class Hello extends React.Component {
          props: { text: string };
          state: Object = {};
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("type-annotations", "static-methods", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("state", "after", "constructor")}},
		{Code: `
        // componentDidMountOk should be placed after getA
        export default class View extends React.Component {
          componentDidMountOk() {}
          getB() {}
          componentWillMount() {}
          getA() {}
          render() {}
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "lifecycle", "/^on.+$/", "/^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$/", "everything-else", "/^render.+$/", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("componentDidMountOk", "after", "getA")}},
		{Code: `
        // Getters should at the top
        class Hello extends React.Component {
          constructor() {}
          get foo() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("getters", "static-methods", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("constructor", "after", "getter functions")}},
		{Code: `
        // Setters should at the top
        class Hello extends React.Component {
          constructor() {}
          set foo(bar) {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("setters", "static-methods", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("constructor", "after", "setter functions")}},
		{Code: `
        // Instance methods should not be at the top
        class Hello extends React.Component {
          constructor() {}
          static bar = () => {}
          classMethod() {}
          foo = function() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-methods", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("foo", "before", "constructor")}},
		{Code: `
        // Instance variables should not be at the top
        class Hello extends React.Component {
          constructor() {}
          state = {}
          static bar = {}
          foo = {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-variables", "lifecycle", "everything-else", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("foo", "before", "constructor")}},
		{Code: `
        // Should not confuse method names with group names
        class Hello extends React.Component {
          setters() {}
          constructor() {}
          render() {}
        }
      `, Tsx: true, Options: sortCompOrder("setters", "lifecycle", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("setters", "after", "render")}},
		{Code: `
        // Explicitly named methods should appear in the correct order
        class Hello extends React.Component {
          render() {}
          foo() {}
        }
      `, Tsx: true, Options: sortCompOrder("foo", "render"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("render", "after", "foo")}},
		{Code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static foo;

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("static-variables", "static-methods"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("getDerivedStateFromProps", "after", "foo")}},
		{Code: `
        class MyComponent extends React.Component {
          static foo;
          bar = 'some-str'
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `, Tsx: true, Options: sortCompOrder("instance-variables", "static-variables", "static-methods"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("foo", "after", "bar")}},
		{Code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static bar;
          render() {
            return null;
          }
          foo = {};
        }
      `, Tsx: true, Options: sortCompOrder("static-methods", "render", "static-variables", "instance-variables"), Errors: []rule_tester.InvalidTestCaseError{sortCompError("bar", "after", "render")}},
	})
}
