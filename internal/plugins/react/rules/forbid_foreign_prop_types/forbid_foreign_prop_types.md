# forbid-foreign-prop-types

## Rule Details

This rule forbids using another component's prop types unless they are
explicitly imported / exported. It allows tools such as
[`babel-plugin-transform-react-remove-prop-types`](https://github.com/oliviertassinari/babel-plugin-transform-react-remove-prop-types)
to safely strip `propTypes` from production builds.

To ensure imports are explicitly exported, pair this rule with the
[`named` rule in `eslint-plugin-import`](https://github.com/import-js/eslint-plugin-import/blob/HEAD/docs/rules/named.md).

Examples of **incorrect** code for this rule:

```js
import SomeComponent from './SomeComponent';
SomeComponent.propTypes;

var { propTypes } = SomeComponent;

SomeComponent['propTypes'];
```

Examples of **correct** code for this rule:

```js
import SomeComponent, { propTypes as someComponentPropTypes } from './SomeComponent';
```

## Rule Options

```json
{ "react/forbid-foreign-prop-types": ["error", { "allowInPropTypes": true }] }
```

### `allowInPropTypes`

If `true`, the rule does not report foreign `propTypes` usage when it
appears inside a `propTypes` declaration. The declaration may be either
an assignment to `<Component>.propTypes` or a `static propTypes` class
field.

Examples of **correct** code with `{ "allowInPropTypes": true }`:

```json
{ "react/forbid-foreign-prop-types": ["error", { "allowInPropTypes": true }] }
```

```jsx
const Hello = (props) => <div>{props.name}</div>;
Hello.propTypes = {
  name: Message.propTypes.message,
};

class MyComponent extends React.Component {
  static propTypes = {
    baz: Qux.propTypes.baz,
  };
}
```

## When Not To Use It

This rule helps make removing `propTypes` from production code safe.
Skip it if you don't strip `propTypes` in production. Higher-order
components that hoist a wrapped component's `propTypes` may also need
this rule disabled.

## Original Documentation

- [eslint-plugin-react / forbid-foreign-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/forbid-foreign-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/forbid-foreign-prop-types.js)
