import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamScriptOptions } from '../config';
import { diagnostic } from '../expectations';

function expectUseCallbackError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'useCallback',
    'Use callback in setState when referencing the previous state.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

// Upstream's suite passes `settings.react.createClass = 'createClass'` to the
// tester, which wires the createReactClass matcher to `<pragma>.createClass(...)`
// (e.g. `React.createClass`). rslint's default is `createReactClass`; since
// JS tests here cannot thread per-case settings through, upstream cases that
// used `React.createClass` are rewritten with `createReactClass` so the
// matcher fires without needing a settings override.

ruleTester.run('no-access-state-in-setstate', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(state => ({value: state.value + 1}))
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = React.createClass({
          multiplyValue: function(obj) {
            return obj.value*2
          },
          onClick: function() {
            var value = this.state.value
            this.multiplyValue({ value: value })
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
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
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({}, () => console.log(this.state));
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({}, () => 1 + 1);
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            var nextValueNotUsed = this.state.value + 1
            var nextValue = 2
            this.setState({value: nextValue})
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        function testFunction({a, b}) {
        };
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class ComponentA extends React.Component {
          state = {
            greeting: 'hello',
          };

          myFunc = () => {
            this.setState({ greeting: 'hi' }, () => this.doStuff());
          };

          doStuff = () => {
            console.log(this.state.greeting);
          };
        }
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: { sourceType: 'script' },
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class Foo extends Abstract {
          update = () => {
            const result = this.getResult ( this.state.foo );
            return this.setState ({ result });
          };
        }
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        class StateContainer extends Container {
          anything() {
            return this.setState({value: this.state.value + 1})
          }
        };
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamScriptOptions,
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
      errors: [expectUseCallbackError([4, 36, 4, 46])],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(() => ({ value: this.state.value + 1 }))
          }
        });
      `,
      errors: [expectUseCallbackError([4, 43, 4, 53])],
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
      errors: [expectUseCallbackError([4, 29, 4, 39])],
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
      errors: [expectUseCallbackError([4, 19, 4, 24])],
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
      errors: [expectUseCallbackError([7, 37, 7, 47])],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(this.state, () => 1 + 1);
          }
        });
      `,
      errors: [expectUseCallbackError([4, 27, 4, 37])],
    },
    {
      code: `
        var Hello = createReactClass({
          onClick: function() {
            this.setState(this.state, () => console.log(this.state));
          }
        });
      `,
      errors: [expectUseCallbackError([4, 27, 4, 37])],
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
      errors: [expectUseCallbackError([4, 29, 4, 39])],
    },
    {
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.state, () => console.log(this.state));
          }
        }
      `,
      errors: [expectUseCallbackError([4, 27, 4, 37])],
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
      errors: [
        expectUseCallbackError([4, 36, 4, 46]),
        expectUseCallbackError([5, 36, 5, 46]),
      ],
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
      errors: [expectUseCallbackError([4, 19, 4, 24])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState({value: this.state.value + 1})
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 35, 4, 45])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(() => ({value: this.state.value + 1}))
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 42, 4, 52])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            var nextValue = this.state.value + 1
            this.setState({value: nextValue})
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 29, 4, 39])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            var {state, ...rest} = this
            this.setState({value: state.value + 1})
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 18, 4, 23])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        function nextState(state) {
          return {value: state.value + 1}
        }
        var Hello = React.createClass({
          onClick: function() {
            this.setState(nextState(this.state))
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([7, 37, 7, 47])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(this.state, () => 1 + 1);
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 27, 4, 37])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var Hello = React.createClass({
          onClick: function() {
            this.setState(this.state, () => console.log(this.state));
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 27, 4, 37])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        var Hello = React.createClass({
          nextState: function() {
            return {value: this.state.value + 1}
          },
          onClick: function() {
            this.setState(nextState())
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 28, 4, 38])],
      languageOptions: upstreamScriptOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Hello extends React.Component {
          onClick() {
            this.setState(this.state, () => console.log(this.state));
          }
        }
      `,
      settings: { react: { createClass: 'createClass' } },
      errors: [expectUseCallbackError([4, 27, 4, 37])],
      languageOptions: upstreamScriptOptions,
    },
  ],
});
