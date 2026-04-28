package no_unused_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedStateRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: stateless function component ----
		{Code: `
        function StatelessFnUnaffectedTest(props) {
          return <SomeComponent foo={props.foo} />;
        };
      `, Tsx: true},

		// ---- Upstream: createReactClass without state ----
		{Code: `
        var NoStateTest = createReactClass({
          render: function() {
            return <SomeComponent />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: createReactClass without state (shorthand method) ----
		{Code: `
        var NoStateMethodTest = createReactClass({
          render() {
            return <SomeComponent />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: getInitialState with used state ----
		{Code: `
        var GetInitialStateTest = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from variable ----
		{Code: `
        var ComputedKeyFromVariableTest = createReactClass({
          getInitialState: function() {
            return { [foo]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from boolean literal ----
		{Code: `
        var ComputedKeyFromBooleanLiteralTest = createReactClass({
          getInitialState: function() {
            return { [true]: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state[true]} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from number literal ----
		{Code: `
        var ComputedKeyFromNumberLiteralTest = createReactClass({
          getInitialState: function() {
            return { [123]: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state[123]} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from expression ----
		{Code: `
        var ComputedKeyFromExpressionTest = createReactClass({
          getInitialState: function() {
            return { [foo + bar]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from binary expression ----
		{Code: `
        var ComputedKeyFromBinaryExpressionTest = createReactClass({
          getInitialState: function() {
            return { ['foo' + 'bar' * 8]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from string literal ----
		{Code: `
        var ComputedKeyFromStringLiteralTest = createReactClass({
          getInitialState: function() {
            return { ['foo']: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: computed key from template literal with expression ----
		{Code: "var ComputedKeyFromTemplateLiteralTest = createReactClass({\n  getInitialState: function() {\n    return { [`foo${bar}`]: 0 };\n  },\n  render: function() {\n    return <SomeComponent />;\n  }\n});", Tsx: true},

		// ---- Upstream: computed key from template literal (no substitution) used ----
		{Code: "var ComputedKeyFromTemplateLiteralTest = createReactClass({\n  getInitialState: function() {\n    return { [`foo`]: 0 };\n  },\n  render: function() {\n    return <SomeComponent foo={this.state['foo']} />;\n  }\n});", Tsx: true},

		// ---- Upstream: getInitialState shorthand method ----
		{Code: `
        var GetInitialStateMethodTest = createReactClass({
          getInitialState() {
            return { foo: 0 };
          },
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState with used state ----
		{Code: `
        var SetStateTest = createReactClass({
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          },
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: multiple setState with used state ----
		{Code: `
        var MultipleSetState = createReactClass({
          getInitialState() {
            return { foo: 0 };
          },
          update() {
            this.setState({foo: 1});
          },
          render() {
            return <SomeComponent onClick={this.update} foo={this.state.foo} />;
          }
        });
      `, Tsx: true},

		// ---- Upstream: class component without state ----
		{Code: `
        class NoStateTest extends React.Component {
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: constructor state with usage ----
		{Code: `
        class CtorStateTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from variable ----
		{Code: `
        class ComputedKeyFromVariableTest extends React.Component {
          constructor() {
            this.state = { [foo]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from boolean literal ----
		{Code: `
        class ComputedKeyFromBooleanLiteralTest extends React.Component {
          constructor() {
            this.state = { [false]: 0 };
          }
          render() {
            return <SomeComponent foo={this.state['false']} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from number literal ----
		{Code: `
        class ComputedKeyFromNumberLiteralTest extends React.Component {
          constructor() {
            this.state = { [345]: 0 };
          }
          render() {
            return <SomeComponent foo={this.state[345]} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from expression ----
		{Code: `
        class ComputedKeyFromExpressionTest extends React.Component {
          constructor() {
            this.state = { [foo + bar]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from binary expression ----
		{Code: `
        class ComputedKeyFromBinaryExpressionTest extends React.Component {
          constructor() {
            this.state = { [1 + 2 * 8]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from string literal ----
		{Code: `
        class ComputedKeyFromStringLiteralTest extends React.Component {
          constructor() {
            this.state = { ['foo']: 0 };
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class computed key from template literal with expression ----
		{Code: "class ComputedKeyFromTemplateLiteralTest extends React.Component {\n  constructor() {\n    this.state = { [`foo${bar}`]: 0 };\n  }\n  render() {\n    return <SomeComponent />;\n  }\n}", Tsx: true},

		// ---- Upstream: class computed key from template literal used ----
		{Code: "class ComputedKeyFromTemplateLiteralTest extends React.Component {\n  constructor() {\n    this.state = { [`foo`]: 0 };\n  }\n  render() {\n    return <SomeComponent foo={this.state.foo} />;\n  }\n}", Tsx: true},

		// ---- Upstream: class setState with usage ----
		{Code: `
        class SetStateTest extends React.Component {
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: class property state ----
		{Code: `
        class ClassPropertyStateTest extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: optional chaining ----
		{Code: `
        class OptionalChaining extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent foo={this.state?.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: variable declaration ----
		{Code: `
        class VariableDeclarationTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const foo = this.state.foo;
            return <SomeComponent foo={foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: destructuring ----
		{Code: `
        class DestructuringTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {foo: myFoo} = this.state;
            return <SomeComponent foo={myFoo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: shorthand destructuring ----
		{Code: `
        class ShorthandDestructuringTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {foo} = this.state;
            return <SomeComponent foo={foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: alias declaration ----
		{Code: `
        class AliasDeclarationTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const state = this.state;
            return <SomeComponent foo={state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: alias assignment ----
		{Code: `
        class AliasAssignmentTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            let state;
            state = this.state;
            return <SomeComponent foo={state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: destructuring alias ----
		{Code: `
        class DestructuringAliasTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {state: myState} = this;
            return <SomeComponent foo={myState.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: shorthand destructuring alias ----
		{Code: `
        class ShorthandDestructuringAliasTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {state} = this;
            return <SomeComponent foo={state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: rest property ----
		{Code: `
        class RestPropertyTest extends React.Component {
          constructor() {
            this.state = {
              foo: 0,
              bar: 1,
            };
          }
          render() {
            const {foo, ...others} = this.state;
            return <SomeComponent foo={foo} bar={others.bar} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: deep destructuring ----
		{Code: `
        class DeepDestructuringTest extends React.Component {
          state = { foo: 0, bar: 0 };
          render() {
            const {state: {foo, ...others}} = this;
            return <SomeComponent foo={foo} bar={others.bar} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — method argument ----
		{Code: `
        class MethodArgFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          consumeFoo(foo) {}
          render() {
            this.consumeFoo(this.state.foo);
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — assigned to object ----
		{Code: `
        class AssignedToObjectFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const obj = { foo: this.state.foo, bar: 0 };
            return <SomeComponent bar={obj.bar} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — computed access ----
		{Code: `
        class ComputedAccessFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0, bar: 1 };
          }
          render() {
            const bar = 'bar';
            return <SomeComponent bar={this.state[bar]} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — JSX spread ----
		{Code: `
        class JsxSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent {...this.state} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — aliased JSX spread ----
		{Code: `
        class AliasedJsxSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const state = this.state;
            return <SomeComponent {...state} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — object spread ----
		{Code: `
        class ObjectSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const attrs = { ...this.state, foo: 1 };
            return <SomeComponent foo={attrs.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — shadowing ----
		{Code: `
        class ShadowingFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const state = this.state;
            let foo;
            {
              const state = { foo: 5 };
              foo = state.foo;
            }
            return <SomeComponent foo={foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: false negative — non-render class method ----
		{Code: `
        class NonRenderClassMethodFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0, bar: 0 };
          }
          doSomething() {
            const { foo } = this.state;
            return this.state.foo;
          }
          doSomethingElse() {
            const { state: { bar }} = this;
            return bar;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow function class method destructuring false negative ----
		{Code: `
        class ArrowFunctionClassMethodDestructuringFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          doSomething = () => {
            const { state: { foo } } = this;
            return foo;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow function class method with class property ----
		{Code: `
        class ArrowFunctionClassMethodWithClassPropertyTransformFalseNegativeTest extends React.Component {
          state = { foo: 0 };
          doSomething = () => {
            const { state:{ foo } } = this;
            return foo;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow function deep destructuring ----
		{Code: `
        class ArrowFunctionClassMethodDeepDestructuringFalseNegativeTest extends React.Component {
          state = { foo: { bar: 0 } };
          doSomething = () => {
            const { state: { foo: { bar }}} = this;
            return bar;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: arrow function destructuring assignment ----
		{Code: `
        class ArrowFunctionClassMethodDestructuringAssignmentFalseNegativeTest extends React.Component {
          state = { foo: 0 };
          doSomething = () => {
            const { state: { foo: bar }} = this;
            return bar;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: this.state as an object (call arg) ----
		{Code: `
        class ThisStateAsAnObject extends React.Component {
          state = {
            active: true
          };
          render() {
            return <div className={classNames('overflowEdgeIndicator', className, this.state)} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: getDerivedStateFromProps ----
		{Code: `
        class GetDerivedStateFromPropsTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123,
            };
          }
          static getDerivedStateFromProps(nextProps, otherState) {
            if (otherState.id === nextProps.id) {
              return {
                selected: true,
              };
            }
            return null;
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: componentDidUpdate ----
		{Code: `
        class ComponentDidUpdateTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123,
            };
          }
          componentDidUpdate(someProps, someState) {
            if (someState.id === someProps.id) {
              doStuff();
            }
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: shouldComponentUpdate ----
		{Code: `
        class ShouldComponentUpdateTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123,
            };
          }
          shouldComponentUpdate(nextProps, nextState) {
            return nextState.id === nextProps.id;
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: nested scopes in lifecycle ----
		{Code: `
        class NestedScopesTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123,
            };
          }
          shouldComponentUpdate(nextProps, nextState) {
            return (function() {
              return nextState.id === nextProps.id;
            })();
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: setState callback with state param ----
		{Code: `
        class Foo extends Component {
          state = {
            initial: 'foo',
          }
          handleChange = () => {
            this.setState(state => ({
              current: state.initial
            }));
          }
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        }
      `, Tsx: true},

		// ---- Upstream: constructor state with setState callback ----
		{Code: `
        class Foo extends Component {
          constructor(props) {
            super(props);
            this.state = {
              initial: 'foo',
            }
          }
          handleChange = () => {
            this.setState(state => ({
              current: state.initial
            }));
          }
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        }
      `, Tsx: true},

		// ---- Upstream: setState callback with two params ----
		{Code: `
        class Foo extends Component {
          constructor(props) {
            super(props);
            this.state = {
              initial: 'foo',
            }
          }
          handleChange = () => {
            this.setState((state, props) => ({
              current: state.initial
            }));
          }
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        }
      `, Tsx: true},

		// ---- Upstream: ES5 setState callback ----
		{Code: `
        var Foo = createReactClass({
          getInitialState: function() {
            return { initial: 'foo' };
          },
          handleChange: function() {
            this.setState(state => ({
              current: state.initial
            }));
          },
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        });
      `, Tsx: true},

		// ---- Upstream: ES5 setState callback with two params ----
		{Code: `
        var Foo = createReactClass({
          getInitialState: function() {
            return { initial: 'foo' };
          },
          handleChange: function() {
            this.setState((state, props) => ({
              current: state.initial
            }));
          },
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        });
      `, Tsx: true},

		// ---- Upstream: setState destructuring callback ----
		{Code: `
        class SetStateDestructuringCallback extends Component {
          state = {
              used: 1, unused: 2
          }
          handleChange = () => {
            this.setState(({unused}) => ({
              used: unused * unused,
            }));
          }
          render() {
            return <div>{this.state.used}</div>
          }
        }
      `, Tsx: true},

		// ---- Upstream: setState callback state condition ----
		{Code: `
        class SetStateCallbackStateCondition extends Component {
          state = {
              isUsed: true,
              foo: 'foo'
          }
          handleChange = () => {
            this.setState((prevState) => (prevState.isUsed ? {foo: 'bar', isUsed: false} : {}));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: setState in regular function expression (no arrow) — don't error ----
		{Code: `
        class Foo extends Component {
          handleChange = function() {
            this.setState(() => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: setState with state param in function expression ----
		{Code: `
        class Foo extends Component {
          handleChange = function() {
            this.setState(state => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: static handler with setState ----
		{Code: `
        class Foo extends Component {
          static handleChange = () => {
            this.setState(state => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Upstream: TS as/unknown expressions ----
		{Code: `
        class Foo extends Component {
          state = {
            thisStateAliasProp,
            thisStateAliasRestProp,
            thisDestructStateAliasProp,
            thisDestructStateAliasRestProp,
            thisDestructStateDestructRestProp,
            thisSetStateProp,
            thisSetStateRestProp,
          } as unknown

          constructor() {
            ((this as unknown).state as unknown) = { thisStateProp } as unknown;
            ((this as unknown).setState as unknown)({ thisStateDestructProp } as unknown);
            ((this as unknown).setState as unknown)(state => ({ thisDestructStateDestructProp } as unknown));
          }

          thisStateAlias() {
            const state = (this as unknown).state as unknown;

            (state as unknown).thisStateAliasProp as unknown;
            const { ...thisStateAliasRest } = state as unknown;
            (thisStateAliasRest as unknown).thisStateAliasRestProp as unknown;
          }

          thisDestructStateAlias() {
            const { state } = this as unknown;

            (state as unknown).thisDestructStateAliasProp as unknown;
            const { ...thisDestructStateAliasRest } = state as unknown;
            (thisDestructStateAliasRest as unknown).thisDestructStateAliasRestProp as unknown;
          }

          thisSetState() {
            ((this as unknown).setState as unknown)(state => (state as unknown).thisSetStateProp as unknown);
            ((this as unknown).setState as unknown)(({ ...thisSetStateRest }) => (thisSetStateRest as unknown).thisSetStateRestProp as unknown);
          }

          render() {
            ((this as unknown).state as unknown).thisStateProp as unknown;
            const { thisStateDestructProp } = (this as unknown).state as unknown;
            const { state: { thisDestructStateDestructProp, ...thisDestructStateDestructRest } } = this as unknown;
            (thisDestructStateDestructRest as unknown).thisDestructStateDestructRestProp as unknown;

            return null;
          }
        }
      `, Tsx: true},

		// ---- Upstream: getDerivedStateFromProps as class property (arrow) ----
		{Code: `
        class TestNoUnusedState extends React.Component {
          constructor(props) {
            super(props);
            this.state = {
              id: null,
            };
          }
          static getDerivedStateFromProps = (props, state) => {
            if (state.id !== props.id) {
              return {
                id: props.id,
              };
            }
            return null;
          };
          render() {
            return <h1>{this.state.id}</h1>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: getDerivedStateFromProps with destructuring ----
		{Code: `
        class Component2 extends React.Component {
          static getDerivedStateFromProps = ({value, disableAnimation}, {isControlled, isOn}) => {
            return { isControlled, isOn };
          };
          render() {
            const { isControlled, isOn } = this.state;
            return <div>{isControlled ? 'controlled' : ''}{isOn ? 'on' : ''}</div>;
          }
        }
      `, Tsx: true},

		// ---- Upstream: cancel button state ----
		{Code: `
        class Foo extends React.Component {
          onCancel = (data) => {
            console.log('Cancelled', data)
            this.setState({ status: 'Cancelled. Try again?' })
          }
          render() {
            const { status } = this.state;
            return <div>{status}</div>
          }
        }
      `, Tsx: true},

		// ---- Edge: non-component class with state ----
		{Code: `
        class NotAComponent {
          state = { foo: 0 };
          render() {
            return null;
          }
        }
      `, Tsx: true},

		// ---- Edge: class expression component ----
		{Code: `
        var Hello = class extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        };
      `, Tsx: true},

		// ---- Edge: nested class — inner component state is independent ----
		{Code: `
        class Outer extends React.Component {
          state = { outerFoo: 0 };
          render() {
            class Inner extends React.Component {
              state = { innerBar: 0 };
              render() {
                return <div>{this.state.innerBar}</div>;
              }
            }
            return <div>{this.state.outerFoo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: parenthesized this.state access ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent foo={(this).state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Edge: TS satisfies expression ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 } satisfies { foo: number };
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `, Tsx: true},

		// ---- Edge: lifecycle state param destructured (componentDidUpdate) ----
		{Code: `
        class Hello extends Component {
          constructor(props) {
            super(props);
            this.state = { id: 123 };
          }
          componentDidUpdate(prevProps, prevState) {
            const { id } = prevState;
            console.log(id);
          }
          render() {
            return <h1>{this.state.id}</h1>;
          }
        }
      `, Tsx: true},

		// ---- Edge: lifecycle state param aliased (shouldComponentUpdate) ----
		{Code: `
        class Hello extends Component {
          constructor(props) {
            super(props);
            this.state = { count: 0 };
          }
          shouldComponentUpdate(nextProps, nextState) {
            const s = nextState;
            return s.count !== this.state.count;
          }
          render() {
            return <div>{this.state.count}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: GDSFP with shadowed state param (TypeChecker resolves correctly) ----
		{Code: `
        class Hello extends React.Component {
          state = { count: 0 };
          static getDerivedStateFromProps(props, state) {
            const inner = () => {
              const state = { local: true };
              console.log(state.local);
            };
            return state.count > 0 ? { count: state.count } : null;
          }
          render() {
            return <div>{this.state.count}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: TS readonly state with generics ----
		{Code: `
        interface Props {}
        interface State {
          flag: boolean;
        }
        class RuleTest extends React.Component<Props, State> {
          readonly state: State = {
            flag: false,
          };
          static getDerivedStateFromProps = (props: Props, state: State) => {
            const newState: Partial<State> = {};
            if (!state.flag) {
              newState.flag = true;
            }
            return newState;
          };
        }
      `, Tsx: true},

		// ---- Edge: non-arrow class property initializer using state (should not false-positive) ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          myProp = this.state.foo;
          render() {
            return <div />;
          }
        }
      `, Tsx: true},

		// ---- Edge: FunctionExpression class property using state ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          handleClick = function() {
            return this.state.foo;
          }
          render() {
            return <div />;
          }
        }
      `, Tsx: true},

		// ---- Edge: static non-GDSFP arrow using state ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          static helper = () => {
            return null;
          }
          render() {
            return <div>{this.state.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: ES5 arrow property value using state ----
		{Code: `
        var Hello = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          handler: () => this.state.foo,
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},

		// ---- Edge: this.state[true] / this.state[false] / this.state[null] should not give up ----
		{Code: `
        class Hello extends React.Component {
          constructor() {
            this.state = { [true]: 0 };
          }
          render() {
            return <div>{this.state[true]}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: ES5 getter using state ----
		{Code: `
        var Hello = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          get computed() {
            return this.state.foo;
          },
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},

		// ---- Edge: this.state['foo'] string element access ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent foo={this.state['foo']} />;
          }
        }
      `, Tsx: true},

		// ---- Edge: getSnapshotBeforeUpdate lifecycle param ----
		{Code: `
        class Hello extends Component {
          constructor(props) {
            super(props);
            this.state = { scroll: 0 };
          }
          getSnapshotBeforeUpdate(prevProps, prevState) {
            return prevState.scroll;
          }
          render() {
            return <div>{this.state.scroll}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: UNSAFE_componentWillUpdate lifecycle param ----
		{Code: `
        class Hello extends Component {
          constructor(props) {
            super(props);
            this.state = { val: 0 };
          }
          UNSAFE_componentWillUpdate(nextProps, nextState) {
            console.log(nextState.val);
          }
          render() {
            return <div>{this.state.val}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: constructor assigns non-object to this.state (should not track) ----
		{Code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props);
            this.state = getInitialState();
          }
          render() {
            return <div>{this.state.foo}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: state used only in non-render method ----
		{Code: `
        class Hello extends React.Component {
          state = { token: 'abc' };
          fetchData() {
            fetch('/api', { headers: { auth: this.state.token } });
          }
          render() {
            return <div />;
          }
        }
      `, Tsx: true},

		// ---- Edge: multiple independent components in same file ----
		{Code: `
        class CompA extends React.Component {
          state = { a: 1 };
          render() { return <div>{this.state.a}</div>; }
        }
        class CompB extends React.Component {
          state = { b: 2 };
          render() { return <div>{this.state.b}</div>; }
        }
      `, Tsx: true},

		// ---- Edge: setState arrow return field used ----
		{Code: `
        class Hello extends Component {
          handleChange = () => {
            this.setState(state => ({
              derived: state.count + 1
            }));
          }
          render() {
            return <div>{this.state.derived}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: state used in ternary expression ----
		{Code: `
        class Hello extends React.Component {
          state = { loading: false };
          render() {
            return this.state.loading ? <div>Loading</div> : <div>Done</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: state used in logical expression ----
		{Code: `
        class Hello extends React.Component {
          state = { error: null };
          render() {
            return <div>{this.state.error && <span>Error</span>}</div>;
          }
        }
      `, Tsx: true},

		// ---- Edge: state used in template literal ----
		{Code: "class Hello extends React.Component {\n  state = { name: '' };\n  render() {\n    return <div>{`Hello ${this.state.name}`}</div>;\n  }\n}", Tsx: true},

		// ---- Edge: state alias in arrow class property does NOT leak to other methods ----
		{Code: `
        class Hello extends React.Component {
          state = { foo: 0, bar: 1 };
          method1 = () => {
            const s = this.state;
            return s.foo;
          }
          method2() {
            return this.state.bar;
          }
          render() {
            return <div />;
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: unused getInitialState ----
		{
			Code: `
        var UnusedGetInitialStateTest = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 22},
			},
		},

		// ---- Upstream: unused computed string literal key ----
		{
			Code: `
        var UnusedComputedStringLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { ['foo']: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 22},
			},
		},

		// ---- Upstream: unused computed template literal key ----
		{
			Code: "var UnusedComputedTemplateLiteralKeyStateTest = createReactClass({\n  getInitialState: function() {\n    return { [`foo`]: 0 };\n  },\n  render: function() {\n    return <SomeComponent />;\n  }\n})",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 3, Column: 14},
			},
		},

		// ---- Upstream: unused computed number literal key ----
		{
			Code: `
        var UnusedComputedNumberLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { [123]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: '123'", Line: 4, Column: 22},
			},
		},

		// ---- Upstream: unused computed boolean literal key ----
		{
			Code: `
        var UnusedComputedBooleanLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { [true]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'true'", Line: 4, Column: 22},
			},
		},

		// ---- Upstream: unused getInitialState (shorthand) ----
		{
			Code: `
        var UnusedGetInitialStateMethodTest = createReactClass({
          getInitialState() {
            return { foo: 0 };
          },
          render() {
            return <SomeComponent />;
          }
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 22},
			},
		},

		// ---- Upstream: unused setState ----
		{
			Code: `
        var UnusedSetStateTest = createReactClass({
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          },
          render() {
            return <SomeComponent />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 29},
			},
		},

		// ---- Upstream: unused class constructor state ----
		{
			Code: `
        class UnusedCtorStateTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 28},
			},
		},

		// ---- Upstream: unused class setState ----
		{
			Code: `
        class UnusedSetStateTest extends React.Component {
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 29},
			},
		},

		// ---- Upstream: unused class property state ----
		{
			Code: `
        class UnusedClassPropertyStateTest extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 3, Column: 21},
			},
		},

		// ---- Upstream: unused computed string literal key (class) ----
		{
			Code: `
        class UnusedComputedStringLiteralKeyStateTest extends React.Component {
          state = { ['foo']: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 3, Column: 21},
			},
		},

		// ---- Upstream: unused computed template literal key (class) ----
		{
			Code: "class UnusedComputedTemplateLiteralKeyStateTest extends React.Component {\n  state = { [`foo`]: 0 };\n  render() {\n    return <SomeComponent />;\n  }\n}",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 2, Column: 13},
			},
		},

		// ---- Upstream: unused computed boolean literal key (class) ----
		{
			Code: `
        class UnusedComputedBooleanLiteralKeyStateTest extends React.Component {
          state = { [true]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'true'", Line: 3, Column: 21},
			},
		},

		// ---- Upstream: unused computed number literal key (class) ----
		{
			Code: `
        class UnusedComputedNumberLiteralKeyStateTest extends React.Component {
          state = { [123]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: '123'", Line: 3, Column: 21},
			},
		},

		// ---- Upstream: unused state when props are spread ----
		{
			Code: `
        class UnusedStateWhenPropsAreSpreadTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent {...this.props} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 28},
			},
		},

		// ---- Upstream: alias out of scope ----
		{
			Code: `
        class AliasOutOfScopeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const state = this.state;
            return <SomeComponent />;
          }
          someMethod() {
            const outOfScope = state.foo;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 28},
			},
		},

		// ---- Upstream: multiple errors ----
		{
			Code: `
        class MultipleErrorsTest extends React.Component {
          constructor() {
            this.state = {
              foo: 0,
              bar: 1,
              baz: 2,
              qux: 3,
            };
          }
          render() {
            let {state} = this;
            return <SomeComponent baz={state.baz} qux={state.qux} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 5, Column: 15},
				{MessageId: "unusedStateField", Message: "Unused state field: 'bar'", Line: 6, Column: 15},
			},
		},

		// ---- Upstream: multiple errors for same key ----
		{
			Code: `
        class MultipleErrorsForSameKeyTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 28},
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 7, Column: 29},
			},
		},

		// ---- Upstream: unused rest property field ----
		{
			Code: `
        class UnusedRestPropertyFieldTest extends React.Component {
          constructor() {
            this.state = {
              foo: 0,
              bar: 1,
            };
          }
          render() {
            const {bar, ...others} = this.state;
            return <SomeComponent bar={bar} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 5, Column: 15},
			},
		},

		// ---- Upstream: unused state with arrow function method ----
		{
			Code: `
        class UnusedStateArrowFunctionMethodTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          doSomething = () => {
            return null;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 28},
			},
		},

		// ---- Upstream: unused deep destructuring ----
		{
			Code: `
        class UnusedDeepDestructuringTest extends React.Component {
          state = { foo: 0, bar: 0 };
          render() {
            const {state: {foo}} = this;
            return <SomeComponent foo={foo} />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'bar'", Line: 3, Column: 29},
			},
		},

		// ---- Upstream: fake prevState variable ----
		{
			Code: `
        class FakePrevStateVariableTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123,
              foo: 456
            };
          }
          componentDidUpdate(someProps, someState) {
            if (someState.id === someProps.id) {
              const prevState = { foo: 789 };
              console.log(prevState.foo);
            }
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 7, Column: 15},
			},
		},

		// ---- Upstream: state parameter in non-lifecycle method ----
		{
			Code: `
        class UseStateParameterOfNonLifecycleTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              foo: 123,
            };
          }
          nonLifecycle(someProps, someState) {
            doStuff(someState.foo)
          }
          render() {
            return (
              <SomeComponent />
            );
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 6, Column: 15},
			},
		},

		// ---- Upstream: missing state parameter ----
		{
			Code: `
        class MissingStateParameterTest extends Component {
          constructor(props) {
            super(props);
            this.state = {
              id: 123
            };
          }
          componentDidUpdate(someProps) {
            const prevState = { id: 456 };
            console.log(prevState.id);
          }
          render() {
            return (
              <h1>{this.state.selected ? 'Selected' : 'Not selected'}</h1>
            );
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'id'", Line: 6, Column: 15},
			},
		},

		// ---- Upstream: setState with unused initial ----
		{
			Code: `
        class Foo extends Component {
          state = {
            initial: 'foo',
          }
          handleChange = () => {
            this.setState(() => ({
              current: 'hi'
            }));
          }
          render() {
            const { current } = this.state;
            return <div>{current}</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'initial'", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: wrapped class expression ----
		{
			Code: `
        wrap(class NotWorking extends React.Component {
            state = {
                dummy: null
            };
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'dummy'", Line: 4, Column: 17},
			},
		},

		// ---- Upstream: TS unused state fields ----
		{
			Code: `
        class Foo extends Component {
          state = {
            thisStateAliasPropUnused,
            thisStateAliasRestPropUnused,
            thisDestructStateAliasPropUnused,
            thisDestructStateAliasRestPropUnused,
            thisDestructStateDestructRestPropUnused,
            thisSetStatePropUnused,
            thisSetStateRestPropUnused,
          } as unknown

          constructor() {
            ((this as unknown).state as unknown) = { thisStatePropUnused } as unknown;
            ((this as unknown).setState as unknown)({ thisStateDestructPropUnused } as unknown);
            ((this as unknown).setState as unknown)(state => ({ thisDestructStateDestructPropUnused } as unknown));
          }

          thisStateAlias() {
            const state = (this as unknown).state as unknown;

            (state as unknown).thisStateAliasProp as unknown;
            const { ...thisStateAliasRest } = state as unknown;
            (thisStateAliasRest as unknown).thisStateAliasRestProp as unknown;
          }

          thisDestructStateAlias() {
            const { state } = this as unknown;

            (state as unknown).thisDestructStateAliasProp as unknown;
            const { ...thisDestructStateAliasRest } = state as unknown;
            (thisDestructStateAliasRest as unknown).thisDestructStateAliasRestProp as unknown;
          }

          thisSetState() {
            ((this as unknown).setState as unknown)(state => (state as unknown).thisSetStateProp as unknown);
            ((this as unknown).setState as unknown)(({ ...thisSetStateRest }) => (thisSetStateRest as unknown).thisSetStateRestProp as unknown);
          }

          render() {
            ((this as unknown).state as unknown).thisStateProp as unknown;
            const { thisStateDestructProp } = (this as unknown).state as unknown;
            const { state: { thisDestructStateDestructProp, ...thisDestructStateDestructRest } } = this as unknown;
            (thisDestructStateDestructRest as unknown).thisDestructStateDestructRestProp as unknown;

            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisStateAliasPropUnused'", Line: 4, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisStateAliasRestPropUnused'", Line: 5, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisDestructStateAliasPropUnused'", Line: 6, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisDestructStateAliasRestPropUnused'", Line: 7, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisDestructStateDestructRestPropUnused'", Line: 8, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisSetStatePropUnused'", Line: 9, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisSetStateRestPropUnused'", Line: 10, Column: 13},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisStatePropUnused'", Line: 14, Column: 54},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisStateDestructPropUnused'", Line: 15, Column: 55},
				{MessageId: "unusedStateField", Message: "Unused state field: 'thisDestructStateDestructPropUnused'", Line: 16, Column: 65},
			},
		},

		// ---- Upstream: unused computed float literal key ----
		{
			Code: `
        class UnusedComputedFloatLiteralKeyStateTest extends React.Component {
          state = { [123.12]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: '123.12'", Line: 3, Column: 21},
			},
		},

		// ---- Edge: nested class — inner unused state is independent ----
		{
			Code: `
        class Outer extends React.Component {
          state = { outerFoo: 0 };
          render() {
            return <div>{this.state.outerFoo}</div>;
          }
        }
        class Inner extends React.Component {
          state = { innerBar: 0 };
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'innerBar'", Line: 9, Column: 21},
			},
		},

		// ---- Edge: parenthesized this.state assignment in constructor ----
		{
			Code: `
        class Hello extends React.Component {
          constructor() {
            (this).state = { foo: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 4, Column: 30},
			},
		},

		// ---- Edge: method shorthand in state object is tracked ----
		{
			Code: `
        class Hello extends React.Component {
          state = { handler() { return 1; } };
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'handler'", Line: 3, Column: 21},
			},
		},

		// ---- Edge: setState arrow-defined field unused ----
		{
			Code: `
        class Hello extends Component {
          handleChange = () => {
            this.setState(state => ({
              derived: state.count + 1
            }));
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'derived'", Line: 5, Column: 15},
			},
		},

		// ---- Edge: ES5 partially used state ----
		{
			Code: `
        var Hello = createReactClass({
          getInitialState: function() {
            return { used: 0, unused: 1 };
          },
          render: function() {
            return <SomeComponent foo={this.state.used} />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'unused'", Line: 4, Column: 31},
			},
		},

		// ---- Edge: multiple components, only one with unused state ----
		{
			Code: `
        class CompA extends React.Component {
          state = { a: 1 };
          render() { return <div>{this.state.a}</div>; }
        }
        class CompB extends React.Component {
          state = { b: 2 };
          render() { return <div />; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'b'", Line: 7, Column: 21},
			},
		},

		// ---- Edge: alias does not leak across methods ----
		{
			Code: `
        class Hello extends React.Component {
          state = { foo: 0 };
          methodA() {
            const s = this.state;
          }
          methodB() {
            const out = s.foo;
          }
          render() {
            return <div />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedStateField", Message: "Unused state field: 'foo'", Line: 3, Column: 21},
			},
		},
	})
}
