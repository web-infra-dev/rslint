import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoSetStateError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSetState',
    message: 'Do not use setState',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-set-state', {
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div onMouseEnter={() => this.setState({dropdownIndex: index})} />;
          }
        };
      `,
      errors: [expectNoSetStateError([4, 45, 4, 58])],
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
      errors: [expectNoSetStateError([4, 13, 4, 28])],
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
      errors: [expectNoSetStateError([4, 13, 4, 27])],
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
      errors: [expectNoSetStateError([5, 15, 5, 28])],
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
      errors: [expectNoSetStateError([5, 13, 5, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
    },
    // ---- Edge: stateless functional component ----
    {
      code: `
        function Hello() {
          this.setState({});
          return <div/>;
        }
      `,
      errors: [expectNoSetStateError([3, 11, 3, 24])],
    },
    // ---- Edge: stateless arrow component ----
    {
      code: `
        const Hello = () => {
          this.setState({});
          return <div/>;
        };
      `,
      errors: [expectNoSetStateError([3, 11, 3, 24])],
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
      errors: [
        expectNoSetStateError([4, 13, 4, 26]),
        expectNoSetStateError([5, 22, 5, 35]),
      ],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError([5, 15, 5, 28])],
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
      errors: [expectNoSetStateError([9, 13, 9, 26])],
    },
  ],
});
