# static-property-placement

Enforce where React component static properties are declared in an ES6 class.

The rule checks `childContextTypes`, `contextTypes`, `contextType`,
`defaultProps` (including legacy `getDefaultProps`), `displayName`, and
`propTypes`. The default placement is `static public field`.

Component detection follows `eslint-plugin-react`, including `settings.react.pragma`,
file-level `@jsx` comments, and adjacent JSDoc `@extends React.Component` or
`@augments React.PureComponent` markers. For class expressions assigned to a
variable, put the JSDoc before the variable declaration:

```tsx
/** @extends React.Component */
const Component = class {
  static propTypes = {};
};
```

## Options

The first option selects the default placement:

- `static public field` — `static propTypes = {...}`
- `static getter` — `static get propTypes() {...}`
- `property assignment` — `MyComponent.propTypes = {...}` outside the class

The second option is an object that overrides the placement for individual
properties. Unspecified properties use the first option.

```json
["property assignment", {"displayName": "static public field"}]
```

## Original Documentation

- [eslint-plugin-react: static-property-placement](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/static-property-placement.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/static-property-placement.js)
