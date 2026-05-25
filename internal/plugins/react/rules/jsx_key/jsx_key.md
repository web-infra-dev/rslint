# jsx-key

Disallow missing `key` props in iterators / collection literals.

## Rule Details

Warn if a JSX element that likely needs a `key` prop — namely, one inside an
array literal or returned from an arrow function / function expression passed
to `Array.prototype.map` or `Array.from` — is missing one.

Examples of **incorrect** code for this rule:

```jsx
[<Hello />, <Hello />, <Hello />];
```

```jsx
data.map(x => <Hello>{x}</Hello>);
```

```jsx
Array.from([1, 2, 3], (x) => <Hello>{x}</Hello>);
```

```jsx
<Hello {...{ key: id, id, caption }} />
```

In the last example the key is being spread, which is currently possible but
discouraged in favor of a statically provided key.

Examples of **correct** code for this rule:

```jsx
[<Hello key="first" />, <Hello key="second" />, <Hello key="third" />];
```

```jsx
data.map((x) => <Hello key={x.id}>{x}</Hello>);
```

```jsx
Array.from([1, 2, 3], (x) => <Hello key={x}>{x}</Hello>);
```

```jsx
<Hello key={id} {...{ id, caption }} />
```

## Rule Options

```json
{
  "react/jsx-key": [
    "error",
    {
      "checkFragmentShorthand": false,
      "checkKeyMustBeforeSpread": false,
      "warnOnDuplicates": false
    }
  ]
}
```

### `checkFragmentShorthand` (default: `false`)

When `true`, report the shorthand fragment syntax (`<></>`) used in arrays /
iterators, since shorthand fragments cannot carry a `key`. The reported
suggestion uses the configured pragma, e.g. `React.Fragment` or, with
`settings.react.pragma`/`settings.react.fragment`, `Act.Frag`.

```json
{ "react/jsx-key": ["error", { "checkFragmentShorthand": true }] }
```

```jsx
[<></>, <></>, <></>];
```

```jsx
data.map(x => <>{x}</>);
```

### `checkKeyMustBeforeSpread` (default: `false`)

When `true`, report any `key` prop that appears after a `{...spread}`
attribute. Required by React's new JSX transform.

```json
{ "react/jsx-key": ["error", { "checkKeyMustBeforeSpread": true }] }
```

```jsx
<span {...spread} key="key-after-spread" />;
```

### `warnOnDuplicates` (default: `false`)

When `true`, report keys whose source-text value is identical within the same
container (array literal or JSX parent).

```json
{ "react/jsx-key": ["error", { "warnOnDuplicates": true }] }
```

```jsx
const spans = [
  <span key="notunique" />,
  <span key="notunique" />,
];
```

Key values are compared as raw source text. `key="a"` and `key={'a'}` do NOT
count as duplicates of each other — they are different source strings.

## When Not To Use It

If you are not using JSX then you can disable this rule.

Also, if you frequently use arrow functions that return JSX but never render
the JSX inside an iterable, you may want to disable this rule.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-key.md
