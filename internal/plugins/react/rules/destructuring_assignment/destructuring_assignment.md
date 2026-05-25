# destructuring-assignment

Enforce consistent usage of destructuring assignment of props, state, and context.

## Rule Details

This rule has two configurations:

- `"always"` (default): always require destructuring of `props`, `state`, and `context` inside React components.
- `"never"`: forbid destructuring of `props`, `state`, and `context` inside React components.

Examples of **incorrect** code for this rule:

```javascript
const MyComponent = (props) => {
  return <div id={props.id} />;
};
```

```javascript
class Foo extends React.Component {
  render() {
    return <div>{this.props.foo}</div>;
  }
}
```

Examples of **correct** code for this rule:

```javascript
const MyComponent = ({ id }) => (
  <div id={id} />
);
```

```javascript
const MyComponent = (props) => {
  const { id } = props;
  return <div id={id} />;
};
```

```javascript
class Foo extends React.Component {
  render() {
    const { foo } = this.props;
    return <div>{foo}</div>;
  }
}
```

Examples of **incorrect** code for this rule with `"never"`:

```json
{ "react/destructuring-assignment": ["error", "never"] }
```

```javascript
const MyComponent = ({ id, className }) => (
  <div id={id} className={className} />
);
```

```javascript
const Foo = class extends React.PureComponent {
  render() {
    const { foo } = this.props;
    return <div>{foo}</div>;
  }
};
```

Examples of **correct** code for this rule with `"never"`:

```json
{ "react/destructuring-assignment": ["error", "never"] }
```

```javascript
const MyComponent = (props) => (
  <div id={props.id} className={props.className} />
);
```

### `ignoreClassFields`

When this option is set to `true`, accessing `this.props` / `this.state` / `this.context` directly inside a class field initializer is allowed even when the rule is configured as `"always"`. The example below would normally be reported, but is permitted with this option:

```json
{ "react/destructuring-assignment": ["error", "always", { "ignoreClassFields": true }] }
```

```javascript
class Foo extends React.Component {
  bar = this.props.bar;
}
```

### `destructureInSignature`

When this option is set to `"always"`, a destructuring `const { … } = props` whose `props` parameter is not used elsewhere must instead be destructured directly in the function signature. Has no effect when the rule is configured as `"never"` or when the parameter is referenced more than once.

```json
{ "react/destructuring-assignment": ["error", "always", { "destructureInSignature": "always" }] }
```

Examples of **incorrect** code with `destructureInSignature: "always"`:

```javascript
function Foo(props) {
  const {a} = props;
  return <p>{a}</p>;
}
```

Examples of **correct** code with `destructureInSignature: "always"`:

```javascript
function Foo({a}) {
  return <p>{a}</p>;
}
```

```javascript
function Foo(props) {
  const {a} = props;
  return <Goo {...props}>{a}</Goo>;
}
```

## Original Documentation

- [eslint-plugin-react `destructuring-assignment`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/destructuring-assignment.md)
