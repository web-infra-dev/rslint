// Ported from eslint-plugin-react v7.37.5, tests/lib/rules/default-props-match-prop-types.js.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
function expectDefaultHasNoTypeError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'defaultHasNoType',
    message: 'defaultProp "baz" has no corresponding propTypes declaration.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectRequiredHasDefaultError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'requiredHasDefault',
    message: 'defaultProp "bar" defined for isRequired propType.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectRequiredHasDefaultError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'requiredHasDefault',
    message: 'defaultProp "foo" defined for isRequired propType.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDefaultHasNoTypeError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'defaultHasNoType',
    message: 'defaultProp "frob" has no corresponding propTypes declaration.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);
const parserOptions = {
  ecmaVersion: 2018,
  sourceType: 'module',
  ecmaFeatures: { jsx: true },
};

ruleTester.run('default-props-match-prop-types', {
  valid: [
    // stateless components
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string.isRequired,
          bar: React.PropTypes.string.isRequired
        };
      `,
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = {
          foo: "foo"
        };
      `,
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.propTypes.foo = React.PropTypes.string;
        MyStatelessComponent.defaultProps = {
          foo: "foo"
        };
      `,
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = {
          bar: "bar"
        };
      `,
      options: [{ allowRequiredDefaults: true }],
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.propTypes.foo = React.PropTypes.string;
        MyStatelessComponent.defaultProps = {};
        MyStatelessComponent.defaultProps.foo = "foo";
      `,
    },
    {
      code: `
        function MyStatelessComponent({ foo }) {
          return <div>{foo}</div>;
        }
        MyStatelessComponent.propTypes = {};
        MyStatelessComponent.propTypes.foo = React.PropTypes.string;
        MyStatelessComponent.defaultProps = {};
        MyStatelessComponent.defaultProps.foo = "foo";
      `,
    },
    {
      code: `
        const types = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = {
          foo: "foo"
        };
      `,
    },
    {
      code: `
        const defaults = {
          foo: "foo"
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = defaults;
      `,
    },
    {
      code: `
        const defaults = {
          foo: "foo"
        };
        const types = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = defaults;
      `,
    },

    // createReactClass components
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          },
          propTypes: {
            foo: React.PropTypes.string.isRequired,
            bar: React.PropTypes.string.isRequired
          }
        });
      `,
    },
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          },
          propTypes: {
            foo: React.PropTypes.string,
            bar: React.PropTypes.string.isRequired
          },
          getDefaultProps: function() {
            return {
              foo: "foo"
            };
          }
        });
      `,
    },
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          },
          propTypes: {
            foo: React.PropTypes.string,
            bar: React.PropTypes.string
          },
          getDefaultProps: function() {
            return {
              foo: "foo",
              bar: "bar"
            };
          }
        });
      `,
    },
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          }
        });
      `,
    },

    // ES6 class component
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        Greeting.defaultProps = {
          foo: "foo"
        };
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        Greeting.propTypes.foo = React.PropTypes.string;
        Greeting.defaultProps = {
          foo: "foo"
        };
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        Greeting.propTypes.foo = React.PropTypes.string;
        Greeting.defaultProps = {};
        Greeting.defaultProps.foo = "foo";
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {};
        Greeting.propTypes.foo = React.PropTypes.string;
        Greeting.defaultProps = {};
        Greeting.defaultProps.foo = "foo";
      `,
    },

    // edge cases

    // not a react component
    {
      code: `
        function NotAComponent({ foo, bar }) {}
        NotAComponent.defaultProps = {
          bar: "bar"
        };
      `,
    },
    {
      code: `
        class Greeting {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.defaulProps = {
          bar: "bar"
        };
      `,
    },
    // external references
    {
      code: `
        const defaults = require("./defaults");
        const types = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = defaults;
      `,
    },
    {
      code: `
        const defaults = {
          foo: "foo"
        };
        const types = require("./propTypes");

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = defaults;
      `,
    },
    {
      code: `
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = require("./defaults").foo;

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
    },
    {
      code: `
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = require("./defaults").foo;
        MyStatelessComponent.defaultProps.bar = "bar";

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
    },
    {
      code: `
        import defaults from "./defaults";

        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = defaults;

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
      languageOptions: {
        parserOptions: Object.assign({ sourceType: 'module' }, parserOptions),
      },
    },
    {
      code: `
        import { foo } from "./defaults";

        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = foo;

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
      languageOptions: {
        parserOptions: Object.assign({ sourceType: 'module' }, parserOptions),
      },
    },
    // using spread operator
    {
      code: `
        const component = rowsOfType(GuestlistEntry, (rowData, ownProps) => ({
            ...rowData,
            onPress: () => ownProps.onPress(rowData.id),
        }));
      `,
    },
    {
      code: `
        MyStatelessComponent.propTypes = {
          ...stuff,
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = {
         foo: "foo"
        };
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
    },
    {
      code: `
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = {
         ...defaults,
        };
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          ...someProps,
          bar: React.PropTypes.string.isRequired
        };
      `,
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        Greeting.defaultProps = {
          ...defaults,
          bar: "bar"
        };
      `,
    },

    // with Flow annotations
    {
      code: `
        type Props = {
          foo: string
        };

        class Hello extends React.Component {
          props: Props;

          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
    },
    {
      code: `
        type Props = {
          foo: string,
          bar?: string
        };

        class Hello extends React.Component {
          props: Props;

          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }

        Hello.defaultProps = {
          bar: "bar"
        };
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          props: {
            foo: string,
            bar?: string
          };

          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }

        Hello.defaultProps = {
          bar: "bar"
        };
      `,
    },
    {
      code: `
        class Hello extends React.Component {
          props: {
            foo: string
          };

          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
    },
    {
      code: `
        function Hello(props: { foo?: string }) {
          return <div>Hello {props.foo}</div>;
        }

        Hello.defaultProps = { foo: "foo" };
      `,
    },
    {
      code: `
        function Hello(props: { foo: string }) {
          return <div>Hello {foo}</div>;
        }
      `,
    },
    {
      code: `
        const Hello = (props: { foo?: string }) => {
          return <div>Hello {props.foo}</div>;
        };

        Hello.defaultProps = { foo: "foo" };
      `,
    },
    {
      code: `
        const Hello = (props: { foo: string }) => {
          return <div>Hello {foo}</div>;
        };
      `,
    },
    {
      code: `
        const Hello = function(props: { foo?: string }) {
          return <div>Hello {props.foo}</div>;
        };

        Hello.defaultProps = { foo: "foo" };
      `,
    },
    {
      code: `
        const Hello = function(props: { foo: string }) {
          return <div>Hello {foo}</div>;
        };
      `,
    },
    {
      code: `
        type Props = {
          foo: string,
          bar?: string
        };

        type Props2 = {
          foo: string,
          baz?: string
        }

        function Hello(props: Props | Props2) {
          return <div>Hello {props.foo}</div>;
        }

        Hello.defaultProps = {
          bar: "bar",
          baz: "baz"
        };
      `,
    },
    {
      code: `
        type PropsA = { foo?: string };
        type PropsB = { bar?: string, fooBar: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props;
          static defaultProps = {
            foo: "foo",
          }

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,
    },
    {
      code: `
        import type Props from "fake";
        class Hello extends React.Component {
          props: Props;
          render () {
            return <div>Hello {this.props.name.firstname}</div>;
          }
        }
      `,
    },
    {
      code: `
        type Props = any;

        const Hello = function({ foo }: Props) {
          return <div>Hello {foo}</div>;
        };
      `,
    },
    {
      code: `
        import type ImportedProps from "fake";
        type Props = ImportedProps;
        function Hello(props: Props) {
          return <div>Hello {props.name.firstname}</div>;
        }
      `,
    },

    // don't error when variable is not in scope
    {
      code: `
        import type { ImportedType } from "fake";
        type Props = ImportedType;
        function Hello(props: Props) {
          return <div>Hello {props.name.firstname}</div>;
        }
      `,
    },
    // make sure error is not thrown with multiple assignments
    {
      code: `
        import type ImportedProps from "fake";
        type NestedProps = ImportedProps;
        type Props = NestedProps;
        function Hello(props: Props) {
          return <div>Hello {props.name.firstname}</div>;
        }
      `,
    },
    // don't error when variable is not in scope with intersection
    {
      code: `
        import type ImportedProps from "fake";
        type Props = ImportedProps & {
          foo: string
        };
        function Hello(props: Props) {
          return <div>Hello {props.name.firstname}</div>;
        }
      `,
    },
    {
      code: `
        import type { FieldProps } from 'redux-form';

        type Props = FieldProps & {
          name: string,
          type: string,
          label?: string,
          placeholder?: string,
          disabled?: boolean,
        };

        TextField.defaultProps = {
          label: '',
          placeholder: '',
          disabled: false,
        };
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 47 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type DefaultProps1 = {|
          bar1?: string
        |};
        type DefaultProps2 = {|
          ...DefaultProps1,
          bar2?: string
        |};
        type Props = {
          foo: string,
          ...DefaultProps2
        };

        function Hello(props: Props) {
          return <div>Hello {props.foo}</div>;
        }

        Hello.defaultProps = {
          bar1: "bar1",
          bar2: "bar2",
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type DefaultProps1 = {|
          bar1?: string
        |};
        type DefaultProps2 = {|
          ...DefaultProps1,
          bar2?: string
        |};
        type Props = {
          foo: string,
          ...DefaultProps2
        };

        class Hello extends React.Component<Props> {
          render() {
            return <div>Hello {props.foo}</div>;
          }
        }

        Hello.defaultProps = {
          bar1: "bar1",
          bar2: "bar2",
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    // stateless components
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = {
          baz: "baz"
        };
      `,
      errors: [expectDefaultHasNoTypeError1([10, 11, 10, 21])],
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = forbidExtraProps({
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        })
        MyStatelessComponent.defaultProps = {
          baz: "baz"
        };
      `,
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectDefaultHasNoTypeError1([10, 11, 10, 21])],
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        const propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.propTypes = forbidExtraProps(propTypes);
        MyStatelessComponent.defaultProps = {
          baz: "baz"
        };
      `,
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectDefaultHasNoTypeError1([11, 11, 11, 21])],
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = {
          baz: "baz"
        };
      `,
      errors: [expectDefaultHasNoTypeError1([10, 11, 10, 21])],
      options: [{ allowRequiredDefaults: true }],
    },
    {
      code: `
        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        MyStatelessComponent.defaultProps = {
          bar: "bar"
        };
        MyStatelessComponent.defaultProps.baz = "baz";
      `,
      errors: [
        expectRequiredHasDefaultError2([10, 11, 10, 21]),
        expectDefaultHasNoTypeError1([12, 9, 12, 54]),
      ],
    },
    {
      code: `
        const types = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = {
          bar: "bar"
        };
      `,
      errors: [expectRequiredHasDefaultError2([12, 11, 12, 21])],
    },
    {
      code: `
        const defaults = {
          foo: "foo"
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = {
          foo: React.PropTypes.string.isRequired,
          bar: React.PropTypes.string
        };
        MyStatelessComponent.defaultProps = defaults;
      `,
      errors: [expectRequiredHasDefaultError3([3, 11, 3, 21])],
    },
    {
      code: `
        const defaults = {
          foo: "foo"
        };
        const types = {
          foo: React.PropTypes.string.isRequired,
          bar: React.PropTypes.string
        };

        function MyStatelessComponent({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        MyStatelessComponent.propTypes = types;
        MyStatelessComponent.defaultProps = defaults;
      `,
      errors: [expectRequiredHasDefaultError3([3, 11, 3, 21])],
    },

    // createReactClass components
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          },
          propTypes: {
            foo: React.PropTypes.string,
            bar: React.PropTypes.string.isRequired
          },
          getDefaultProps: function() {
            return {
              baz: "baz"
            };
          }
        });
      `,
      errors: [expectDefaultHasNoTypeError1([12, 15, 12, 25])],
    },
    {
      code: `
        var Greeting = createReactClass({
          render: function() {
            return <div>Hello {this.props.foo} {this.props.bar}</div>;
          },
          propTypes: {
            foo: React.PropTypes.string.isRequired,
            bar: React.PropTypes.string
          },
          getDefaultProps: function() {
            return {
              foo: "foo"
            };
          }
        });
      `,
      errors: [expectRequiredHasDefaultError3([12, 15, 12, 25])],
    },

    // ES6 class component
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        Greeting.defaultProps = {
          baz: "baz"
        };
      `,
      errors: [expectDefaultHasNoTypeError1([14, 11, 14, 21])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          foo: React.PropTypes.string.isRequired,
          bar: React.PropTypes.string
        };
        Greeting.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectRequiredHasDefaultError3([14, 11, 14, 21])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          bar: React.PropTypes.string.isRequired
        };
        Greeting.propTypes.foo = React.PropTypes.string.isRequired;
        Greeting.defaultProps = {};
        Greeting.defaultProps.foo = "foo";
      `,
      errors: [expectRequiredHasDefaultError3([14, 9, 14, 42])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {
          bar: React.PropTypes.string
        };
        Greeting.propTypes.foo = React.PropTypes.string;
        Greeting.defaultProps = {};
        Greeting.defaultProps.baz = "baz";
      `,
      errors: [expectDefaultHasNoTypeError1([14, 9, 14, 42])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        Greeting.propTypes = {};
        Greeting.propTypes.foo = React.PropTypes.string.isRequired;
        Greeting.defaultProps = {};
        Greeting.defaultProps.foo = "foo";
      `,
      errors: [expectRequiredHasDefaultError3([12, 9, 12, 42])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        const props = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        Greeting.propTypes = props;
        const defaults = {
          bar: "bar"
        };
        Greeting.defaultProps = defaults;
      `,
      errors: [expectRequiredHasDefaultError2([15, 11, 15, 21])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
        }
        const props = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string
        };
        const defaults = {
          baz: "baz"
        };
        Greeting.propTypes = props;
        Greeting.defaultProps = defaults;
      `,
      errors: [expectDefaultHasNoTypeError1([14, 11, 14, 21])],
    },

    // ES6 classes with static getter methods
    {
      code: `
        class Hello extends React.Component {
          static get propTypes() {
            return {
              name: React.PropTypes.string.isRequired
            };
          }
          static get defaultProps() {
            return {
              name: "name"
            };
          }
          render() {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [
        {
          messageId: 'requiredHasDefault',
          message: 'defaultProp "name" defined for isRequired propType.',
          line: 10,
          column: 15,
          endLine: 10,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        class Hello extends React.Component {
          static get propTypes() {
            return {
              foo: React.PropTypes.string,
              bar: React.PropTypes.string
            };
          }
          static get defaultProps() {
            return {
              baz: "world"
            };
          }
          render() {
            return <div>Hello {this.props.bar}</div>;
          }
        }
      `,
      errors: [expectDefaultHasNoTypeError1([11, 15, 11, 27])],
    },
    {
      code: `
        const props = {
          foo: React.PropTypes.string
        };
        const defaults = {
          baz: "baz"
        };

        class Hello extends React.Component {
          static get propTypes() {
            return props;
          }
          static get defaultProps() {
            return defaults;
          }
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectDefaultHasNoTypeError1([6, 11, 6, 21])],
    },
    {
      code: `
        const defaults = {
          bar: "world"
        };

        class Hello extends React.Component {
          static get propTypes() {
            return {
              foo: React.PropTypes.string,
              bar: React.PropTypes.string.isRequired
            };
          }
          static get defaultProps() {
            return defaults;
          }
          render() {
            return <div>Hello {this.props.bar}</div>;
          }
        }
      `,
      errors: [expectRequiredHasDefaultError2([3, 11, 3, 23])],
    },

    // ES6 classes with property initializers
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
          static propTypes = {
            foo: React.PropTypes.string,
            bar: React.PropTypes.string.isRequired
          };
          static defaultProps = {
            bar: "bar"
          };
        }
      `,

      errors: [expectRequiredHasDefaultError2([13, 13, 13, 23])],
    },
    {
      code: `
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
          static propTypes = {
            foo: React.PropTypes.string,
            bar: React.PropTypes.string
          };
          static defaultProps = {
            baz: "baz"
          };
        }
      `,

      errors: [expectDefaultHasNoTypeError1([13, 13, 13, 23])],
    },
    {
      code: `
        const props = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string.isRequired
        };
        const defaults = {
          bar: "bar"
        };
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
          static propTypes = props;
          static defaultProps = defaults;
        }
      `,

      errors: [expectRequiredHasDefaultError2([7, 11, 7, 21])],
    },
    {
      code: `
        const props = {
          foo: React.PropTypes.string,
          bar: React.PropTypes.string
        };
        const defaults = {
          baz: "baz"
        };
        class Greeting extends React.Component {
          render() {
            return (
              <h1>Hello, {this.props.foo} {this.props.bar}</h1>
            );
          }
          static propTypes = props;
          static defaultProps = defaults;
        }
      `,

      errors: [expectDefaultHasNoTypeError1([7, 11, 7, 21])],
    },

    // edge cases
    {
      code: `
        let Greetings = {};
        Greetings.Hello = class extends React.Component {
          render () {
            return <div>Hello {this.props.foo}</div>;
          }
        }
        Greetings.Hello.propTypes = {
          foo: React.PropTypes.string.isRequired
        };
        Greetings.Hello.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectRequiredHasDefaultError3([12, 11, 12, 21])],
    },
    {
      code: `
        var Greetings = ({ foo = "foo" }) => {
          return <div>Hello {this.props.foo}</div>;
        }
        Greetings.propTypes = {
          foo: React.PropTypes.string.isRequired
        };
        Greetings.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectRequiredHasDefaultError3([9, 11, 9, 21])],
    },

    // with Flow annotations
    {
      code: `
        class Hello extends React.Component {
          props: {
            foo: string,
            bar?: string
          };

          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }

        Hello.defaultProps = {
          foo: "foo"
        };
      `,

      errors: [expectRequiredHasDefaultError3([14, 11, 14, 21])],
    },
    {
      code: `
        function Hello(props: { foo: string }) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = {
          foo: "foo"
        }
      `,

      errors: [expectRequiredHasDefaultError3([6, 11, 6, 21])],
    },
    {
      code: `
        type Props = {
          foo: string
        };

        function Hello(props: Props) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = {
          foo: "foo"
        }
      `,

      errors: [expectRequiredHasDefaultError3([10, 11, 10, 21])],
    },
    {
      code: `
        const Hello = (props: { foo: string, bar?: string }) => {
          return <div>Hello {props.foo}</div>;
        };
        Hello.defaultProps = { foo: "foo", bar: "bar" };
      `,

      errors: [expectRequiredHasDefaultError3([5, 32, 5, 42])],
    },

    {
      code: `
        type PropsA = { foo: string };
        type PropsB = { bar: string };
        type Props = PropsA & PropsB;

        class Bar extends React.Component {
          props: Props;
          static defaultProps = {
            fooBar: "fooBar",
            foo: "foo",
          }

          render() {
            return <div>{this.props.foo} - {this.props.bar}</div>
          }
        }
      `,

      errors: [
        {
          messageId: 'defaultHasNoType',
          message:
            'defaultProp "fooBar" has no corresponding propTypes declaration.',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 29,
        },
        expectRequiredHasDefaultError3([10, 13, 10, 23]),
      ],
    },
    {
      code: `
        class SomeComponent extends React.Component {
          render() {
            return <div />;
          }
        }
        SomeComponent.propTypes = {
          "firstProperty": PropTypes.string.isRequired,
        };

        SomeComponent.defaultProps = {
          "firstProperty": () => undefined
        };
      `,
      errors: [
        {
          messageId: 'requiredHasDefault',
          message:
            'defaultProp "firstProperty" defined for isRequired propType.',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 43,
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 31 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for default-props-match-prop-types; the native TypeScript annotation path reports 0 diagnostics instead of 2.',
      code: `
        type Props = {
          foo: string,
          bar?: string
        };

        type Props2 = {
          foo: string,
          baz?: string
        }

        function Hello(props: Props | Props2) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = { foo: "foo", frob: "frob" };
      `,
      errors: [
        expectRequiredHasDefaultError3([15, 32, 15, 42]),
        expectDefaultHasNoTypeError4([15, 44, 15, 56]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type DefaultProps = {
          baz?: string,
          bar?: string
        };

        type Props = {
          foo: string,
          ...DefaultProps
        }

        function Hello(props: Props) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = { foo: "foo", frob: "frob", baz: "bar" };
      `,
      errors: [
        expectRequiredHasDefaultError3([15, 32, 15, 42]),
        expectDefaultHasNoTypeError4([15, 44, 15, 56]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type DefaultProps = {
          baz?: string,
          bar?: string
        };

        type Props = {
          foo: string,
          ...DefaultProps
        }

        class Hello extends React.Component<Props> {
          render() {
            return <div>Hello {props.foo}</div>;
          }
        }
        Hello.defaultProps = { foo: "foo", frob: "frob", baz: "bar" };
      `,
      errors: [
        expectRequiredHasDefaultError3([17, 32, 17, 42]),
        expectDefaultHasNoTypeError4([17, 44, 17, 56]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        export type SharedProps = {|
            disabled: boolean,
        |};

        type Props = {|
            ...SharedProps,
            focused?: boolean,
        |};

        class Foo extends React.Component<Props> {
            static defaultProps = {
              disabled: false
            };
        };
      `,
      errors: [
        {
          messageId: 'requiredHasDefault',
          message: 'defaultProp "disabled" defined for isRequired propType.',
          line: 13,
          column: 15,
          endLine: 13,
          endColumn: 30,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
