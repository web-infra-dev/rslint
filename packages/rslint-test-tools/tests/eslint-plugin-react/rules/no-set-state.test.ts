import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-set-state', {} as never, {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        var Hello = function() {
          this.setState({})
        };
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
    },
    // ---- Edge: plain class without React extends ----
    {
      code: `
        class Hello {
          someMethod() {
            this.setState({});
          }
        }
      `,
    },
    // ---- Edge: bracket access `this['setState']` is not matched ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['setState']({});
          }
        }
      `,
    },
    // ---- Edge: super.setState — receiver is SuperKeyword ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            super.setState({});
          }
        }
      `,
    },
    // ---- Edge: TS as-expression breaks the receiver match ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this as any).setState({});
          }
        }
      `,
    },
    // ---- Edge: lowercase function — not a stateless component ----
    {
      code: `
        function hello() {
          this.setState({});
          return <div/>;
        }
      `,
    },
    // ---- Edge: capital fn not returning JSX — not a component ----
    {
      code: `
        function Hello() {
          this.setState({});
          return 42;
        }
      `,
    },
    // ---- Edge: JSX onClick reference (not call) ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <button onClick={this.setState}/>;
          }
        }
      `,
    },
    // ---- Edge: this.setState.bind(this, {}) — outer call is on .bind ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <button onClick={this.setState.bind(this, {})}/>;
          }
        }
      `,
    },
    // ---- Edge: multi-level receiver this.x.setState ----
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            this.x.setState({});
          }
        }
      `,
    },
    // ---- Edge: extends arbitrary base, not React ----
    {
      code: `
        class Hello extends MyOwnBase {
          someMethod() {
            this.setState({});
          }
        }
      `,
    },
    // ---- Edge: custom pragma — React.Component is not the pragma ----
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({});
          }
        }
      `,
      settings: { react: { pragma: 'Preact' } },
    },
    // ---- Edge: settings.react is not a map (defensive) ----
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            other.setState({});
          }
        }
      `,
      settings: { react: true },
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          },
          render: function() {
            return <div>Hello {this.state.name}</div>;
          }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        var Hello = createReactClass({
          someMethod: function() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          },
          render: function() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          someMethod() {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        };
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          someMethod = () => {
            this.setState({
              name: this.props.name.toUpperCase()
            });
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello {this.state.name}</div>;
          }
        };
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div onMouseEnter={() => this.setState({dropdownIndex: index})} />;
          }
        };
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: parenthesized receiver `(this).setState` ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            (this).setState({});
            return <div/>;
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: optional-chain receiver `this?.setState` ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            this?.setState({});
            return <div/>;
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: setState inside setTimeout callback ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            setTimeout(() => {
              this.setState({});
            }, 100);
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: setState inside constructor ----
    {
      code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props);
            this.setState({});
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: extends PureComponent ----
    {
      code: `
        class Hello extends React.PureComponent {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: bare extends Component ----
    {
      code: `
        class Hello extends Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: stateless functional component ----
    {
      code: `
        function Hello() {
          this.setState({});
          return <div/>;
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: stateless arrow component ----
    {
      code: `
        const Hello = () => {
          this.setState({});
          return <div/>;
        };
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: multiple setState calls in same component ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            this.setState({});
            if (x) { this.setState({}); }
            return <div/>;
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }, { messageId: 'noSetState' }],
    },
    // ---- Edge: bare PureComponent ----
    {
      code: `
        class Hello extends PureComponent {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: custom pragma — Preact.Component IS a component ----
    {
      code: `
        class Hello extends Preact.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
      settings: { react: { pragma: 'Preact' } },
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Edge: custom createClass — myCreate IS a component ----
    {
      code: `
        var Hello = myCreate({
          someMethod: function() {
            this.setState({});
          },
          render: function() { return <div/>; }
        });
      `,
      settings: { react: { createClass: 'myCreate' } },
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Real-world: React.memo wrapping class ----
    {
      code: `
        const Hello = React.memo(class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Real-world: redux connect(...)(class) ----
    {
      code: `
        const Hello = connect(mapState)(class extends React.Component {
          someMethod() {
            this.setState({});
          }
          render() { return <div/>; }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Real-world: addEventListener callback ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            window.addEventListener("resize", () => {
              this.setState({});
            });
          }
          render() { return <div/>; }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- Multi-component file: only React component reports ----
    {
      code: `
        class A {
          foo() {
            this.setState({});
          }
        }
        class B extends React.Component {
          foo() {
            this.setState({});
          }
          render() { return <div/>; }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
  ],
});
