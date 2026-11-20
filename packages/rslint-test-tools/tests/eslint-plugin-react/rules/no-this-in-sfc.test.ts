import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoThisInSFCError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noThisInSFC',
    message: 'Stateless functional components should not use `this`',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-this-in-sfc', {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `
        function Foo(props) {
          const { foo } = props;
          return <div bar={foo} />;
        }
      `,
    },
    {
      code: `
        function Foo({ foo }) {
          return <div bar={foo} />;
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          render() {
            const { foo } = this.props;
            return <div bar={foo} />;
          }
        }
      `,
    },
    {
      code: `
        const Foo = createReactClass({
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `,
    },
    {
      code: `
        function foo(bar) {
          this.bar = bar;
          this.props = 'baz';
          this.getFoo = function() {
            return this.bar + this.props;
          }
        }
      `,
    },
    {
      code: `
        function Foo(props) {
          return props.foo ? <span>{props.bar}</span> : null;
        }
      `,
    },
    {
      code: `
        function Foo(props) {
          if (props.foo) {
            return <div>{props.bar}</div>;
          }
          return null;
        }
      `,
    },
    {
      code: `
        function Foo(props) {
          if (props.foo) {
            something();
          }
          return null;
        }
      `,
    },
    { code: `const Foo = (props) => <span>{props.foo}</span>` },
    { code: `const Foo = ({ foo }) => <span>{foo}</span>` },
    {
      code: `const Foo = (props) => props.foo ? <span>{props.bar}</span> : null;`,
    },
    { code: `const Foo = ({ foo, bar }) => foo ? <span>{bar}</span> : null;` },
    {
      code: `
        class Foo {
          bar() {
            () => {
              this.something();
              return null;
            };
          }
        }
      `,
    },
    {
      code: `
        class Foo {
          bar = () => {
            this.something();
            return null;
          };
        }
      `,
    },
    {
      code: `
        export const Example = ({ prop }) => {
          return {
            handleClick: () => {},
            renderNode() {
              return <div onClick={this.handleClick} />;
            },
          };
        };
      `,
    },
    {
      code: `
        export const prepareLogin = new ValidatedMethod({
          name: "user.prepare",
          validate: new SimpleSchema({
          }).validator(),
          run({ remember }) {
              if (Meteor.isServer) {
                  const connectionId = this.connection.id;
                  return Methods.prepareLogin(connectionId, remember);
              }
              return null;
          },
        });
      `,
    },
    {
      code: `
        obj.notAComponent = function () {
          return this.a || null;
        };
      `,
    },
    {
      code: `
        $.fn.getValueAsStringWeak = function (): string | null {
          const val = this.length === 1 ? this.val() : null;

          return typeof val === 'string' ? val : null;
        };
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const Foo = React.createClass({
          render: function() {
            return <div>{this.props.foo}</div>;
          }
        });
      `,
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        export const prepareLogin = new ValidatedMethod({
          name: "user.prepare",
          validate: new SimpleSchema({
          }).validator(),
          run({ remember }) {
              if (Meteor.isServer) {
                  const connectionId = this.connection.id; // react/no-this-in-sfc
                  return Methods.prepareLogin(connectionId, remember);
              }
              return null;
          },
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `
        function Foo(props) {
          const { foo } = this.props;
          return <div>{foo}</div>;
        }
      `,
      errors: [expectNoThisInSFCError([3, 27, 3, 37])],
    },
    {
      code: `
        function Foo(props) {
          return <div>{this.props.foo}</div>;
        }
      `,
      errors: [expectNoThisInSFCError([3, 24, 3, 34])],
    },
    {
      code: `
        function Foo(props) {
          return <div>{this.state.foo}</div>;
        }
      `,
      errors: [expectNoThisInSFCError([3, 24, 3, 34])],
    },
    {
      code: `
        function Foo(props) {
          const { foo } = this.state;
          return <div>{foo}</div>;
        }
      `,
      errors: [expectNoThisInSFCError([3, 27, 3, 37])],
    },
    {
      code: `
        function Foo(props) {
          return props.foo ? <div>{this.props.bar}</div> : null;
        }
      `,
      errors: [expectNoThisInSFCError([3, 36, 3, 46])],
    },
    {
      code: `
        function Foo(props) {
          if (props.foo) {
            return <div>{this.props.bar}</div>;
          }
          return null;
        }
      `,
      errors: [expectNoThisInSFCError([4, 26, 4, 36])],
    },
    {
      code: `
        function Foo(props) {
          if (this.props.foo) {
            something();
          }
          return null;
        }
      `,
      errors: [expectNoThisInSFCError([3, 15, 3, 25])],
    },
    {
      code: `const Foo = (props) => <span>{this.props.foo}</span>`,
      errors: [expectNoThisInSFCError([1, 31, 1, 41])],
    },
    {
      code: `const Foo = (props) => this.props.foo ? <span>{props.bar}</span> : null;`,
      errors: [expectNoThisInSFCError([1, 24, 1, 34])],
    },
    {
      code: `
        function Foo(props) {
          function onClick(bar) {
            this.props.onClick();
          }
          return <div onClick={onClick}>{this.props.foo}</div>;
        }
      `,
      errors: [
        expectNoThisInSFCError([4, 13, 4, 23]),
        expectNoThisInSFCError([6, 42, 6, 52]),
      ],
    },
  ],
});
