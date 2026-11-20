import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectUnsortedPropsError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'unsortedProps',
    'render should be placed after displayName',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('sort-comp', {
  valid: [
    {
      code: `class Hello extends React.Component {
        static displayName = 'Hello';
        componentDidMount() {}
        render() {}
      }`,
    },
    {
      code: `var Hello = createReactClass({
        displayName: 'Hello',
        onClick() {},
        render() {},
      });`,
    },
    {
      code: `class Hello extends React.Component {
        onClick() {}
        render() {}
      }`,
      options: [{ order: ['lifecycle', 'everything-else', 'render'] }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        // Must validate a class with missing groups
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        // Must put a custom method in 'everything-else'
        var Hello = createReactClass({
          onClick: function() {},
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        // Must allow us to re-order the groups
        var Hello = createReactClass({
          displayName : 'Hello',
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          },
          onClick: function() {}
        });
      `,
      options: [{ order: ['lifecycle', 'render', 'everything-else'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
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
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        // Must allow us to create a RegExp-based group
        class Hello extends React.Component {
          customHandler() {}
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `,
      options: [
        { order: ['lifecycle', 'everything-else', 'render', '/on.*/'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        // Must allow us to create a named group
        class Hello extends React.Component {
          customHandler() {}
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `,
      options: [
        {
          order: ['lifecycle', 'everything-else', 'render', 'customGroup'],
          groups: { customGroup: ['/on.*/'] },
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        // Must allow a method to be in different places if it's matches multiple patterns
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
          onClick() {}
        }
      `,
      options: [{ order: ['/on.*/', 'render', '/.*Click/'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        // Must allow us to use 'constructor' as a method name
        class Hello extends React.Component {
          constructor() {}
          displayName() {}
          render() {
            return <div>Hello</div>;
          }
        }
      `,
      options: [
        { order: ['constructor', 'lifecycle', 'everything-else', 'render'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        // Must ignore stateless components
        function Hello(props) {
          return <div>Hello {props.name}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        // Must ignore stateless components (arrow function with explicit return)
        var Hello = props => (
          <div>Hello {props.name}</div>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        // Must ignore spread operator
        var Hello = createReactClass({
          ...proto,
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        // Type Annotations should be first
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'type-annotations',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        // Properties with Type Annotations should not be at the top
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          state: Object = {};
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'type-annotations',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        // Non-react classes should be ignored, even in expressions
        return class Hello {
          render() {
            return <div>{this.props.text}</div>;
          }
          props: { text: string };
          constructor() {}
          state: Object = {};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        // Non-react classes should be ignored, even in expressions
        return class {
          render() {
            return <div>{this.props.text}</div>;
          }
          props: { text: string };
          constructor() {}
          state: Object = {};
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        // Getters should be at the top
        class Hello extends React.Component {
          get foo() {}
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'getters',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        // Setters should be at the top
        class Hello extends React.Component {
          set foo(bar) {}
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'setters',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
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
      `,
      options: [
        {
          order: ['instance-methods', 'lifecycle', 'everything-else', 'render'],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
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
      `,
      options: [
        {
          order: [
            'instance-variables',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        // Methods can be grouped with any matching group (with statics)
        class Hello extends React.Component {
          static onFoo() {}
          static renderFoo() {}
          render() {
            return <div>{this.props.text}</div>;
          }
          getFoo() {}
        }
      `,
      options: [
        {
          order: [
            'static-methods',
            'render',
            '/^get.+$/',
            '/^on.+$/',
            '/^render.+$/',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        // Methods can be grouped with any matching group (with RegExp)
        class Hello extends React.Component {
          render() {
            return <div>{this.props.text}</div>;
          }
          getFoo() {}
          static onFoo() {}
          static renderFoo() {}
        }
      `,
      options: [
        {
          order: [
            'static-methods',
            'render',
            '/^get.+$/',
            '/^on.+$/',
            '/^render.+$/',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        // static lifecycle methods can be grouped (with statics)
        class Hello extends React.Component {
          static getDerivedStateFromProps() {}
          constructor() {}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        // static lifecycle methods can be grouped (with lifecycle)
        class Hello extends React.Component {
          constructor() {}
          static getDerivedStateFromProps() {}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          state = {};
          foo;
          static propTypes;

          render() {
            return null;
          }
        }
      `,
      options: [
        {
          order: [
            'state',
            'instance-variables',
            'static-methods',
            'lifecycle',
            'render',
            'everything-else',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static foo;
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `,
      options: [{ order: ['static-variables', 'static-methods'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static foo = 'some-str';

          render() {
            return null;
          }
        }
      `,
      options: [{ order: ['static-methods', 'static-variables'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          foo = {};
          static bar = 0;
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `,
      options: [
        { order: ['instance-variables', 'static-variables', 'static-methods'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static bar = 1;
          foo = {};
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `,
      options: [
        { order: ['static-variables', 'instance-variables', 'static-methods'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          render() {
            return null;
          }
          static bar;
          foo = {};
        }
      `,
      options: [
        {
          order: [
            'static-methods',
            'render',
            'static-variables',
            'instance-variables',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
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
      `,
      options: [
        {
          order: [
            'static-variables',
            'instance-variables',
            'constructor',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        class ClassName extends React.Component {
          static defaultProps = {};
          static parseDateString(date?: Date) {}
          state = {};
          render() {
            return <div />;
          }
        }
      `,
      options: [
        {
          order: [
            'static-variables',
            'static-methods',
            'type-annotations',
            'instance-variables',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `class Hello extends React.Component {
        render() {}
        componentDidMount() {}
      }`,
      errors: [
        diagnostic(
          'unsortedProps',
          'render should be placed after componentDidMount',
          [2, 9, 2, 20],
        ),
      ],
    },
    {
      code: `var Hello = createReactClass({
        render() {},
        displayName: 'Hello',
      });`,
      errors: [expectUnsortedPropsError([2, 9, 2, 20])],
    },
    {
      code: `class Hello extends React.Component {
        onClick() {}
        render() {}
      }`,
      options: [{ order: ['lifecycle', 'render'] }],
      errors: [
        diagnostic(
          'unsortedProps',
          'onClick should be placed after render',
          [2, 9, 2, 21],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        // Must force a lifecycle method to be placed before render
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          },
          displayName : 'Hello',
        });
      `,
      errors: [expectUnsortedPropsError([4, 11, 6, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        // Must run rule when render uses createElement instead of JSX
        var Hello = createReactClass({
          render: function() {
            return React.createElement("div", null, "Hello");
          },
          displayName : 'Hello',
        });
      `,
      errors: [expectUnsortedPropsError([4, 11, 6, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        // Must force a custom method to be placed before render
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          },
          onClick: function() {},
        });
      `,
      errors: [
        diagnostic(
          'unsortedProps',
          'render should be placed after onClick',
          [4, 11, 6, 12],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        // Must force a custom method to be placed before render, even in function
        var Hello = () => {
          return class Test extends React.Component {
            render () {
              return <div>Hello</div>;
            }
            onClick () {}
          }
        };
      `,
      errors: [
        diagnostic(
          'unsortedProps',
          'render should be placed after onClick',
          [5, 13, 7, 14],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        // Must force a custom method to be placed after render if no 'everything-else' group is specified
        var Hello = createReactClass({
          displayName: 'Hello',
          onClick: function() {},
          render: function() {
            return <button onClick={this.onClick}>Hello</button>;
          }
        });
      `,
      options: [{ order: ['lifecycle', 'render'] }],
      errors: [
        diagnostic(
          'unsortedProps',
          'onClick should be placed after render',
          [5, 11, 5, 33],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        // Must validate static properties
        class Hello extends React.Component {
          render() {
            return <div></div>
          }
          static displayName = 'Hello';
        }
      `,
      errors: [expectUnsortedPropsError([4, 11, 6, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        // Type Annotations should not be at the top by default
        class Hello extends React.Component {
          props: { text: string };
          constructor() {}
          state: Object = {};
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      errors: [
        diagnostic(
          'unsortedProps',
          'props should be placed after state',
          [4, 11, 4, 35],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        // Type Annotations should be first
        class Hello extends React.Component {
          constructor() {}
          props: { text: string };
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'type-annotations',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'constructor should be placed after props',
          [4, 11, 4, 27],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        // Properties with Type Annotations should not be at the top
        class Hello extends React.Component {
          props: { text: string };
          state: Object = {};
          constructor() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'type-annotations',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'state should be placed after constructor',
          [5, 11, 5, 30],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        // componentDidMountOk should be placed after getA
        export default class View extends React.Component {
          componentDidMountOk() {}
          getB() {}
          componentWillMount() {}
          getA() {}
          render() {}
        }
      `,
      options: [
        {
          order: [
            'static-methods',
            'lifecycle',
            '/^on.+$/',
            '/^(get|set)(?!(InitialState$|DefaultProps$|ChildContext$)).+$/',
            'everything-else',
            '/^render.+$/',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'componentDidMountOk should be placed after getA',
          [4, 11, 4, 35],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        // Getters should at the top
        class Hello extends React.Component {
          constructor() {}
          get foo() {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'getters',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'constructor should be placed after getter functions',
          [4, 11, 4, 27],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        // Setters should at the top
        class Hello extends React.Component {
          constructor() {}
          set foo(bar) {}
          render() {
            return <div>{this.props.text}</div>;
          }
        }
      `,
      options: [
        {
          order: [
            'setters',
            'static-methods',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'constructor should be placed after setter functions',
          [4, 11, 4, 27],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
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
      `,
      options: [
        {
          order: ['instance-methods', 'lifecycle', 'everything-else', 'render'],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'foo should be placed before constructor',
          [7, 11, 7, 30],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
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
      `,
      options: [
        {
          order: [
            'instance-variables',
            'lifecycle',
            'everything-else',
            'render',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'foo should be placed before constructor',
          [7, 11, 7, 19],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        // Should not confuse method names with group names
        class Hello extends React.Component {
          setters() {}
          constructor() {}
          render() {}
        }
      `,
      options: [{ order: ['setters', 'lifecycle', 'render'] }],
      errors: [
        diagnostic(
          'unsortedProps',
          'setters should be placed after render',
          [4, 11, 4, 23],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        // Explicitly named methods should appear in the correct order
        class Hello extends React.Component {
          render() {}
          foo() {}
        }
      `,
      options: [{ order: ['foo', 'render'] }],
      errors: [
        diagnostic(
          'unsortedProps',
          'render should be placed after foo',
          [4, 11, 4, 22],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static foo;

          render() {
            return null;
          }
        }
      `,
      options: [{ order: ['static-variables', 'static-methods'] }],
      errors: [
        diagnostic(
          'unsortedProps',
          'getDerivedStateFromProps should be placed after foo',
          [3, 11, 3, 47],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static foo;
          bar = 'some-str'
          static getDerivedStateFromProps() {}

          render() {
            return null;
          }
        }
      `,
      options: [
        { order: ['instance-variables', 'static-variables', 'static-methods'] },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'foo should be placed after bar',
          [3, 11, 3, 22],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          static getDerivedStateFromProps() {}
          static bar;
          render() {
            return null;
          }
          foo = {};
        }
      `,
      options: [
        {
          order: [
            'static-methods',
            'render',
            'static-variables',
            'instance-variables',
          ],
        },
      ],
      errors: [
        diagnostic(
          'unsortedProps',
          'bar should be placed after render',
          [4, 11, 4, 22],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
