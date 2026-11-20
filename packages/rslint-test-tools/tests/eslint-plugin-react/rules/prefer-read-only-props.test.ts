import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectReadOnlyPropError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('readOnlyProp', "Prop 'name' should be read-only.", range);
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('prefer-read-only-props', {
  valid: [
    {
      code: `type Props = { readonly name: string }; function Hello(props: Props) { return <div>{props.name}</div>; }`,
    },
    {
      code: `class Hello extends React.Component<{ readonly name: string }> { render() { return <div/>; } }`,
    },
    {
      code: `import React from 'react'; type Props = { readonly name: string }; const Hello: React.FC<Props> = (props) => <div>{props.name}</div>;`,
    },
    {
      code: `const notAComponent = (props: { name: string }) => props.name;`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          +name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        class Hello extends React.Component {
          props: {
            +name: string,
          }

          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        function Hello(props: {+name: string}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type PropsA = {+firstName: string};
        type PropsB = {+lastName: string};
        type Props = PropsA & PropsB;

        function Hello({firstName, lastName}: Props) {
          return <div>Hello {firstName} {lastName}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        const Hello = (props: {+name: string}) => (
          <div>Hello {props.name}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream flow syntax (')' expected.).",
      code: `
        const Hello = ({name}: {+name: string}) => (
          <div>Hello {props.name}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        const notAComponent = (props: {n: number}) => {
          return props.n + 1;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
        Hello.propTypes = {
          name: PropTypes.string,
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        type Props = $ReadOnly<{
          name: string,
        }>

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = $ReadOnly<{
          +firstName: string,
          lastName: string
        }>

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.firstName} {this.props.lastName}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        import React from "react";

        interface Props {
          readonly name: string;
        }

        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };

        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          readonly firstName: string;
          readonly lastName: string;
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          readonly name: string;
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          readonly person: {
            name: string;
          }
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `type Props = { name: string }; function Hello(props: Props) { return <div>{props.name}</div>; }`,
      errors: [expectReadOnlyPropError([1, 16, 1, 28])],
      output: `type Props = { readonly name: string }; function Hello(props: Props) { return <div>{props.name}</div>; }`,
    },
    {
      code: `interface Props { name: string } class Hello extends React.Component<Props> { render() { return <div/>; } }`,
      errors: [expectReadOnlyPropError([1, 19, 1, 31])],
      output: `interface Props { readonly name: string } class Hello extends React.Component<Props> { render() { return <div/>; } }`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        type Props = {
          name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectReadOnlyPropError([3, 11, 3, 24])],
      output: `
        type Props = {
          readonly name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        type Props = {
          -name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectReadOnlyPropError([3, 11, 3, 24])],
      output: `
        type Props = {
          +name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {
            name: string,
          }

          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectReadOnlyPropError([4, 13, 4, 26])],
      output: `
        class Hello extends React.Component {
          props: {
            readonly name: string,
          }

          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        function Hello(props: {name: string}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectReadOnlyPropError([2, 32, 2, 44])],
      output: `
        function Hello(props: {readonly name: string}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        function Hello(props: {|name: string|}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectReadOnlyPropError([2, 33, 2, 45])],
      output: `
        function Hello(props: {|+name: string|}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        function Hello({name}: {name: string}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      errors: [expectReadOnlyPropError([2, 33, 2, 45])],
      output: `
        function Hello({name}: {readonly name: string}) {
          return <div>Hello {props.name}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        type PropsA = {firstName: string};
        type PropsB = {lastName: string};
        type Props = PropsA & PropsB;

        function Hello({firstName, lastName}: Props) {
          return <div>Hello {firstName} {lastName}</div>;
        }
      `,
      errors: [
        diagnostic(
          'readOnlyProp',
          "Prop 'firstName' should be read-only.",
          [2, 24, 2, 41],
        ),
        diagnostic(
          'readOnlyProp',
          "Prop 'lastName' should be read-only.",
          [3, 24, 3, 40],
        ),
      ],
      output: `
        type PropsA = {readonly firstName: string};
        type PropsB = {readonly lastName: string};
        type Props = PropsA & PropsB;

        function Hello({firstName, lastName}: Props) {
          return <div>Hello {firstName} {lastName}</div>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Property or signature expected.).',
      code: `
        const Hello = (props: {-name: string}) => (
          <div>Hello {props.name}</div>
        );
      `,
      errors: [expectReadOnlyPropError([2, 32, 2, 45])],
      output: `
        const Hello = (props: {+name: string}) => (
          <div>Hello {props.name}</div>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        type Props = {
          name: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectReadOnlyPropError([3, 11, 3, 24])],
      output: `
        type Props = {
          readonly name: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        interface Props {
          name: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [expectReadOnlyPropError([3, 11, 3, 24])],
      output: `
        interface Props {
          readonly name: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        import React from "react";
        type Props = {
          name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      errors: [expectReadOnlyPropError([4, 11, 4, 26])],
      output: `
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        type Props = {
          readonly firstName: string;
          lastName: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      errors: [
        diagnostic(
          'readOnlyProp',
          "Prop 'lastName' should be read-only.",
          [4, 11, 4, 28],
        ),
      ],
      output: `
        type Props = {
          readonly firstName: string;
          readonly lastName: string;
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
