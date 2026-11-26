import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectComponentShouldBePureError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'componentShouldBePure',
    message: 'Component should be written as a pure function',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('prefer-stateless-function', {
  valid: [
    // ---- Already a stateless function ----
    {
      code: `
        const Foo = function(props) {
          return <div>{props.foo}</div>;
        };
      `,
    },
    // ---- Already a stateless arrow function ----
    {
      code: `const Foo = ({foo}) => <div>{foo}</div>;`,
    },
    // ---- PureComponent + props + ignorePureComponents ----
    {
      code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      options: [{ ignorePureComponents: true }],
    },
    // ---- PureComponent + context + ignorePureComponents ----
    {
      code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.context.foo}</div>;
          }
        }
      `,
      options: [{ ignorePureComponents: true }],
    },
    // ---- PureComponent in expression context + ignorePureComponents ----
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      options: [{ ignorePureComponents: true }],
    },
    // ---- Has lifecycle method ----
    {
      code: `
        class Foo extends React.Component {
          shouldComponentUpdate() {
            return false;
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- Has state ----
    {
      code: `
        class Foo extends React.Component {
          changeState() {
            this.setState({foo: "clicked"});
          }
          render() {
            return <div onClick={this.changeState.bind(this)}>{this.state.foo || "bar"}</div>;
          }
        }
      `,
    },
    // ---- Uses this.refs ----
    {
      code: `
        class Foo extends React.Component {
          doStuff() {
            this.refs.foo.style.backgroundColor = "red";
          }
          render() {
            return <div ref="foo" onClick={this.doStuff}>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- Has additional method ----
    {
      code: `
        class Foo extends React.Component {
          doStuff() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- Empty (no super) constructor ----
    {
      code: `
        class Foo extends React.Component {
          constructor() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- Constructor with non-super body ----
    {
      code: `
        class Foo extends React.Component {
          constructor() {
            doSpecialStuffs();
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- this.bar — useThis ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.bar}</div>;
          }
        }
      `,
    },
    // ---- destructure this.bar — useThis ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            let {props:{foo}, bar} = this;
            return <div>{foo}</div>;
          }
        }
      `,
    },
    // ---- this[bar] / this['bar'] — useThis ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this[bar]}</div>;
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this['bar']}</div>;
          }
        }
      `,
    },
    // ---- nested ClassExpression ----
    {
      code: `
        export default (Component) => (
          class Test extends React.Component {
            componentDidMount() {}
            render() {
              return <Component />;
            }
          }
        );
      `,
    },
    // ---- external Foo.childContextTypes = ... ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.children}</div>;
          }
        }
        Foo.childContextTypes = {
          color: PropTypes.string
        };
      `,
    },
    // ---- decorator on class ----
    {
      code: `
        @foo
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- multiple decorators ----
    {
      code: `
        @foo
        @bar()
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
    },
    // ---- bare PureComponent + ignorePureComponents ----
    {
      code: `
        class Child extends PureComponent {
          render() {
            return <h1>I don't</h1>;
          }
        }
      `,
      options: [{ ignorePureComponents: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        class Foo extends React.Component {
          constructor() {
            foo;
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Foo extends React.Component {
          constructor(props)

          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class Foo extends React.Component {
          render() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      settings: { react: { version: '0.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        var Foo = createReactClass({
          render: function() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        });
      `,
      settings: { react: { version: '0.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Foo extends React.Component {
          render() {
            return true ? <div /> : null;
          }
        }
      `,
      settings: { react: { version: '0.14.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        @foo("bar")
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        import React, {PureComponent, PropTypes} from 'react'

        export default function errorDecorator (options) {
          return WrappedComponent => {
            class Wrapper extends PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
            return Wrapper
          }
        }
      `,
      options: [{ ignorePureComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        import React, {PureComponent, PropTypes} from 'react'

        export default function errorDecorator (options) {
          return WrappedComponent =>
            class Wrapper extends PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
        }
      `,
      options: [{ ignorePureComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        export default function errorDecorator (options) {
          return WrappedComponent =>
            class Wrapper extends React.PureComponent {
              static propTypes = {
                error: PropTypes.string
              }
              render () {
                const {error, ...props} = this.props
                if (error) {
                  return <div>Error! {error}</div>
                } else {
                  return <WrappedComponent {...props} />
                }
              }
            }
        }
      `,
      options: [{ ignorePureComponents: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        /**
         * @param a.
         */
        function Comp() {
          return <a></a>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- only uses this.props ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- this['props'] ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>{this['props'].foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- PureComponent without ignorePureComponents ----
    {
      code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>foo</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- PureComponent + props (default ignorePureComponents=false) ----
    {
      code: `
        class Foo extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- static get displayName ----
    {
      code: `
        class Foo extends React.Component {
          static get displayName() {
            return 'Foo';
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- static displayName field ----
    {
      code: `
        class Foo extends React.Component {
          static displayName = 'Foo';
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 7, 10])],
    },
    // ---- static get propTypes ----
    {
      code: `
        class Foo extends React.Component {
          static get propTypes() {
            return {
              name: PropTypes.string
            };
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 11, 10])],
    },
    // ---- static propTypes field ----
    {
      code: `
        class Foo extends React.Component {
          static propTypes = {
            name: PropTypes.string
          };
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- props with type annotation ----
    {
      code: `
        class Foo extends React.Component {
          props: {
            name: string;
          };
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- useless constructor ----
    {
      code: `
        class Foo extends React.Component {
          constructor() {
            super();
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- destructures only props/context ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            let {props:{foo}, context:{bar}} = this;
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 7, 10])],
    },
    // ---- render returns null on default (>= 15) ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- createReactClass with null return ----
    {
      code: `
        var Foo = createReactClass({
          render: function() {
            if (!this.props.foo) {
              return null;
            }
            return <div>{this.props.foo}</div>;
          }
        });
      `,
      errors: [expectComponentShouldBePureError([2, 36, 9, 10])],
    },
    // ---- shorthand-if returning null at default (>= 15) ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            return true ? <div /> : null;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- defaultProps as class field ----
    {
      code: `
        class Foo extends React.Component {
          static defaultProps = {
            foo: true
          }
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 10, 10])],
    },
    // ---- external Foo.defaultProps ----
    {
      code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
        Foo.defaultProps = {
          foo: true
        };
      `,
      errors: [expectComponentShouldBePureError([2, 9, 7, 10])],
    },
    // ---- contextTypes static field ----
    {
      code: `
        class Foo extends React.Component {
          static contextTypes = {
            foo: PropTypes.boolean
          }
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 10, 10])],
    },
    // ---- bare Component (not React.Component) ----
    {
      code: `
        class Foo extends Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },
    // ---- ClassExpression assigned to var ----
    {
      code: `
        var Foo = class extends React.Component {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      errors: [expectComponentShouldBePureError([2, 19, 6, 10])],
    },
    // ---- useless constructor with super(...arguments) ----
    {
      code: `
        class Foo extends React.Component {
          constructor() {
            super(...arguments);
          }
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 9, 10])],
    },
    // ---- bare PureComponent without ignorePureComponents ----
    {
      code: `
        class Foo extends PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 6, 10])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static get defaultProps() {
            return {
              foo: true
            };
          }
          render() {
            const { foo } = this.props;
            return foo ? <div /> : null;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 12, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        class Foo extends React.Component {
          static get contextTypes() {
            return {
              foo: PropTypes.boolean
            };
          }
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
      `,
      errors: [expectComponentShouldBePureError([2, 9, 12, 10])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.context;
            return foo ? <div /> : null;
          }
        }
        Foo.contextTypes = {
          foo: PropTypes.boolean
        };
      `,
      errors: [expectComponentShouldBePureError([2, 9, 7, 10])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
