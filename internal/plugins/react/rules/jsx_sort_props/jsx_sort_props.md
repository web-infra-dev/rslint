# jsx-sort-props

Enforce a consistent order for JSX props.

## Rule Details

By default, props are sorted alphabetically and a spread prop starts a new
sortable group. This lets an explicit prop after a spread retain the runtime
override semantics of JSX.

```jsx
// incorrect
<Hello lastName="Smith" firstName="John" />;

// correct
<Hello firstName="John" lastName="Smith" />;
<Hello tel={5555555} {...props} firstName="John" lastName="Smith" />;
```

## Options

The rule accepts one object with these optional properties:

- `callbacksLast`: put `onXxx` callback props after other props.
- `shorthandFirst` / `shorthandLast`: place props without a value first or last.
- `multiline`: use `"ignore"` (default), `"first"`, or `"last"` for props whose
  source text spans multiple lines.
- `ignoreCase`: compare prop names without case distinctions.
- `noSortAlphabetically`: disable alphabetical ordering while retaining the
  other selected ordering constraints.
- `reservedFirst`: put React's reserved props (`children`,
  `dangerouslySetInnerHTML`, `key`, and `ref`) first. It may instead be a
  non-empty subset of that list. `dangerouslySetInnerHTML` is reserved only on
  intrinsic DOM elements.
- `locale`: use `"auto"` (default) or a locale name for locale-aware sorting.

## Original Documentation

- [eslint-plugin-react: jsx-sort-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-sort-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-sort-props.js)
