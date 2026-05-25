# forbid-prop-types

By default this rule prevents vague prop types with more specific alternatives available (`any`, `array`, `object`), but any prop type can be disabled if desired. The defaults are chosen because they have obvious replacements. `any` should be replaced with, well, anything. `array` and `object` can be replaced with `arrayOf` and `shape`, respectively.

## Rule Details

This rule checks all JSX components and verifies that no forbidden propTypes are used. This rule is off by default.

Examples of **incorrect** code for this rule:

```jsx
var Component = createReactClass({
  propTypes: {
    a: PropTypes.any,
    r: PropTypes.array,
    o: PropTypes.object,
  },
});

class Component extends React.Component {
  render() {
    return <div />;
  }
}
Component.propTypes = {
  a: PropTypes.any,
  r: PropTypes.array,
  o: PropTypes.object,
};

class Component extends React.Component {
  static propTypes = {
    a: PropTypes.any,
    r: PropTypes.array,
    o: PropTypes.object,
  };
  render() {
    return <div />;
  }
}
```

Examples of **correct** code for this rule:

```jsx
class Component extends React.Component {
  render() {
    return <div />;
  }
}
Component.propTypes = {
  s: PropTypes.string,
  n: PropTypes.number,
  i: PropTypes.instanceOf(HTMLElement),
  b: PropTypes.bool,
};
```

## Rule Options

```json
{
  "react/forbid-prop-types": [
    "error",
    {
      "forbid": ["any", "array", "object"],
      "checkContextTypes": false,
      "checkChildContextTypes": false
    }
  ]
}
```

### `forbid`

An array of strings, with the names of `PropTypes` keys that are forbidden. The default value for this option is `["any", "array", "object"]`.

Examples of **incorrect** code for this rule with `{ "forbid": ["number"] }`:

```json
{ "react/forbid-prop-types": ["error", { "forbid": ["number"] }] }
```

```jsx
class Component extends React.Component {
  static propTypes = {
    n: PropTypes.number,
  };
}
```

### `checkContextTypes`

Whether or not to check `contextTypes` for forbidden prop types. The default value is `false`.

Examples of **incorrect** code for this rule with `{ "checkContextTypes": true }`:

```json
{ "react/forbid-prop-types": ["error", { "checkContextTypes": true }] }
```

```jsx
class Foo extends React.Component {
  static contextTypes = {
    a: PropTypes.any,
  };
}
```

### `checkChildContextTypes`

Whether or not to check `childContextTypes` for forbidden prop types. The default value is `false`.

Examples of **incorrect** code for this rule with `{ "checkChildContextTypes": true }`:

```json
{ "react/forbid-prop-types": ["error", { "checkChildContextTypes": true }] }
```

```jsx
class Foo extends React.Component {
  static childContextTypes = {
    a: PropTypes.any,
  };
}
```

## When Not To Use It

This rule is a formatting/documenting preference and not following it won't negatively affect the quality of your code. This rule encourages prop types that more specifically document their usage.

## Original Documentation

- [eslint-plugin-react / forbid-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/forbid-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/forbid-prop-types.js)
