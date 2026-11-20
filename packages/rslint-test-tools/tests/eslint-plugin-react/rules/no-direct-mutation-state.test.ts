import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoDirectMutationError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noDirectMutation',
    message: 'Do not mutate state directly. Use setState().',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-direct-mutation-state', {
  valid: [
    // ---- Upstream valid cases ----
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
          render: function() {
            var obj = {state: {}};
            obj.state.name = "foo";
            return <div>Hello {obj.state.name}</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = "foo";
        module.exports = {};
      `,
    },
    {
      code: `
        class Hello {
          getFoo() {
            this.state.foo = 'bar'
            return this.state.foo;
          }
        }
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          constructor() {
            this.state.foo = "bar"
          }
        }
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          constructor() {
            this.state.foo = 1;
          }
        }
      `,
    },
    {
      code: `
        class OneComponent extends Component {
          constructor() {
            super();
            class AnotherComponent extends Component {
              constructor() {
                super();
              }
            }
            this.state = {};
          }
        }
      `,
    },
    // ---- Edge: bracket access `this['state']` is not matched ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['state'].foo = 'bar';
          }
        }
      `,
    },
    // ---- Edge: lowercase function — not a stateless component ----
    {
      code: `
        function hello() {
          this.state.x = 1;
          return <div/>;
        }
      `,
    },
    // ---- Edge: Capital fn not returning JSX — not a component ----
    {
      code: `
        function Hello() {
          this.state.x = 1;
          return 42;
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            this.state.foo = "bar"
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            this.state.foo++;
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name= "bar"
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 30])],
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name.first = "bar"
            return <div>Hello</div>;
          }
        });
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 35])],
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name.first = "bar"
            this.state.person.name.last = "baz"
            return <div>Hello</div>;
          }
        });
      `,
      errors: [
        expectNoDirectMutationError([4, 13, 4, 35]),
        expectNoDirectMutationError([5, 13, 5, 35]),
      ],
    },
    {
      code: `
        class Hello extends React.Component {
          constructor() {
            someFn()
          }
          someFn() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([7, 13, 7, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props)
            doSomethingAsync(() => {
              this.state = "bad";
            });
          }
        }
      `,
      errors: [expectNoDirectMutationError([6, 15, 6, 19])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillMount() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillReceiveProps() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          shouldComponentUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillUnmount() {
            this.state.foo = "bar"
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    // ---- Edge: compound assignment is still a mutation ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.count += 1;
          }
        }
      `,
      errors: [expectNoDirectMutationError([4, 13, 4, 23])],
    },
    // ---- Edge: stateless FunctionDeclaration component ----
    {
      code: `
        function Hello() {
          this.state.x = 1;
          return <div/>;
        }
      `,
      errors: [expectNoDirectMutationError([3, 11, 3, 21])],
    },
    // ---- Edge: stateless arrow assigned to capital-cased variable ----
    {
      code: `
        const Hello = () => {
          this.state.x = 1;
          return <div/>;
        };
      `,
      errors: [expectNoDirectMutationError([3, 11, 3, 21])],
    },
  ],
});
