# state-in-constructor

## Rule details

Enforce a consistent place to initialize state in React class components.

The rule does not require components to declare state and does not provide automatic fixes.

## Options

- `"always"` (default): Initialize state in the constructor.
- `"never"`: Initialize state with a class field.

```json
{ "react/state-in-constructor": ["error", "always"] }
```

Examples of **incorrect** code for the default `"always"` option:

```jsx
class Counter extends React.Component {
  state = { count: 0 };
}
```

Examples of **correct** code for the default `"always"` option:

```jsx
class Counter extends React.Component {
  constructor(props) {
    super(props);
    this.state = { count: 0 };
  }
}
```

With `"never"`, the constructor assignment above is **incorrect**. Use the class field instead:

```json
{ "react/state-in-constructor": ["error", "never"] }
```

Examples of **correct** code with `"never"`:

```jsx
class Counter extends React.Component {
  state = { count: 0 };
}
```

The `"always"` mode ignores static fields. Both modes apply only to recognized React class components.

## Settings

The rule supports `settings.react.pragma` and `@jsx` annotations when identifying React class components.

## When not to use it

Disable this rule if your project allows both initialization styles.

## Original documentation

- [eslint-plugin-react: state-in-constructor](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/state-in-constructor.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/state-in-constructor.js)
