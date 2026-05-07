# jsx-max-depth

Enforce JSX maximum depth.

## Rule Details

This rule validates a specific depth for JSX elements.

Examples of **incorrect** code for this rule:

```jsx
<App>
  <Foo>
    <Bar>
      <Baz />
    </Bar>
  </Foo>
</App>
```

## Rule Options

It takes one option object with a positive `max` integer (default: `2`).

```json
{ "react/jsx-max-depth": ["error", { "max": 2 }] }
```

Examples of **incorrect** code for this rule with `{ "max": 1 }`:

```json
{ "react/jsx-max-depth": ["error", { "max": 1 }] }
```

```jsx
<App>
  <Foo>
    <Bar />
  </Foo>
</App>
```

```json
{ "react/jsx-max-depth": ["error", { "max": 1 }] }
```

```jsx
const foobar = (
  <Foo>
    <Bar />
  </Foo>
);
<App>{foobar}</App>;
```

Examples of **incorrect** code for this rule with `{ "max": 2 }`:

```json
{ "react/jsx-max-depth": ["error", { "max": 2 }] }
```

```jsx
<App>
  <Foo>
    <Bar>
      <Baz />
    </Bar>
  </Foo>
</App>
```

Examples of **correct** code for this rule with `{ "max": 1 }`:

```json
{ "react/jsx-max-depth": ["error", { "max": 1 }] }
```

```jsx
<App>
  <Hello />
</App>
```

Examples of **correct** code for this rule with `{ "max": 2 }`:

```json
{ "react/jsx-max-depth": ["error", { "max": 2 }] }
```

```jsx
<App>
  <Foo>
    <Bar />
  </Foo>
</App>
```

Examples of **correct** code for this rule with `{ "max": 3 }`:

```json
{ "react/jsx-max-depth": ["error", { "max": 3 }] }
```

```jsx
<App>
  <Foo>
    <Bar>
      <Baz />
    </Bar>
  </Foo>
</App>
```

## When Not To Use It

If you are not using JSX then you can disable this rule.

## Original Documentation

- [eslint-plugin-react/docs/rules/jsx-max-depth.md](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-max-depth.md)
