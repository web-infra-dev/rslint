import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-read-only-props', {} as never, {
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
  ],
  invalid: [
    {
      code: `type Props = { name: string }; function Hello(props: Props) { return <div>{props.name}</div>; }`,
      errors: [
        {
          messageId: 'readOnlyProp',
          message: "Prop 'name' should be read-only.",
        },
      ],
      output: `type Props = { readonly name: string }; function Hello(props: Props) { return <div>{props.name}</div>; }`,
    },
    {
      code: `interface Props { name: string } class Hello extends React.Component<Props> { render() { return <div/>; } }`,
      errors: [
        {
          messageId: 'readOnlyProp',
          message: "Prop 'name' should be read-only.",
        },
      ],
      output: `interface Props { readonly name: string } class Hello extends React.Component<Props> { render() { return <div/>; } }`,
    },
  ],
});
