import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-stateless-function', {} as never, {
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
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
      errors: [{ messageId: 'componentShouldBePure' }],
    },
  ],
});
