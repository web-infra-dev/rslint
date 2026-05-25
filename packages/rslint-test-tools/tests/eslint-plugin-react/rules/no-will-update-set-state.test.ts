import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-will-update-set-state', {} as never, {
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
          componentWillUpdate: function() {}
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
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
          componentWillUpdate: function() {
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
          componentWillUpdate() {
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
          componentWillUpdate: function() {
            this.setState({ data: data });
          }
        });
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
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
          componentWillUpdate() {
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
          componentWillUpdate() {
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
          componentWillUpdate() {
            if (true) {
              this.setState({ data: data });
            }
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
    // ---- UNSAFE_ alias flagged when react version unset (defaults to latest) ----
    {
      code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({ data: data });
          }
        }
      `,
      errors: [{ messageId: 'noSetState' }],
    },
  ],
});
