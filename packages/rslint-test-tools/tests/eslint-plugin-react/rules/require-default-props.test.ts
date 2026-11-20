// Mirrors the parser-neutral and TypeScript cases from eslint-plugin-react v7.37.5.
// Flow-only cases are retained as explicit skips in the Go upstream suite.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';
function expectShouldHaveDefaultError1(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldHaveDefault',
    'propType "foo" is not required, but has no corresponding defaultProps declaration.',
    range,
  );
}

function expectShouldHaveDefaultError2(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldHaveDefault',
    'propType "baz" is not required, but has no corresponding defaultProps declaration.',
    range,
  );
}

function expectShouldHaveDefaultError3(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldHaveDefault',
    'propType "bar" is not required, but has no corresponding defaultProps declaration.',
    range,
  );
}

function expectShouldHaveDefaultError4(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldHaveDefault',
    'propType "name" is not required, but has no corresponding defaultProps declaration.',
    range,
  );
}

function expectNoDefaultWithRequiredError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noDefaultWithRequired',
    'propType "foo" is required and should not have a defaultProps declaration.',
    range,
  );
}

function expectShouldHaveDefaultError6(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'shouldHaveDefault',
    'propType "two" is not required, but has no corresponding defaultProps declaration.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);
ruleTester.run('require-default-props', {
  valid: [
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string.isRequired\n        };\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.foo = PropTypes.string;\n        MyStatelessComponent.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.foo = PropTypes.string;\n        MyStatelessComponent.defaultProps = {};\n        MyStatelessComponent.defaultProps.foo = "foo";\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo }) {\n          return <div>{foo}</div>;\n        }\n        MyStatelessComponent.propTypes = {};\n        MyStatelessComponent.propTypes.foo = PropTypes.string;\n        MyStatelessComponent.defaultProps = {};\n        MyStatelessComponent.defaultProps.foo = "foo";\n      ',
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = defaults;\n      ',
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ',
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo = "test", bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo = "test", bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n        };\n      ',
      options: [
        {
          forbidDefaultForRequired: true,
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        export function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        export default function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: "\n        function Hello(props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string\n        };\n        Hello.defaultProps = {\n          foo: 'bar;'\n        }\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
    },
    {
      code: "\n        function Hello({ foo = 'asdf', bar }) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: "\n        function Hello({ foo = 'asdf', bar = 'qwer' }) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        function Hello(props) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string.isRequired,\n        };\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: "\n        function MyStatelessComponent({ foo, bar = 'asdf', ...props }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string,\n          hello: PropTypes.string.isRequired,\n          world: PropTypes.string.isRequired\n        }\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar, ...props }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string.isRequired,\n          hello: PropTypes.string.isRequired,\n          world: PropTypes.string.isRequired\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        const MyComponent = function({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        };\n\n        MyComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n\n        export default MyComponent;\n      ",
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        export const MyComponent = function({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        };\n\n        MyComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ",
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: "\n        const Hello = function({ foo = 'asdf', bar = 'qwer' }) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        const NoPropsComponent = function () {\n          return <div>Hello, world!</div>;\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        function NoPropsComponent() {\n          return <div>Hello, world!</div>;\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        const NoPropsComponent = () => {\n          return <div>Hello, world!</div>;\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        const NoPropsComponent = function () {\n          return <div>Hello, world!</div>;\n        }\n        NoPropsComponent.propTypes = {};\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        const MyComponent = ({ foo, bar }) => {\n          return <div>{foo}{bar}</div>;\n        };\n\n        MyComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n\n        export default MyComponent;\n      ",
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        export const MyComponent = ({ foo, bar }) => {\n          return <div>{foo}{bar}</div>;\n        };\n\n        MyComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n\n        export default MyComponent;\n      ",
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: "\n        const Hello = ({ foo = 'asdf', bar = 'qwer' }) => {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string.isRequired,\n            bar: PropTypes.string.isRequired\n          }\n        });\n      ',
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo"\n            };\n          }\n        });\n      ',
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo",\n              bar: "bar"\n            };\n          }\n        });\n      ',
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          }\n        });\n      ',
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo",\n              bar: "bar"\n            };\n          }\n        });\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo",\n              bar: "bar"\n            };\n          }\n        });\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = "foo";\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {};\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = "foo";\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          classes: 'ignore',
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
      options: [
        {
          classes: 'ignore',
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      options: [
        {
          classes: 'ignore',
        },
      ],
    },
    {
      code: '\n        function NotAComponent({ foo, bar }) {}\n        NotAComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
    },
    {
      code: '\n        class Greeting {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n      ',
    },
    {
      code: '\n        let Greetings = {};\n        Greetings.Hello = class extends React.Component {\n          render () {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Greetings.Hello.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      options: [
        {
          classes: 'ignore',
        },
      ],
    },
    {
      code: '\n        const defaults = require("./defaults");\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ',
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        const types = require("./propTypes");\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ',
    },
    {
      code: '\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = require("./defaults").foo;\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = require("./defaults").foo;\n        MyStatelessComponent.defaultProps.bar = "bar";\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        import defaults from "./defaults";\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = defaults;\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        import { foo } from "./defaults";\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = foo;\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        const component = rowsOfType(GuestlistEntry, (rowData, ownProps) => ({\n            ...rowData,\n            onPress: () => ownProps.onPress(rowData.id),\n        }));\n      ',
    },
    {
      code: '\n        MyStatelessComponent.propTypes = {\n          ...stuff,\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = {\n        foo: "foo"\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = {\n        ...defaults,\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
    },
    {
      code: '\n        MyStatelessComponent.propTypes = {\n          ...stuff,\n          foo: PropTypes.string\n        };\n        function MyStatelessComponent({ foo = "foo", bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          ...someProps,\n          bar: PropTypes.string.isRequired\n        };\n      ',
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          ...defaults,\n          bar: "bar"\n        };\n      ',
    },
    {
      code: '\n        type Props = {\n          foo: string\n        };\n        class Hello extends React.Component {\n          props: Props;\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ',
    },
    {
      code: '\n        type Props = {\n          foo: string,\n          bar?: string\n        };\n        class Hello extends React.Component {\n          props: Props;\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Hello.defaultProps = {\n          bar: "bar"\n        };\n      ',
    },
    {
      code: '\n        class Hello extends React.Component {\n          props: {\n            foo: string,\n            bar?: string\n          };\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Hello.defaultProps = {\n          bar: "bar"\n        };\n      ',
    },
    {
      code: '\n        class Hello extends React.Component {\n          props: {\n            foo: string\n          };\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ',
    },
    {
      code: '\n        function Hello(props: { foo?: string }) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.defaultProps = { foo: "foo" };\n      ',
    },
    {
      code: '\n        function Hello(props: { foo: string }) {\n          return <div>Hello {foo}</div>;\n        }\n      ',
    },
    {
      code: '\n        const Hello = (props: { foo?: string }) => {\n          return <div>Hello {props.foo}</div>;\n        };\n        Hello.defaultProps = { foo: "foo" };\n      ',
    },
    {
      code: '\n        const Hello = (props: { foo: string }) => {\n          return <div>Hello {foo}</div>;\n        };\n      ',
    },
    {
      code: '\n        const Hello = function(props: { foo?: string }) {\n          return <div>Hello {props.foo}</div>;\n        };\n        Hello.defaultProps = { foo: "foo" };\n      ',
    },
    {
      code: '\n        const Hello = function(props: { foo: string }) {\n          return <div>Hello {foo}</div>;\n        };\n      ',
    },
    {
      code: '\n        type Props = {\n          foo: string,\n          bar?: string\n        };\n        type Props2 = {\n          foo: string,\n          baz?: string\n        }\n        function Hello(props: Props | Props2) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.defaultProps = {\n          bar: "bar",\n          baz: "baz"\n        };\n      ',
    },
    {
      code: '\n        import type Props from "fake";\n        class Hello extends React.Component {\n          props: Props;\n          render () {\n            return <div>Hello {this.props.name.firstname}</div>;\n          }\n        }\n      ',
    },
    {
      code: '\n        type Props = any;\n        const Hello = function({ foo }: Props) {\n          return <div>Hello {foo}</div>;\n        };\n      ',
    },
    {
      code: '\n        import type ImportedProps from "fake";\n        type Props = ImportedProps;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ',
    },
    {
      code: '\n        import type { ImportedType } from "fake";\n        type Props = ImportedType;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ',
    },
    {
      code: '\n        import type ImportedProps from "fake";\n        type NestedProps = ImportedProps;\n        type Props = NestedProps;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ',
    },
    {
      code: '\n        function Hello(props) {\n          return <div>Hello {props.bar}</div>;\n        }\n        Hello.propTypes = {\n          bar: PropTypes.string\n        };\n        Hello.defaultProps = {\n          "bar": "bar"\n        };\n      ',
    },
    {
      code: '\n        class Hello extends React.Component {\n          static propTypes = {\n            foo: PropTypes.string.isRequired\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ',
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
    },
    {
      code: "\n        type Props = {\n          foo?: string,\n          bar?: string\n        }\n        function Hello({ foo = 'asdf', bar = 'qwer' }: Props) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: "\n        type Props = {\n          foo?: string,\n          bar: string\n        }\n        function Hello({ foo = 'asdf', bar }: Props) {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },
    {
      code: '\n        import React from "react";\n\n        interface Props {\n          name: string;\n        }\n\n        const MyComponent: React.FC<Props> = ({ name }) => {\n          return <div>{name}</div>;\n        };\n\n        export default MyComponent;\n      ',
    },
    {
      code: "\n        interface Props {\n          foo?: string,\n          bar?: string\n        }\n        const Hello: React.FC<Props> = ({ foo = 'asdf', bar = 'qwer' }) => {\n          return <div>Hello {foo} and {bar}</div>;\n        }\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 80 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        class Hello extends React.Component<HelloProps> {
          static defaultProps = {
            bar: "bar"
          }
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      options: [{ forbidDefaultForRequired: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        class Hello extends Component<HelloProps> {
          static defaultProps = {
            bar: "bar"
          }
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      options: [{ forbidDefaultForRequired: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 82 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        type HelloState = {
          dummyState: string
        };
        class Hello extends Component<HelloProps, HelloState> {
          static defaultProps = {
            bar: "bar"
          }
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      options: [{ forbidDefaultForRequired: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
      code: `
        type HelloProps = {
          foo?: string,
          bar?: string
        };
        class Hello extends Component<HelloProps> {
          static defaultProps = {
            foo: "foo",
            bar: "bar"
          }
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      options: [{ forbidDefaultForRequired: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          +name?: string,
        };
        function Hello(props: Props) {
          return <div>Hello {props.name}</div>;
        }
        Hello.defaultProps = {
          name: 'foo'
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      errors: [expectShouldHaveDefaultError1([6, 11, 6, 32])],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = forbidExtraProps({\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        });\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectShouldHaveDefaultError1([6, 11, 6, 32])],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        const propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes = forbidExtraProps(propTypes);\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectShouldHaveDefaultError1([6, 11, 6, 32])],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.baz = React.propTypes.string;\n      ',
      errors: [
        expectShouldHaveDefaultError1([6, 11, 6, 32]),
        expectShouldHaveDefaultError2([9, 9, 9, 68]),
      ],
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n      ',
      errors: [expectShouldHaveDefaultError1([3, 11, 3, 32])],
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = defaults;\n      ',
      errors: [expectShouldHaveDefaultError3([10, 11, 10, 32])],
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ',
      errors: [expectShouldHaveDefaultError3([7, 11, 7, 32])],
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        function MyStatelessComponent({ foo, bar = "asdf" }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        diagnostic(
          'shouldAssignObjectDefault',
          'propType "foo" is not required, but has no corresponding default argument value.',
          [6, 41, 6, 44],
        ),
      ],
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n        };\n        function MyStatelessComponent(props) {\n          return <div>{props.foo}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        diagnostic(
          'destructureInSignature',
          'Must destructure props in the function signature to initialize an optional prop.',
          [5, 39, 5, 44],
        ),
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar, ...props }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired,\n          hello: PropTypes.string.isRequired,\n          world: PropTypes.string.isRequired\n        }\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        diagnostic(
          'shouldAssignObjectDefault',
          'propType "foo" is not required, but has no corresponding default argument value.',
          [2, 41, 2, 44],
        ),
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          }\n        });\n      ',
      errors: [expectShouldHaveDefaultError1([7, 13, 7, 34])],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          }\n        });\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [expectShouldHaveDefaultError1([7, 13, 7, 34])],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo"\n            };\n          }\n        });\n      ',
      errors: [expectShouldHaveDefaultError3([8, 13, 8, 34])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      errors: [expectShouldHaveDefaultError1([10, 11, 10, 32])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [expectShouldHaveDefaultError1([10, 11, 10, 32])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
      errors: [expectShouldHaveDefaultError3([11, 11, 11, 32])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n      ',
      errors: [expectShouldHaveDefaultError1([12, 9, 12, 50])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = "foo";\n      ',
      errors: [expectShouldHaveDefaultError3([10, 11, 10, 32])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {};\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.bar = "bar";\n      ',
      errors: [expectShouldHaveDefaultError1([10, 9, 10, 50])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes = props;\n      ',
      errors: [expectShouldHaveDefaultError1([10, 11, 10, 32])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        const defaults = {\n          foo: "foo"\n        };\n        Greeting.propTypes = props;\n        Greeting.defaultProps = defaults;\n      ',
      errors: [expectShouldHaveDefaultError3([11, 11, 11, 32])],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [expectShouldHaveDefaultError4([5, 15, 5, 37])],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [expectShouldHaveDefaultError4([5, 15, 5, 37])],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: PropTypes.string,\n              bar: PropTypes.string\n            };\n          }\n          static get defaultProps() {\n            return {\n              bar: "world"\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([5, 15, 5, 36])],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string\n        };\n        class Hello extends React.Component {\n          static get propTypes() {\n            return props;\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([3, 11, 3, 32])],
    },
    {
      code: '\n        const defaults = {\n          bar: "world"\n        };\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: PropTypes.string,\n              bar: PropTypes.string\n            };\n          }\n          static get defaultProps() {\n            return defaults;\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([8, 15, 8, 36])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          };\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([9, 13, 9, 34])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          };\n        }\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [expectShouldHaveDefaultError1([9, 13, 9, 34])],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          };\n          static defaultProps = {\n            foo: "foo"\n          };\n        }\n      ',
      errors: [expectShouldHaveDefaultError3([10, 13, 10, 34])],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([3, 11, 3, 32])],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        const defaults = {\n          foo: "foo"\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n          static defaultProps = defaults;\n        }\n      ',
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 32])],
    },
    {
      code: '\n        let Greetings = {};\n        Greetings.Hello = class extends React.Component {\n          render () {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Greetings.Hello.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [expectShouldHaveDefaultError1([9, 11, 9, 32])],
    },
    {
      code: '\n        var Greetings = ({ foo = "foo" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        Greetings.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [expectShouldHaveDefaultError1([6, 11, 6, 32])],
    },
    {
      code: '\n        var ComponentWithNoProps = ({ bar = "bar" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        var Greetings = ({ foo = "foo" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        Greetings.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [expectShouldHaveDefaultError1([9, 11, 9, 32])],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static propTypes = {\n            foo: PropTypes.string\n          };\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ',
      errors: [expectShouldHaveDefaultError1([4, 13, 4, 34])],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          static defaultProps() {\n            return {\n              name: 'John'\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ",
      errors: [expectShouldHaveDefaultError4([5, 15, 5, 37])],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              'first-name': PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props['first-name']}</div>;\n          }\n        }\n      ",
      errors: [
        diagnostic(
          'shouldHaveDefault',
          'propType "first-name" is not required, but has no corresponding defaultProps declaration.',
          [5, 15, 5, 45],
        ),
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [expectNoDefaultWithRequiredError([8, 11, 8, 43])],
    },
    {
      code: "\n        function Hello(props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [expectNoDefaultWithRequiredError([6, 11, 6, 43])],
    },
    {
      code: "\n        const Hello = (props) => {\n          return <div>Hello {props.foo}</div>;\n        };\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [expectNoDefaultWithRequiredError([6, 11, 6, 43])],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static propTypes = {\n            foo: PropTypes.string.isRequired\n          }\n          static defaultProps = {\n            foo: 'bar'\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [expectNoDefaultWithRequiredError([4, 13, 4, 45])],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes () {\n            return {\n              foo: PropTypes.string.isRequired\n            };\n          }\n          static get defaultProps() {\n            return {\n              foo: 'bar'\n            };\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [expectNoDefaultWithRequiredError([5, 15, 5, 47])],
    },
    {
      code: "\n        function Hello(props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        diagnostic(
          'noDefaultPropsWithFunction',
          'Don’t use defaultProps with function components.',
          [2, 9, 4, 10],
        ),
      ],
    },
    {
      code: '\n        function Hello(props) {\n          return <div>Hello {props.foo} and {props.bar}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired,\n          bar: PropTypes.string\n        };\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        diagnostic(
          'destructureInSignature',
          'Must destructure props in the function signature to initialize an optional prop.',
          [2, 24, 2, 29],
        ),
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        const MyComponent= (props) => {\n          switch (props.usedProp) {\n            case 1:\n              return (<div />);\n            default:\n              return <div />;\n          }\n        };\n\n        MyComponent.propTypes = {\n          usedProp: PropTypes.string,\n        };\n\n        export default MyComponent;\n      ",
      errors: [
        diagnostic(
          'shouldHaveDefault',
          'propType "usedProp" is not required, but has no corresponding defaultProps declaration.',
          [15, 11, 15, 37],
        ),
      ],
    },
    {
      code: '\n        Foo.propTypes = {\n          a: PropTypes.string,\n        }\n\n        export default function Foo(props) {\n          return <p>{props.a}</p>\n        };\n      ',
      errors: [
        diagnostic(
          'shouldHaveDefault',
          'propType "a" is not required, but has no corresponding defaultProps declaration.',
          [3, 11, 3, 30],
        ),
      ],
    },
    {
      code: '\n        type Props = {\n          history: {\n            push: Function,\n          },\n          match?: {\n            params?: {\n              date?: string\n            },\n          },\n        };\n\n        export const Daily = ({ history, match: {\n          params: {\n            date = moment().toISOString(),\n          } = {},\n        } = {} }: Props) => (\n              <div>\n                <div style={{ textAlign: \'right\', paddingRight: \'25px\' }}>\n                  Show <DatePicker\n                    selected={moment(date)}\n                    className="datetime"\n                    onChange={d => history.push(`./${d.toISOString()}`)}\n                  />\n                </div>\n                <WithData url="/payments/daily" body={{ date: moment(date).toISOString() }}>\n                  <Flashcards date={moment(date)} />\n                </WithData>\n              </div>\n        );\n      ',
      errors: [
        diagnostic(
          'shouldHaveDefault',
          'propType "match" is not required, but has no corresponding defaultProps declaration.',
          [6, 11, 10, 13],
        ),
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        const propTypes = forbidExtraProps({\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        });\n        MyStatelessComponent.propTypes = propTypes;\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [expectShouldHaveDefaultError1([6, 11, 6, 32])],
    },
    {
      code: "\n      function MyStatelessComponent({ foo = 'foo' }) {\n        return <div>{foo}{bar}</div>;\n      }\n      MyStatelessComponent.propTypes = {\n        foo: PropTypes.string.isRequired,\n      };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [expectNoDefaultWithRequiredError([2, 39, 2, 50])],
    },
    {
      code: "\n      function MyStatelessComponent({ foo = 'foo', bar }) {\n        return <div>{foo}{bar}</div>;\n      }\n      MyStatelessComponent.propTypes = {\n        foo: PropTypes.string.isRequired,\n        bar: PropTypes.string\n      };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        expectNoDefaultWithRequiredError([2, 39, 2, 50]),
        diagnostic(
          'shouldAssignObjectDefault',
          'propType "bar" is not required, but has no corresponding default argument value.',
          [2, 52, 2, 55],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            foo?: string,
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
      errors: [expectShouldHaveDefaultError3([5, 13, 5, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
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
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        type Props = {
          foo?: string
        };
        class Hello extends React.Component {
          props: Props;
          static defaultProps: { foo: string };
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectShouldHaveDefaultError1([3, 11, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
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
      `,
      errors: [expectShouldHaveDefaultError3([5, 13, 5, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            foo?: string,
            bar?: string
          };
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [
        expectShouldHaveDefaultError1([4, 13, 4, 26]),
        expectShouldHaveDefaultError3([5, 13, 5, 25]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            foo?: string
          };
          static defaultProps: { foo: string };
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectShouldHaveDefaultError1([4, 13, 4, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        type Props = {
          foo?: string,
          bar?: string
        };
        class Hello extends React.Component {
          props: Props;
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
        Hello.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        type Props = {
          foo?: string,
          bar?: string
        };
        class Hello extends React.Component {
          props: Props;
          static defaultProps: { foo: string, bar: string };
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
        Hello.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            foo?: string,
            bar?: string
          };
          static defaultProps: { foo: string, bar: string };
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
        Hello.defaultProps = {
          foo: "foo"
        };
      `,
      errors: [expectShouldHaveDefaultError3([5, 13, 5, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        function Hello(props: { foo?: string }) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [expectShouldHaveDefaultError1([2, 33, 2, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        function Hello({ foo = "foo" }: { foo?: string }) {
          return <div>Hello {foo}</div>;
        }
      `,
      errors: [expectShouldHaveDefaultError1([2, 43, 2, 55])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        function Hello(props: { foo?: string, bar?: string }) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = { foo: "foo" };
      `,
      errors: [expectShouldHaveDefaultError3([2, 47, 2, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        function Hello(props: { foo?: string, bar?: string }) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [
        expectShouldHaveDefaultError1([2, 33, 2, 46]),
        expectShouldHaveDefaultError3([2, 47, 2, 59]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        type Props = {
          foo?: string
        };
        function Hello(props: Props) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [expectShouldHaveDefaultError1([3, 11, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        const Hello = (props: { foo?: string }) => {
          return <div>Hello {props.foo}</div>;
        };
      `,
      errors: [expectShouldHaveDefaultError1([2, 33, 2, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        const Hello = (props: { foo?: string, bar?: string }) => {
          return <div>Hello {props.foo}</div>;
        };
        Hello.defaultProps = { foo: "foo" };
      `,
      errors: [expectShouldHaveDefaultError3([2, 47, 2, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        const Hello = function(props: { foo?: string }) {
          return <div>Hello {props.foo}</div>;
        };
      `,
      errors: [expectShouldHaveDefaultError1([2, 41, 2, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        const Hello = function(props: { foo?: string, bar?: string }) {
          return <div>Hello {props.foo}</div>;
        };
        Hello.defaultProps = { foo: "foo" };
      `,
      errors: [expectShouldHaveDefaultError3([2, 55, 2, 67])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
        type Props = {
          foo?: string,
          bar?: string
        };
        function Hello(props: Props) {
          return <div>Hello {props.foo}</div>;
        }
        Hello.defaultProps = { foo: "foo" };
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for require-default-props; the native TypeScript annotation path reports 0 diagnostics instead of 2.',
      code: `
        function Hello(props: { one?: string } | { two?: string }) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [
        diagnostic(
          'shouldHaveDefault',
          'propType "one" is not required, but has no corresponding defaultProps declaration.',
          [2, 33, 2, 45],
        ),
        expectShouldHaveDefaultError6([2, 52, 2, 64]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for require-default-props; the native TypeScript annotation path reports 0 diagnostics instead of 2.',
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
      `,
      errors: [
        expectShouldHaveDefaultError3([4, 11, 4, 23]),
        expectShouldHaveDefaultError2([8, 11, 8, 23]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for require-default-props; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
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
          bar: "bar"
        };
      `,
      errors: [expectShouldHaveDefaultError2([8, 11, 8, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for require-default-props; the native TypeScript annotation path reports 0 diagnostics instead of 2.',
      code: `
        type HelloProps = {
          two?: string,
          three: string
        };
        function Hello(props: { one?: string } | HelloProps) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [
        expectShouldHaveDefaultError6([3, 11, 3, 23]),
        diagnostic(
          'shouldHaveDefault',
          'propType "one" is not required, but has no corresponding defaultProps declaration.',
          [6, 33, 6, 45],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      skip: 'Upstream checks Flow prop annotations for require-default-props; the native TypeScript annotation path reports 0 diagnostics instead of 1.',
      code: `
        type HelloProps = {
          two?: string,
          three: string
        };
        function Hello(props: ExternalProps | HelloProps) {
          return <div>Hello {props.foo}</div>;
        }
      `,
      errors: [expectShouldHaveDefaultError6([3, 11, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 68 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        class Hello extends React.Component<HelloProps> {
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 69 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        class Hello extends Component<HelloProps> {
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 70 (7.37.5)',
      code: `
        type HelloProps = {
          foo: string,
          bar?: string
        };
        type HelloState = {
          dummyState: string
        };
        class Hello extends Component<HelloProps, HelloState> {
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [expectShouldHaveDefaultError3([4, 11, 4, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 71 (7.37.5)',
      code: `
        type HelloProps = {
          foo?: string,
          bar?: string
        };
        class Hello extends Component<HelloProps> {
          render() {
            return <div>Hello {this.props.foo}</div>;
          }
        }
      `,
      errors: [
        expectShouldHaveDefaultError1([3, 11, 3, 24]),
        expectShouldHaveDefaultError3([4, 11, 4, 23]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 72 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          +name?: string,
        };
        function Hello(props: Props) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectShouldHaveDefaultError4([3, 11, 3, 25])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
