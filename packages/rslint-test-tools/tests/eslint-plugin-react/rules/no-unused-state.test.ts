import { RuleTester } from '@rslint/rule-tester';
import {
  fixtureConfig,
  upstreamScriptOptions,
  upstreamModuleOptions,
} from '../config';
import { diagnostic } from '../expectations';

function expectUnusedStateFieldError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('unusedStateField', "Unused state field: 'foo'", range);
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-unused-state', {
  valid: [
    // ---- Stateless function component ----
    {
      code: `
        function StatelessFnUnaffectedTest(props) {
          return <SomeComponent foo={props.foo} />;
        };
      `,
    },
    // ---- createReactClass without state ----
    {
      code: `
        var NoStateTest = createReactClass({
          render: function() {
            return <SomeComponent />;
          }
        });
      `,
    },
    // ---- getInitialState with used state ----
    {
      code: `
        var GetInitialStateTest = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `,
    },
    // ---- Computed key from variable ----
    {
      code: `
        var ComputedKeyFromVariableTest = createReactClass({
          getInitialState: function() {
            return { [foo]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `,
    },
    // ---- Class component without state ----
    {
      code: `
        class NoStateTest extends React.Component {
          render() {
            return <SomeComponent />;
          }
        }
      `,
    },
    // ---- Constructor state with usage ----
    {
      code: `
        class CtorStateTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
    },
    // ---- Class property state ----
    {
      code: `
        class ClassPropertyStateTest extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
    },
    // ---- Optional chaining ----
    {
      code: `
        class OptionalChaining extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent foo={this.state?.foo} />;
          }
        }
      `,
    },
    // ---- Destructuring ----
    {
      code: `
        class DestructuringTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {foo: myFoo} = this.state;
            return <SomeComponent foo={myFoo} />;
          }
        }
      `,
    },
    // ---- Alias declaration ----
    {
      code: `
        class AliasDeclarationTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const state = this.state;
            return <SomeComponent foo={state.foo} />;
          }
        }
      `,
    },
    // ---- Shorthand destructuring alias ----
    {
      code: `
        class ShorthandDestructuringAliasTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const {state} = this;
            return <SomeComponent foo={state.foo} />;
          }
        }
      `,
    },
    // ---- Rest property ----
    {
      code: `
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
      `,
    },
    // ---- JSX spread (gives up) ----
    {
      code: `
        class JsxSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent {...this.state} />;
          }
        }
      `,
    },
    // ---- Object spread (gives up) ----
    {
      code: `
        class ObjectSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const attrs = { ...this.state, foo: 1 };
            return <SomeComponent foo={attrs.foo} />;
          }
        }
      `,
    },
    // ---- getDerivedStateFromProps ----
    {
      code: `
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
      `,
    },
    // ---- shouldComponentUpdate ----
    {
      code: `
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
      `,
    },
    // ---- setState callback with state param ----
    {
      code: `
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
      `,
    },
    // ---- getDerivedStateFromProps as class property ----
    {
      code: `
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
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var NoStateMethodTest = createReactClass({
          render() {
            return <SomeComponent />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var ComputedKeyFromBooleanLiteralTest = createReactClass({
          getInitialState: function() {
            return { [true]: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state[true]} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        var ComputedKeyFromNumberLiteralTest = createReactClass({
          getInitialState: function() {
            return { [123]: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state[123]} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        var ComputedKeyFromExpressionTest = createReactClass({
          getInitialState: function() {
            return { [foo + bar]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        var ComputedKeyFromBinaryExpressionTest = createReactClass({
          getInitialState: function() {
            return { ['foo' + 'bar' * 8]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        var ComputedKeyFromStringLiteralTest = createReactClass({
          getInitialState: function() {
            return { ['foo']: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        var ComputedKeyFromTemplateLiteralTest = createReactClass({
          getInitialState: function() {
            return { [\`foo\${bar}\`]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        var ComputedKeyFromTemplateLiteralTest = createReactClass({
          getInitialState: function() {
            return { [\`foo\`]: 0 };
          },
          render: function() {
            return <SomeComponent foo={this.state['foo']} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        var GetInitialStateMethodTest = createReactClass({
          getInitialState() {
            return { foo: 0 };
          },
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        var SetStateTest = createReactClass({
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          },
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        });
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class ComputedKeyFromVariableTest extends React.Component {
          constructor() {
            this.state = { [foo]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class ComputedKeyFromBooleanLiteralTest extends React.Component {
          constructor() {
            this.state = { [false]: 0 };
          }
          render() {
            return <SomeComponent foo={this.state['false']} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class ComputedKeyFromNumberLiteralTest extends React.Component {
          constructor() {
            this.state = { [345]: 0 };
          }
          render() {
            return <SomeComponent foo={this.state[345]} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class ComputedKeyFromExpressionTest extends React.Component {
          constructor() {
            this.state = { [foo + bar]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class ComputedKeyFromBinaryExpressionTest extends React.Component {
          constructor() {
            this.state = { [1 + 2 * 8]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class ComputedKeyFromStringLiteralTest extends React.Component {
          constructor() {
            this.state = { ['foo']: 0 };
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class ComputedKeyFromTemplateLiteralTest extends React.Component {
          constructor() {
            this.state = { [\`foo\${bar}\`]: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class ComputedKeyFromTemplateLiteralTest extends React.Component {
          constructor() {
            this.state = { [\`foo\`]: 0 };
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class SetStateTest extends React.Component {
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class VariableDeclarationTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            const foo = this.state.foo;
            return <SomeComponent foo={foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
      class DestructuringTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const {foo: myFoo} = this.state;
          return <SomeComponent foo={myFoo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
      class ShorthandDestructuringTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const {foo} = this.state;
          return <SomeComponent foo={foo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
      class AliasDeclarationTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const state = this.state;
          return <SomeComponent foo={state.foo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
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
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
      class DestructuringAliasTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const {state: myState} = this;
          return <SomeComponent foo={myState.foo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
      class ShorthandDestructuringAliasTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const {state} = this;
          return <SomeComponent foo={state.foo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
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
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class DeepDestructuringTest extends React.Component {
          state = { foo: 0, bar: 0 };
          render() {
            const {state: {foo, ...others}} = this;
            return <SomeComponent foo={foo} bar={others.bar} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
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
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
      class AssignedToObjectFalseNegativeTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const obj = { foo: this.state.foo, bar: 0 };
          return <SomeComponent bar={obj.bar} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
      class ComputedAccessFalseNegativeTest extends React.Component {
        constructor() {
          this.state = { foo: 0, bar: 1 };
        }
        render() {
          const bar = 'bar';
          return <SomeComponent bar={this.state[bar]} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
      class JsxSpreadFalseNegativeTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          return <SomeComponent {...this.state} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
      class AliasedJsxSpreadFalseNegativeTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const state = this.state;
          return <SomeComponent {...state} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
      class ObjectSpreadFalseNegativeTest extends React.Component {
        constructor() {
          this.state = { foo: 0 };
        }
        render() {
          const attrs = { ...this.state, foo: 1 };
          return <SomeComponent foo={attrs.foo} />;
        }
      }
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
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
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
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
    `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (')' expected.).",
      code: `
        class TypeCastExpressionSpreadFalseNegativeTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent {...(this.state: any)} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        class ThisStateAsAnObject extends React.Component {
          state = {
            active: true
          };

          render() {
            return <div className={classNames('overflowEdgeIndicator', className, this.state)} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `
        class Foo extends Component {
          handleChange = function() {
            this.setState(() => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `
        class Foo extends Component {
          handleChange = function() {
            this.setState(state => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
        class Foo extends Component {
          static handleChange = () => {
            this.setState(state => ({ foo: value }));
          }
          render() {
            return <SomeComponent foo={this.state.foo} />;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
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
            // other methods of defining state props
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        interface Props {}

        interface State {
          flag: boolean;
        }

        export default class RuleTest extends React.Component<Props, State> {
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        class KarmaRefundPillComponent extends GenericPillComponent {
          renderContent = () => {
            const { action } = this.props

            return (
              <Box fontSize={[1]} mx={[2]} minWidth="10px" minHeight="26px" alignItems="center">
                <FormattedText
                  fields={getKarmaClaimLevel1Fields(action)}
                  i18nKey="pillTemplates.karmarefund.summary"
                  fontSize={[1]}
                />
              </Box>
            )
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
        class AutoControlledComponent<P = {}, S = {}> extends UIComponent<P, S> {
          static getDerivedStateFromProps: React.GetDerivedStateFromProps<any, any>
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
        export const commonMixinWrapper = ComposeComponent => class extends ComposeComponent {
          static getDerivedStateFromProps = ComposeComponent.getDerivedStateFromProps;
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
      code: `
        import React, { PureComponent } from 'react';

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

        export default TestNoUnusedState;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        class Component extends React.Component {
          static getDerivedStateFromProps = ({value, disableAnimation}: ToggleProps, {isControlled, isOn}: ToggleState) => {
            return { isControlled, isOn };
          };

          render() {
            const { isControlled, isOn } = this.state;
            return <div>{isControlled ? 'controlled' : ''}{isOn ? 'on' : ''}</div>;
          }
        }
      `,
      languageOptions: upstreamScriptOptions,
    },
  ],
  invalid: [
    // ---- Unused getInitialState ----
    {
      code: `
        var UnusedGetInitialStateTest = createReactClass({
          getInitialState: function() {
            return { foo: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [expectUnusedStateFieldError([4, 22, 4, 28])],
    },
    // ---- Unused class constructor state ----
    {
      code: `
        class UnusedCtorStateTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([4, 28, 4, 34])],
    },
    // ---- Unused class property state ----
    {
      code: `
        class UnusedClassPropertyStateTest extends React.Component {
          state = { foo: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([3, 21, 3, 27])],
    },
    // ---- Unused setState ----
    {
      code: `
        class UnusedSetStateTest extends React.Component {
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([4, 29, 4, 40])],
    },
    // ---- Multiple unused fields ----
    {
      code: `
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
      errors: [
        expectUnusedStateFieldError([5, 15, 5, 21]),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'bar'",
          [6, 15, 6, 21],
        ),
      ],
    },
    // ---- Wrapped class expression ----
    {
      code: `
        wrap(class NotWorking extends React.Component {
            state = {
                dummy: null
            };
        });
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'dummy'",
          [4, 17, 4, 28],
        ),
      ],
    },
    // ---- setState with unused initial ----
    {
      code: `
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
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'initial'",
          [4, 13, 4, 27],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var UnusedComputedStringLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { ['foo']: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [expectUnusedStateFieldError([4, 22, 4, 32])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var UnusedComputedTemplateLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { [\`foo\`]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [expectUnusedStateFieldError([4, 22, 4, 32])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        var UnusedComputedNumberLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { [123]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: '123'",
          [4, 22, 4, 30],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var UnusedComputedBooleanLiteralKeyStateTest = createReactClass({
          getInitialState: function() {
            return { [true]: 0 };
          },
          render: function() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'true'",
          [4, 22, 4, 31],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var UnusedGetInitialStateMethodTest = createReactClass({
          getInitialState() {
            return { foo: 0 };
          },
          render() {
            return <SomeComponent />;
          }
        })
      `,
      errors: [expectUnusedStateFieldError([4, 22, 4, 28])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var UnusedSetStateTest = createReactClass({
          onFooChange(newFoo) {
            this.setState({ foo: newFoo });
          },
          render() {
            return <SomeComponent />;
          }
        });
      `,
      errors: [expectUnusedStateFieldError([4, 29, 4, 40])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class UnusedComputedStringLiteralKeyStateTest extends React.Component {
          state = { ['foo']: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([3, 21, 3, 31])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class UnusedComputedTemplateLiteralKeyStateTest extends React.Component {
          state = { [\`foo\`]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([3, 21, 3, 31])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class UnusedComputedTemplateLiteralKeyStateTest extends React.Component {
          state = { [\`foo \\n bar\`]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'foo \n bar'",
          [3, 21, 3, 38],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        class UnusedComputedBooleanLiteralKeyStateTest extends React.Component {
          state = { [true]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'true'",
          [3, 21, 3, 30],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        class UnusedComputedNumberLiteralKeyStateTest extends React.Component {
          state = { [123]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: '123'",
          [3, 21, 3, 29],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class UnusedComputedFloatLiteralKeyStateTest extends React.Component {
          state = { [123.12]: 0 };
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: '123.12'",
          [3, 21, 3, 32],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class UnusedStateWhenPropsAreSpreadTest extends React.Component {
          constructor() {
            this.state = { foo: 0 };
          }
          render() {
            return <SomeComponent {...this.props} />;
          }
        }
      `,
      errors: [expectUnusedStateFieldError([4, 28, 4, 34])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
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
      errors: [expectUnusedStateFieldError([4, 28, 4, 34])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
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
      errors: [
        expectUnusedStateFieldError([4, 28, 4, 34]),
        expectUnusedStateFieldError([7, 29, 7, 40]),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
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
      errors: [expectUnusedStateFieldError([5, 15, 5, 21])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
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
      errors: [expectUnusedStateFieldError([4, 28, 4, 34])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax ('=>' expected.).",
      code: `
        class TypeCastExpressionTest extends React.Component {
          constructor() {
            this.state = {
              foo: 0,
              bar: 1,
              baz: 2,
              qux: 3,
            };
          }
          render() {
            const foo = ((this: any).state: any).foo;
            const {bar, ...others} = (this.state: any);
            let baz;
            baz = (others: any)['baz'];
            return <SomeComponent foo={foo} bar={bar} baz={baz} />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'qux'",
          [8, 15, 8, 21],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class UnusedDeepDestructuringTest extends React.Component {
          state = { foo: 0, bar: 0 };
          render() {
            const {state: {foo}} = this;
            return <SomeComponent foo={foo} />;
          }
        }
      `,
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'bar'",
          [3, 29, 3, 35],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
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
      errors: [expectUnusedStateFieldError([7, 15, 7, 23])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
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
      errors: [expectUnusedStateFieldError([6, 15, 6, 23])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
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
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'id'",
          [6, 15, 6, 22],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
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
            // other methods of defining state props
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
      errors: [
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisStateAliasPropUnused'",
          [4, 13, 4, 37],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisStateAliasRestPropUnused'",
          [5, 13, 5, 41],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisDestructStateAliasPropUnused'",
          [6, 13, 6, 45],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisDestructStateAliasRestPropUnused'",
          [7, 13, 7, 49],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisDestructStateDestructRestPropUnused'",
          [8, 13, 8, 52],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisSetStatePropUnused'",
          [9, 13, 9, 35],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisSetStateRestPropUnused'",
          [10, 13, 10, 39],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisStatePropUnused'",
          [15, 54, 15, 73],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisStateDestructPropUnused'",
          [16, 55, 16, 82],
        ),
        diagnostic(
          'unusedStateField',
          "Unused state field: 'thisDestructStateDestructPropUnused'",
          [17, 65, 17, 100],
        ),
      ],
      languageOptions: upstreamScriptOptions,
    },
  ],
});
