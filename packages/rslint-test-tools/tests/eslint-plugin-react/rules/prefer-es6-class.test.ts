import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-es6-class', {} as never, {
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
      errors: [{ messageId: 'shouldUseES6Class' }],
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
      errors: [{ messageId: 'shouldUseES6Class' }],
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
      errors: [{ messageId: 'shouldUseCreateClass' }],
    },

    // ---- Edge: pragma-qualified React.createReactClass ----
    {
      code: `
        var Hello = React.createReactClass({
          render: function() { return <div/>; }
        });
      `,
      errors: [{ messageId: 'shouldUseES6Class' }],
    },

    // ---- Edge: PureComponent variant reported in mode "never" ----
    {
      code: `
        class Hello extends React.PureComponent {
          render() { return <div/>; }
        }
      `,
      options: ['never'],
      errors: [{ messageId: 'shouldUseCreateClass' }],
    },
    {
      code: `
        class Hello extends PureComponent {
          render() { return <div/>; }
        }
      `,
      options: ['never'],
      errors: [{ messageId: 'shouldUseCreateClass' }],
    },

    // ---- Edge: `new createReactClass({...})` — NewExpression also has
    // `.callee` in ESTree, upstream fires. ----
    {
      code: `
        var Hello = new createReactClass({
          render: function() { return <div/>; }
        });
      `,
      errors: [{ messageId: 'shouldUseES6Class' }],
    },
  ],
});
