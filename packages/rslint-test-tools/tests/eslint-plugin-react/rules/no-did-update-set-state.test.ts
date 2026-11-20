import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoSetStateError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noSetState',
    message: 'Do not use setState in componentDidUpdate',
    line,
    column,
    endLine,
    endColumn,
  };
}

// Reuse each lifecycle input under the upstream React 16.3 version override.
const lifecycleCases: TestCase[] = [
  {
    name: 'Upstream invalid 0 (7.37.5)',
    code: `
      var Hello = createReactClass({
        componentDidUpdate: function() {
          this.setState({
            data: data
          });
        }
      });
    `,
    errors: [expectNoSetStateError([4, 11, 4, 24])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 1 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate() {
          this.setState({
            data: data
          });
        }
      }
    `,
    errors: [expectNoSetStateError([4, 11, 4, 24])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 2 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate = () => {
          this.setState({
            data: data
          });
        }
      }
    `,
    errors: [expectNoSetStateError([4, 11, 4, 24])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 3 (7.37.5)',
    code: `
    var Hello = createReactClass({
      componentDidUpdate: function() {
        this.setState({
          data: data
        });
      }
    });
  `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([4, 9, 4, 22])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 4 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate() {
          this.setState({
            data: data
          });
        }
      }
    `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([4, 11, 4, 24])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 5 (7.37.5)',
    code: `
      var Hello = createReactClass({
        componentDidUpdate: function() {
          someClass.onSomeEvent(function(data) {
            this.setState({
              data: data
            });
          })
        }
      });
    `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([5, 13, 5, 26])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 6 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate() {
          someClass.onSomeEvent(function(data) {
            this.setState({
              data: data
            });
          })
        }
      }
    `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([5, 13, 5, 26])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 7 (7.37.5)',
    code: `
      var Hello = createReactClass({
        componentDidUpdate: function() {
          if (true) {
            this.setState({
              data: data
            });
          }
        }
      });
    `,
    errors: [expectNoSetStateError([5, 13, 5, 26])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 8 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate() {
          if (true) {
            this.setState({
              data: data
            });
          }
        }
      }
    `,
    errors: [expectNoSetStateError([5, 13, 5, 26])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 9 (7.37.5)',
    code: `
      var Hello = createReactClass({
        componentDidUpdate: function() {
          someClass.onSomeEvent((data) => this.setState({data: data}));
        }
      });
    `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([4, 43, 4, 56])],
    languageOptions: upstreamModuleOptions,
  },
  {
    name: 'Upstream invalid 10 (7.37.5)',
    code: `
      class Hello extends React.Component {
        componentDidUpdate() {
          someClass.onSomeEvent((data) => this.setState({data: data}));
        }
      }
    `,
    options: ['disallow-in-func'],
    errors: [expectNoSetStateError([4, 43, 4, 56])],
    languageOptions: upstreamModuleOptions,
  },
];
const lifecycleVersionNames = [
  'Upstream valid 5 (7.37.5)',
  'Upstream valid 6 (7.37.5)',
  'Upstream valid 7 (7.37.5)',
  'Upstream valid 8 (7.37.5)',
  'Upstream valid 9 (7.37.5)',
  'Upstream valid 10 (7.37.5)',
  'Upstream valid 11 (7.37.5)',
  'Upstream valid 12 (7.37.5)',
  'Upstream valid 13 (7.37.5)',
  'Upstream valid 14 (7.37.5)',
  'Upstream valid 15 (7.37.5)',
];

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-did-update-set-state', {
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
          componentDidUpdate: function() {}
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
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
          componentDidUpdate: function() {
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
          componentDidUpdate() {
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
          componentDidUpdate: function() {
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
          componentDidUpdate: function() {
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
    ...lifecycleCases.map(
      ({ errors: _errors, output: _output, name: _name, ...entry }, index) => ({
        ...entry,
        name: lifecycleVersionNames[index],
        settings: { react: { version: '16.3.0' } },
      }),
    ),
  ],
  invalid: [
    {
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            this.setState({ data: data });
          }
        });
      `,
      errors: [expectNoSetStateError([4, 13, 4, 26])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.setState({ data: data });
          }
        }
      `,
      errors: [expectNoSetStateError([4, 13, 4, 26])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate = () => {
            this.setState({ data: data });
          }
        }
      `,
      errors: [expectNoSetStateError([4, 13, 4, 26])],
    },
    // ---- disallow-in-func flags setState in nested callbacks ----
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            someClass.onSomeEvent(function(data) {
              this.setState({ data: data });
            });
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError([5, 15, 5, 28])],
    },
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            someClass.onSomeEvent((data) => this.setState({ data: data }));
          }
        }
      `,
      options: ['disallow-in-func'],
      errors: [expectNoSetStateError([4, 45, 4, 58])],
    },
    // ---- setState inside if-block ----
    {
      code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            if (true) {
              this.setState({ data: data });
            }
          }
        }
      `,
      errors: [expectNoSetStateError([5, 15, 5, 28])],
    },

    // Additional input/configuration combinations from pinned upstream.
    lifecycleCases[0],
    lifecycleCases[1],
    lifecycleCases[2],
    lifecycleCases[3],
    lifecycleCases[4],
    lifecycleCases[5],
    lifecycleCases[6],
    lifecycleCases[7],
    lifecycleCases[8],
    lifecycleCases[9],
    lifecycleCases[10],
  ],
});
