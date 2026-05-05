import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-did-mount-set-state', {} as never, {
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
          componentDidMount: function() {}
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            someNonMemberFunction(arg);
            this.someHandler = this.setState;
          }
        });
      `,
    },
    // ---- Default mode allows setState in nested callbacks ----
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({ data: data });
            });
          }
        });
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent((data) => this.setState({ data: data }));
          }
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            this.setState({ data: data });
          }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.setState({ data: data });
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidMount = () => {
            this.setState({ data: data });
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- disallow-in-func flags setState in nested callbacks ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent(function(data) {
              this.setState({ data: data });
            });
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            someClass.onSomeEvent((data) => this.setState({ data: data }));
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- setState inside if-block ----
    {
      code: `
        class Hello extends React.Component {
          componentDidMount() {
            if (true) {
              this.setState({ data: data });
            }
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
  ],
});
