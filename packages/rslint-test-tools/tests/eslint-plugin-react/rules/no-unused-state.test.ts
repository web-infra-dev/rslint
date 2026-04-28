import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-state', {} as never, {
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
      errors: [{ message: "Unused state field: 'foo'" }],
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
      errors: [{ message: "Unused state field: 'foo'" }],
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
      errors: [{ message: "Unused state field: 'foo'" }],
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
      errors: [{ message: "Unused state field: 'foo'" }],
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
        { message: "Unused state field: 'foo'" },
        { message: "Unused state field: 'bar'" },
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
      errors: [{ message: "Unused state field: 'dummy'" }],
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
      errors: [{ message: "Unused state field: 'initial'" }],
    },
  ],
});
