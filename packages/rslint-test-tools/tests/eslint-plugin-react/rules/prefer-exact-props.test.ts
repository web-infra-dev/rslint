// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/prefer-exact-props.js
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const settings = {
  propWrapperFunctions: [{ property: 'exact', exact: true }],
};

const propTypesError = "Component propTypes should be exact by using 'exact'.";

ruleTester.run('prefer-exact-props', {} as never, {
  valid: [
    {
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {};
      `,
      settings,
    },
    {
      code: `
        class Component extends React.Component {
          static propTypes = {};
          render() {
            return <div />;
          }
        }
      `,
      settings,
    },
    {
      code: `
        class Component extends React.Component {
          props: {};
          render() {
            return <div />;
          }
        }
      `,
      settings,
    },
    {
      code: `
        function Component(props) {
          return <div />;
        }
        Component.propTypes = {};
      `,
      settings,
    },
    {
      code: `
        function Component(props: {}) {
          return <div />;
        }
      `,
    },
    {
      code: `
        type Props = {}
        function Component(props: Props) {
          return <div />;
        }
      `,
    },
    {
      code: `
        import type Props from 'foo';
        function Component(props: Props) {
          return <div />;
        }
      `,
    },
    {
      code: `
        const props = {};
        function Component(props) {
          return <div />;
        }
        Component.propTypes = props;
      `,
      settings,
    },
    {
      code: `
        const props = {};
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `,
      settings,
    },
    {
      code: `
        import props from 'foo';
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `,
      settings,
    },
    {
      code: `
        class Component extends React.Component {
          state = {hi: 'hi'}
          render() {
            return <div>{this.state.hi}</div>;
          }
        }
      `,
    },
    {
      code: `
        import exact from "prop-types-exact";
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = exact({
          foo: PropTypes.string,
          bar: PropTypes.string,
        });
      `,
      settings,
    },
    {
      code: `
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
    },
    {
      code: `
        class Component extends React.Component {
          render() {
            const { foo, bar } = this.props;
            return <div>{foo}{bar}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `,
    },
    {
      code: `
        import somethingElse from "something-else";
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = somethingElse(props);
      `,
    },
    {
      code: `
        import somethingElse from "something-else";
        const props =
        class Component extends React.Component {
          static propTypes = somethingElse({
            foo: PropTypes.string,
            bar: PropTypes.shape({
              baz: PropTypes.string
            })
          });
          render() {
            return <div />;
          }
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          foo: PropTypes.string
        };
      `,
      settings,
      errors: [
        {
          messageId: 'propTypes',
          message: propTypesError,
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        class Component extends React.Component {
          static propTypes = {
            foo: PropTypes.string
          }
          render() {
            return <div />;
          }
        }
      `,
      settings,
      errors: [
        {
          messageId: 'propTypes',
          message: propTypesError,
          line: 3,
          column: 11,
          endLine: 5,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
        const props = {
          foo: PropTypes.string
        };
        function Component(props) {
          return <div />;
        }
        Component.propTypes = props;
      `,
      settings,
      errors: [{ messageId: 'propTypes', message: propTypesError }],
    },
    {
      code: `
        const props = {
          foo: PropTypes.string
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `,
      settings,
      errors: [{ messageId: 'propTypes', message: propTypesError }],
    },
    {
      code: `
        const props = {
          foo: PropTypes.string
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `,
      settings: {
        propWrapperFunctions: [
          { property: 'exact', exact: true },
          { property: 'forbidExtraProps', exact: true },
        ],
      },
      errors: [
        {
          messageId: 'propTypes',
          message:
            "Component propTypes should be exact by using one of 'exact', 'forbidExtraProps'.",
        },
      ],
    },
    {
      code: `
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `,
      settings: {
        propWrapperFunctions: [
          { property: 'exact', exact: true },
          { property: 'forbidExtraProps', exact: true },
        ],
      },
      errors: [
        {
          messageId: 'propTypes',
          message:
            "Component propTypes should be exact by using one of 'exact', 'forbidExtraProps'.",
        },
      ],
    },
    {
      code: `
        import somethingElse from "something-else";
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = somethingElse({
          foo: PropTypes.string,
          bar: PropTypes.string,
        });
      `,
      settings,
      errors: [{ messageId: 'propTypes', message: propTypesError }],
    },
    {
      code: `
        import somethingElse from "something-else";
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = somethingElse(props);
      `,
      settings,
      errors: [{ messageId: 'propTypes', message: propTypesError }],
    },
    {
      code: `
        import somethingElse from "something-else";
        const props =
        class Component extends React.Component {
          static propTypes = somethingElse({
            foo: PropTypes.string,
            bar: PropTypes.shape({
              baz: PropTypes.string
            })
          });
          render() {
            return <div />;
          }
        }
      `,
      settings,
      errors: [{ messageId: 'propTypes', message: propTypesError }],
    },
  ],
});
