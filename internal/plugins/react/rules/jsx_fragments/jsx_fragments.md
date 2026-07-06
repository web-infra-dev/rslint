# jsx-fragments

Enforce shorthand or standard form for React fragments.

## Rule Details

In JSX, a React fragment can be written as either
`<React.Fragment>...</React.Fragment>` or as shorthand `<>...</>`.

Examples of **incorrect** code for this rule in the default `syntax` mode:

```jsx
<React.Fragment>
  <Foo />
</React.Fragment>;
```

Examples of **correct** code for this rule in the default `syntax` mode:

```jsx
<>
  <Foo />
</>;
```

```jsx
<React.Fragment key="key">
  <Foo />
</React.Fragment>;
```

## Rule Options

The first option is `"syntax"` (default) or `"element"`.

Examples of **incorrect** code for this rule with `"element"`:

```json
{ "react/jsx-fragments": ["error", "element"] }
```

```jsx
<>
  <Foo />
</>;
```

Examples of **correct** code for this rule with `"element"`:

```jsx
<React.Fragment>
  <Foo />
</React.Fragment>;
```

Support for fragments was added in React v16.2, so the rule reports either form
when an older React version is specified in shared settings.

## Original Documentation

- [react/jsx-fragments](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-fragments.md)
