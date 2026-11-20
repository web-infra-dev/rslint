import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoSetStateError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noSetState',
    'Do not use setState in componentWillUpdate',
    range,
  );
}

function expectNoSetStateError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noSetState',
    'Do not use setState in UNSAFE_componentWillUpdate',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-will-update-set-state', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            function handleEvent(data) {
              this.setState({
                data: data
              });
            }
            someClass.onSomeEvent(handleEvent)
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
      settings: { react: { version: '16.2.0' } },
      languageOptions: upstreamModuleOptions,
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
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({ data: data });
          }
        }
      `,
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
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
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
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
      errors: [expectNoSetStateError1([4, 45, 4, 58])],
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
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
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
      errors: [expectNoSetStateError2([4, 13, 4, 26])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        });
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent(function(data) {
              this.setState({
                data: data
              });
            })
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        });
      `,
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            if (true) {
              this.setState({
                data: data
              });
            }
          }
        }
      `,
      errors: [expectNoSetStateError1([5, 15, 5, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        });
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([4, 45, 4, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            someClass.onSomeEvent((data) => this.setState({data: data}));
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError1([4, 45, 4, 58])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class Hello extends React.Component {
          UNSAFE_componentWillUpdate() {
            this.setState({
              data: data
            });
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectNoSetStateError2([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {
            this.setState({
              data: data
            });
          }
        });
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectNoSetStateError2([4, 13, 4, 26])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
