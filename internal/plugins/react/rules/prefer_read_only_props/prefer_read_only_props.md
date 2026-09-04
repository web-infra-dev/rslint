# prefer-read-only-props

Require React component props to be declared as read-only.

## Rule Details

This rule checks TypeScript properties used as React component props and reports
properties that are missing the `readonly` modifier. It checks props declared
in component type arguments, class `props` fields, and function component
parameter types.

Examples of **incorrect** code for this rule:

```tsx
type Props = { name: string };

function Hello(props: Props) {
  return <div>{props.name}</div>;
}
```

```tsx
interface Props {
  name: string;
}

class Hello extends React.Component<Props> {
  render() {
    return <div>{this.props.name}</div>;
  }
}
```

Examples of **correct** code for this rule:

```tsx
type Props = { readonly name: string };

function Hello(props: Props) {
  return <div>{props.name}</div>;
}
```

```tsx
class Hello extends React.Component<{ readonly name: string }> {
  render() {
    return <div>{this.props.name}</div>;
  }
}
```

## Differences from ESLint

- rslint analyzes TypeScript prop declarations. Referenced prop types are
  resolved only from declarations in the current source file.

## Original Documentation

- [eslint-plugin-react: prefer-read-only-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/prefer-read-only-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/prefer-read-only-props.js)
