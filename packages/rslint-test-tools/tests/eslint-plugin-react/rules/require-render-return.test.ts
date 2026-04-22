import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-render-return', {} as never, {
  valid: [
    // ---- Upstream: ES6 class ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },
    // ---- Upstream: ES6 class with render property (arrow + block + return) ----
    {
      code: `
        class Hello extends React.Component {
          render = () => {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
    },
    // ---- Upstream: ES6 class with render property (implicit return) ----
    {
      code: `
        class Hello extends React.Component {
          render = () => (
            <div>Hello {this.props.name}</div>
          )
        }
      `,
    },
    // ---- Upstream: ES5 class ----
    {
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {
            return <div></div>
          }
        });
      `,
    },
    // ---- Upstream: Stateless function ----
    {
      code: `
        function Hello() {
          return <div></div>;
        }
      `,
    },
    // ---- Upstream: Stateless arrow function ----
    {
      code: `
        var Hello = () => (
          <div></div>
        );
      `,
    },
    // ---- Upstream: Return in a switch...case ----
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            switch (this.props.name) {
              case 'Foo':
                return <div>Hello Foo</div>;
              default:
                return <div>Hello {this.props.name}</div>;
            }
          }
        });
      `,
    },
    // ---- Upstream: Return in a if...else ----
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            if (this.props.name === 'Foo') {
              return <div>Hello Foo</div>;
            } else {
              return <div>Hello {this.props.name}</div>;
            }
          }
        });
      `,
    },
    // ---- Upstream: Not a React component (class doesn't extend Component) ----
    {
      code: `
        class Hello {
          render() {}
        }
      `,
    },
    // ---- Upstream: ES6 class without a render method ----
    {
      code: 'class Hello extends React.Component {}',
    },
    // ---- Upstream: ES5 class without a render method ----
    {
      code: 'var Hello = createReactClass({});',
    },
    // ---- Upstream: ES5 class with imported (shorthand) render method ----
    {
      code: `
        var render = require('./render');
        var Hello = createReactClass({
          render
        });
      `,
    },
    // ---- Upstream: Invalid render method (field without initializer) ----
    {
      code: `
        class Foo extends Component {
          render
        }
      `,
    },
  ],
  invalid: [
    // ---- Upstream: Missing return in ES5 class ----
    {
      code: `
        var Hello = createReactClass({
          displayName: 'Hello',
          render: function() {}
        });
      `,
      errors: [{ messageId: 'noRenderReturn' }],
    },
    // ---- Upstream: Missing return in ES6 class ----
    {
      code: `
        class Hello extends React.Component {
          render() {}
        }
      `,
      errors: [{ messageId: 'noRenderReturn' }],
    },
    // ---- Upstream: Missing return (but one is present in a sub-function) ----
    {
      code: `
        class Hello extends React.Component {
          render() {
            const names = this.props.names.map(function(name) {
              return <div>{name}</div>
            });
          }
        }
      `,
      errors: [{ messageId: 'noRenderReturn' }],
    },
    // ---- Upstream: Missing return ES6 class render property ----
    {
      code: `
        class Hello extends React.Component {
          render = () => {
            <div>Hello {this.props.name}</div>
          }
        }
      `,
      errors: [{ messageId: 'noRenderReturn' }],
    },
  ],
});
