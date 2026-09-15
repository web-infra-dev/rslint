# jsx-one-expression-per-line

## Rule Details

Require one JSX element or expression per line. The fixer inserts line breaks
between expressions that share a line and preserves meaningful spaces with an
explicit `{' '}` expression.

Examples of **incorrect** code for this rule:

```jsx
<App><Hello /></App>

<App>
  <Hello /> World
</App>

<App>
  {"Hello"} {"World"}
</App>
```

Examples of **correct** code for this rule:

```jsx
<App>
  <Hello />
</App>

<App>
  <Hello />
  World
</App>

<App>
  {"Hello"}
  {"World"}
</App>
```

## Rule Options

The rule accepts an optional object with an `allow` property:

```json
{ "react/jsx-one-expression-per-line": ["error", { "allow": "none" }] }
```

### `allow`

Defaults to `"none"`.

- `"none"`: Require every child expression or element to be on its own line.
- `"literal"`: Allow a single literal or JSX text child on the same line.
- `"single-child"`: Allow any single child on the same line.
- `"non-jsx"`: Allow children when there is no direct JSX element or fragment child.

The `"literal"` and `"single-child"` checks apply to the lines occupied by the JSX
itself. Line breaks and comments before its opening tag do not affect the exception.

Examples of **correct** code with `"literal"`:

```jsx
<App>Hello</App>

const render = () => (
  <App>Hello</App>
);
```

Examples of **correct** code with `"single-child"`:

```jsx
<App>Hello</App>
<App>{"Hello"}</App>
<App><Hello /></App>
```

Examples of **correct** code with `"non-jsx"`:

```jsx
<App>Hello {someVariable}</App>
<App>Hello {<Hello />} there!</App>
```

## Differences from ESLint

- For member-expression tags, rslint uses the complete authored tag name in
  diagnostics, such as `Foo.Bar`. ESLint's v7.37.5 implementation reads a
  missing `name` property and reports `undefined` instead.
- For namespaced tags, rslint uses the complete namespace-qualified name,
  such as `svg:path`. ESLint's v7.37.5 implementation coerces the AST object
  and reports `[object Object]` instead. These readable descriptors are an
  intentional divergence from those upstream artifacts.

## Original Documentation

- [eslint-plugin-react: jsx-one-expression-per-line](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-one-expression-per-line.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-one-expression-per-line.js)
