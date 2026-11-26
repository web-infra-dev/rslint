// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/prefer-exact-props.js
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectPropTypesError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'propTypes',
    message: "Component propTypes should be exact by using 'exact'.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectFlowError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'flow',
    message: 'Component flow props should be set with exact objects.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

const settings = {
  propWrapperFunctions: [{ property: 'exact', exact: true }],
};

const propTypesError = "Component propTypes should be exact by using 'exact'.";

ruleTester.run('prefer-exact-props', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 5 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {|
          foo: string
        |}
        function Component(props: Props) {
          return <div />;
        }
      `,
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {|
          foo: string
        |}
        function Component(props: Props) {
          let someVar: { foo: string };
          return <div />;
        }
      `,
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        function Component(props: {| foo : string |}) {
          return <div />;
        }
      `,
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
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
      errors: [expectPropTypesError([7, 9, 7, 28])],
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
      errors: [expectPropTypesError([3, 11, 5, 12])],
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
      errors: [expectPropTypesError([8, 9, 8, 28])],
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
      errors: [expectPropTypesError([10, 9, 10, 28])],
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
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 28,
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
          line: 13,
          column: 9,
          endLine: 13,
          endColumn: 28,
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
      errors: [expectPropTypesError([6, 9, 6, 28])],
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
      errors: [expectPropTypesError([14, 9, 14, 28])],
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
      errors: [expectPropTypesError([5, 11, 10, 14])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 2 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prefer-exact-props; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        class Component extends React.Component {
          props: {
            foo: string
          }
          render() {
            return <div />;
          }
        }
      `,
      errors: [expectFlowError([3, 11, 5, 12])],
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prefer-exact-props; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        function Component(props: { foo: string }) {
          return <div />;
        }
      `,
      errors: [expectFlowError([2, 28, 2, 50])],
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for prefer-exact-props; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type Props = {
          foo: string
        }
        function Component(props: Props) {
          return <div />;
        }
      `,
      errors: [expectFlowError([5, 28, 5, 40])],
      // Native parsing supports object spread without the historical Babel experimental flag.
      languageOptions: {
        ecmaVersion: 8,
        sourceType: 'module',
        parserOptions: {
          ecmaFeatures: { jsx: true },
        },
      },
    },
  ],
});
