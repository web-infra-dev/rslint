import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectShouldUseES6ClassError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'shouldUseES6Class',
    'Component should use es6 class instead of createClass',
    range,
  );
}

function expectShouldUseCreateClassError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldUseCreateClass',
    'Component should use createClass instead of es6 class',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('prefer-es6-class', {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
      `,
    },
    {
      code: `
        export default class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.displayName = 'Hello'
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
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      options: ['never'],
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      options: ['always'],
    },

    // ---- Edge: class expression extending React.Component not reported ----
    {
      code: `
        const Hello = class extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        };
      `,
      options: ['never'],
    },

    // ---- Edge: non-matching callee not reported ----
    {
      code: `
        var x = something({ foo: 1 });
      `,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      errors: [expectShouldUseES6ClassError([2, 38, 7, 10])],
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      options: ['always'],
      errors: [expectShouldUseES6ClassError([2, 38, 6, 10])],
    },
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      options: ['never'],
      errors: [expectShouldUseCreateClassError([2, 9, 6, 10])],
    },

    // ---- Edge: pragma-qualified React.createReactClass ----
    {
      code: `
        var Hello = React.createReactClass({
          render: function() { return <div/>; }
        });
      `,
      errors: [expectShouldUseES6ClassError([2, 44, 4, 10])],
    },

    // ---- Edge: PureComponent variant reported in mode "never" ----
    {
      code: `
        class Hello extends React.PureComponent {
          render() { return <div/>; }
        }
      `,
      options: ['never'],
      errors: [expectShouldUseCreateClassError([2, 9, 4, 10])],
    },
    {
      code: `
        class Hello extends PureComponent {
          render() { return <div/>; }
        }
      `,
      options: ['never'],
      errors: [expectShouldUseCreateClassError([2, 9, 4, 10])],
    },

    // ---- Edge: `new createReactClass({...})` — NewExpression also has
    // `.callee` in ESTree, upstream fires. ----
    {
      code: `
        var Hello = new createReactClass({
          render: function() { return <div/>; }
        });
      `,
      errors: [expectShouldUseES6ClassError([2, 42, 4, 10])],
    },
  ],
});
