# jsx-sort-props

Enforce a consistent order for JSX props.

## Rule Details

This rule sorts JSX props alphabetically by default. A spread prop starts a new
sortable group, so explicitly written props are never moved across a spread and
their runtime override behavior is preserved. The default comparison is
case-sensitive.

Examples of **incorrect** code for this rule:

```jsx
<Hello lastName="Smith" firstName="John" />;
```

Examples of **correct** code for this rule:

```jsx
<Hello firstName="John" lastName="Smith" />;
<Hello tel={5555555} {...props} firstName="John" lastName="Smith" />;
```

## Options

The rule accepts one options object. Ordering constraints apply in this order:
reserved props, callbacks, shorthand props, multiline props, then alphabetical
sorting.

### `ignoreCase`

When `true`, prop names are compared without case distinctions.

```json
{ "react/jsx-sort-props": ["error", { "ignoreCase": true }] }
```

```jsx
<Hello name="John" Number="2" />;
```

### `callbacksLast`

When `true`, callback props whose names start with `on` followed by an uppercase
letter must appear after other props. This takes precedence over shorthand and
multiline ordering.

```json
{ "react/jsx-sort-props": ["error", { "callbacksLast": true }] }
```

```jsx
<Hello name="John" tel={5555555} onClick={handleClick} />;
```

### `shorthandFirst`

When `true`, shorthand props without a value appear before props with values.
They remain alphabetically sorted within each group.

```json
{ "react/jsx-sort-props": ["error", { "shorthandFirst": true }] }
```

```jsx
<Hello active validate name="John" tel={5555555} />;
```

### `shorthandLast`

When `true`, shorthand props without a value appear after props with values.
They remain alphabetically sorted within each group.

```json
{ "react/jsx-sort-props": ["error", { "shorthandLast": true }] }
```

```jsx
<Hello name="John" tel={5555555} active validate />;
```

### `multiline`

Use `"ignore"` (the default), `"first"`, or `"last"` to control props whose
source text spans multiple lines. Callback and shorthand constraints take
precedence over multiline ordering.

```json
{ "react/jsx-sort-props": ["error", { "multiline": "first" }] }
```

```jsx
<Hello
  classes={{
    greeting: "hello",
  }}
  name="John"
  tel={5555555}
/>
```

### `noSortAlphabetically`

When `true`, alphabetical ordering is disabled while the other selected
constraints still apply.

```json
{ "react/jsx-sort-props": ["error", { "noSortAlphabetically": true }] }
```

```jsx
<Hello tel={5555555} name="John" />;
```

### `reservedFirst`

Set this to `true` to place React's reserved props (`children`,
`dangerouslySetInnerHTML`, `key`, and `ref`) before other props. The
`dangerouslySetInnerHTML` prop is reserved only on intrinsic DOM elements.

```json
{ "react/jsx-sort-props": ["error", { "reservedFirst": true }] }
```

```jsx
<Hello key={id} ref={helloRef} name="John" />;
<div dangerouslySetInnerHTML={{ __html: html }} ref={divRef} />;
```

You can instead provide a non-empty subset of the reserved prop list.

```json
{ "react/jsx-sort-props": ["error", { "reservedFirst": ["key"] }] }
```

```jsx
<Hello key={id} name="John" ref={helloRef} />;
```

### `locale`

`"auto"` (the default) uses rslint's default collation for locale-aware
comparisons when `ignoreCase` is enabled. Provide a locale name to use
locale-aware ordering even without `ignoreCase`. Unicode collation extensions
used by `Intl.Collator` (`co`, `kf`, and `kn`) are accepted; unsupported
collations use the locale's default ordering. Nordic locales order ASCII prop
names like Node 24, including the `aa` contraction and locale-specific case
order.

```json
{ "react/jsx-sort-props": ["error", { "locale": "de" }] }
```

```jsx
<Hello ä="a-umlaut" z="zee" />;
```

## Differences from ESLint

- With `locale: "auto"`, rslint uses a deterministic default collation instead
  of the host environment's locale. Set an explicit locale such as `"de"` when
  locale-specific ordering is required.
- With an explicit `locale`, rslint can order some Unicode or mixed-case prop
  names differently from ESLint.
- With a malformed `locale`, rslint continues linting with its default order;
  ESLint stops with a `RangeError`.
- If sorting a comment-attached block would reverse duplicate props, rslint
  reports the order without a fix; ESLint can change which prop value wins.
- rslint accepts only strings in a `reservedFirst` array. ESLint accepts other
  values and reports them as invalid rule options during linting; rslint rejects
  those configurations during schema validation.

## When Not To Use It

This rule is a formatting preference. If alphabetical JSX prop ordering is not
part of your project's style, leave it disabled.

## Original Documentation

- [eslint-plugin-react: jsx-sort-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-sort-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-sort-props.js)
