// Mirrors the parser-neutral and TypeScript cases from eslint-plugin-react v7.37.5.
// Flow-only cases are retained as explicit skips in the Go upstream suite.
import { RuleTester } from '../rule-tester';
const ruleTester = new RuleTester();
ruleTester.run('require-default-props', {} as never, {
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
  ],
  invalid: [
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = forbidExtraProps({\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        });\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        const propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes = forbidExtraProps(propTypes);\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.baz = React.propTypes.string;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "baz" is not required, but has no corresponding defaultProps declaration.',
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 68,
        },
      ],
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = defaults;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        const defaults = {\n          foo: "foo"\n        };\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        const types = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        function MyStatelessComponent({ foo, bar = "asdf" }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n      ',
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        {
          messageId: 'shouldAssignObjectDefault',
          message:
            'propType "foo" is not required, but has no corresponding default argument value.',
          line: 6,
          column: 41,
          endLine: 6,
          endColumn: 44,
        },
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
        {
          messageId: 'destructureInSignature',
          message:
            'Must destructure props in the function signature to initialize an optional prop.',
          line: 5,
          column: 39,
          endLine: 5,
          endColumn: 44,
        },
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
        {
          messageId: 'shouldAssignObjectDefault',
          message:
            'propType "foo" is not required, but has no corresponding default argument value.',
          line: 2,
          column: 41,
          endLine: 2,
          endColumn: 44,
        },
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          }\n        });\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          }\n        });\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: "foo"\n            };\n          }\n        });\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        Greeting.defaultProps = {\n          foo: "foo"\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 12,
          column: 9,
          endLine: 12,
          endColumn: 50,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: PropTypes.string\n        };\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = "foo";\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {};\n        Greeting.propTypes.foo = PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.bar = "bar";\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 50,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        Greeting.propTypes = props;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        const defaults = {\n          foo: "foo"\n        };\n        Greeting.propTypes = props;\n        Greeting.defaultProps = defaults;\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "name" is not required, but has no corresponding defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 37,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "name" is not required, but has no corresponding defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 37,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: PropTypes.string,\n              bar: PropTypes.string\n            };\n          }\n          static get defaultProps() {\n            return {\n              bar: "world"\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 36,
        },
      ],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string\n        };\n        class Hello extends React.Component {\n          static get propTypes() {\n            return props;\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        const defaults = {\n          bar: "world"\n        };\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: PropTypes.string,\n              bar: PropTypes.string\n            };\n          }\n          static get defaultProps() {\n            return defaults;\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 8,
          column: 15,
          endLine: 8,
          endColumn: 36,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          };\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string.isRequired\n          };\n        }\n      ',
      options: [
        {
          ignoreFunctionalComponents: true,
        },
      ],
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: PropTypes.string,\n            bar: PropTypes.string\n          };\n          static defaultProps = {\n            foo: "foo"\n          };\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 34,
        },
      ],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        const props = {\n          foo: PropTypes.string,\n          bar: PropTypes.string\n        };\n        const defaults = {\n          foo: "foo"\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n          static defaultProps = defaults;\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "bar" is not required, but has no corresponding defaultProps declaration.',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        let Greetings = {};\n        Greetings.Hello = class extends React.Component {\n          render () {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Greetings.Hello.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        var Greetings = ({ foo = "foo" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        Greetings.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        var ComponentWithNoProps = ({ bar = "bar" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        var Greetings = ({ foo = "foo" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        Greetings.propTypes = {\n          foo: PropTypes.string\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 32,
        },
      ],
    },
    {
      code: '\n        class Hello extends React.Component {\n          static propTypes = {\n            foo: PropTypes.string\n          };\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 34,
        },
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: PropTypes.string\n            };\n          }\n          static defaultProps() {\n            return {\n              name: 'John'\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ",
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "name" is not required, but has no corresponding defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 37,
        },
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              'first-name': PropTypes.string\n            };\n          }\n          render() {\n            return <div>Hello {this.props['first-name']}</div>;\n          }\n        }\n      ",
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "first-name" is not required, but has no corresponding defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 45,
        },
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 43,
        },
      ],
    },
    {
      code: "\n        function Hello(props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 43,
        },
      ],
    },
    {
      code: "\n        const Hello = (props) => {\n          return <div>Hello {props.foo}</div>;\n        };\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 43,
        },
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static propTypes = {\n            foo: PropTypes.string.isRequired\n          }\n          static defaultProps = {\n            foo: 'bar'\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 45,
        },
      ],
    },
    {
      code: "\n        class Hello extends React.Component {\n          static get propTypes () {\n            return {\n              foo: PropTypes.string.isRequired\n            };\n          }\n          static get defaultProps() {\n            return {\n              foo: 'bar'\n            };\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ",
      options: [
        {
          forbidDefaultForRequired: true,
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 47,
        },
      ],
    },
    {
      code: "\n        function Hello(props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.propTypes = {\n          foo: PropTypes.string.isRequired\n        };\n        Hello.defaultProps = {\n          foo: 'bar'\n        };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        {
          messageId: 'noDefaultPropsWithFunction',
          message: 'Don’t use defaultProps with function components.',
          line: 2,
          column: 9,
          endLine: 4,
          endColumn: 10,
        },
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
        {
          messageId: 'destructureInSignature',
          message:
            'Must destructure props in the function signature to initialize an optional prop.',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
    {
      code: "\n        import PropTypes from 'prop-types';\n        import React from 'react';\n\n        const MyComponent= (props) => {\n          switch (props.usedProp) {\n            case 1:\n              return (<div />);\n            default:\n              return <div />;\n          }\n        };\n\n        MyComponent.propTypes = {\n          usedProp: PropTypes.string,\n        };\n\n        export default MyComponent;\n      ",
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "usedProp" is not required, but has no corresponding defaultProps declaration.',
          line: 15,
          column: 11,
          endLine: 15,
          endColumn: 37,
        },
      ],
    },
    {
      code: '\n        Foo.propTypes = {\n          a: PropTypes.string,\n        }\n\n        export default function Foo(props) {\n          return <p>{props.a}</p>\n        };\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "a" is not required, but has no corresponding defaultProps declaration.',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 30,
        },
      ],
    },
    {
      code: '\n        type Props = {\n          history: {\n            push: Function,\n          },\n          match?: {\n            params?: {\n              date?: string\n            },\n          },\n        };\n\n        export const Daily = ({ history, match: {\n          params: {\n            date = moment().toISOString(),\n          } = {},\n        } = {} }: Props) => (\n              <div>\n                <div style={{ textAlign: \'right\', paddingRight: \'25px\' }}>\n                  Show <DatePicker\n                    selected={moment(date)}\n                    className="datetime"\n                    onChange={d => history.push(`./${d.toISOString()}`)}\n                  />\n                </div>\n                <WithData url="/payments/daily" body={{ date: moment(date).toISOString() }}>\n                  <Flashcards date={moment(date)} />\n                </WithData>\n              </div>\n        );\n      ',
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "match" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 10,
          endColumn: 13,
        },
      ],
    },
    {
      code: '\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        const propTypes = forbidExtraProps({\n          foo: PropTypes.string,\n          bar: PropTypes.string.isRequired\n        });\n        MyStatelessComponent.propTypes = propTypes;\n      ',
      settings: {
        propWrapperFunctions: ['forbidExtraProps'],
      },
      errors: [
        {
          messageId: 'shouldHaveDefault',
          message:
            'propType "foo" is not required, but has no corresponding defaultProps declaration.',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 32,
        },
      ],
    },
    {
      code: "\n      function MyStatelessComponent({ foo = 'foo' }) {\n        return <div>{foo}{bar}</div>;\n      }\n      MyStatelessComponent.propTypes = {\n        foo: PropTypes.string.isRequired,\n      };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 2,
          column: 39,
          endLine: 2,
          endColumn: 50,
        },
      ],
    },
    {
      code: "\n      function MyStatelessComponent({ foo = 'foo', bar }) {\n        return <div>{foo}{bar}</div>;\n      }\n      MyStatelessComponent.propTypes = {\n        foo: PropTypes.string.isRequired,\n        bar: PropTypes.string\n      };\n      ",
      options: [
        {
          functions: 'defaultArguments',
        },
      ],
      errors: [
        {
          messageId: 'noDefaultWithRequired',
          message:
            'propType "foo" is required and should not have a defaultProps declaration.',
          line: 2,
          column: 39,
          endLine: 2,
          endColumn: 50,
        },
        {
          messageId: 'shouldAssignObjectDefault',
          message:
            'propType "bar" is not required, but has no corresponding default argument value.',
          line: 2,
          column: 52,
          endLine: 2,
          endColumn: 55,
        },
      ],
    },
  ],
});
