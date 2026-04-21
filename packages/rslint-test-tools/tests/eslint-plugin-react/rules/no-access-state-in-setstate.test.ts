import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Upstream's suite passes `settings.react.createClass = 'createClass'` to the
// tester, which wires the createReactClass matcher to `<pragma>.createClass(...)`
// (e.g. `React.createClass`). rslint's default is `createReactClass`; since
// JS tests here cannot thread per-case settings through, upstream cases that
// used `React.createClass` are rewritten with `createReactClass` so the
// matcher fires without needing a settings override.

ruleTester.run('no-access-state-in-setstate', {} as never, {
  valid: [
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(state => ({ value: state.value + 1 }))
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          multiplyValue: function(obj) {
            return obj.value*2
          },
          onClick: function() {
            var value = this.state.value
            this.multiplyValue({ value: value })
          }
        });
      `,
    },
    {
      code: `
        var SearchForm = createReactClass({
          render: function () {
            return (
              <div>
                {(function () {
                  if (this.state.prompt) {
                    return <div>{this.state.prompt}</div>
                  }
                }).call(this)}
              </div>
            );
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState({}, () => console.log(this.state));
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState({}, () => 1 + 1);
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            var nextValueNotUsed = this.state.value + 1
            var nextValue = 2
            this.setState({ value: nextValue })
          }
        });
      `,
    },
    {
      code: `
        function testFunction({ a, b }) {
        };
      `,
    },
    {
      code: `
        class ComponentA extends React.Component {
          state = { greeting: 'hello' };
          myFunc = () => {
            this.setState({ greeting: 'hi' }, () => this.doStuff());
          };
          doStuff = () => {
            console.log(this.state.greeting);
          };
        }
      `,
    },
    {
      code: `
        class Foo extends Abstract {
          update = () => {
            const result = this.getResult(this.state.foo);
            return this.setState({ result });
          };
        }
      `,
    },
    {
      code: `
        class StateContainer extends Container {
          anything() {
            return this.setState({ value: this.state.value + 1 })
          }
        };
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({ value: this['state'].value + 1 });
          }
        }
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this['setState']({ value: this.state.value + 1 });
          }
        }
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            var { state: aliased } = this;
            this.setState({ value: aliased.value + 1 });
          }
        }
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState();
          }
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState({ value: this.state.value + 1 })
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(() => ({ value: this.state.value + 1 }))
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            var nextValue = this.state.value + 1
            this.setState({ value: nextValue })
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            var { state, ...rest } = this
            this.setState({ value: state.value + 1 })
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        function nextState(state) {
          return { value: state.value + 1 }
        }
        var Hello = createReactClass({
          onClick: function() {
            this.setState(nextState(this.state))
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(this.state, () => 1 + 1);
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(this.state, () => console.log(this.state));
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        var Hello = createReactClass({
          nextState: function() {
            return { value: this.state.value + 1 }
          },
          onClick: function() {
            this.setState(nextState())
          }
        });
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.state, () => console.log(this.state));
          }
        }
      `,
      errors: [{ messageId: 'useCallback' }],
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState({ value: this.state.value + 1 });
            this.setState({ value: this.state.value + 2 });
          }
        }
      `,
      errors: [{ messageId: 'useCallback' }, { messageId: 'useCallback' }],
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            var { state } = this;
            this.setState({ state });
          }
        }
      `,
      errors: [{ messageId: 'useCallback' }],
    },
  ],
});
