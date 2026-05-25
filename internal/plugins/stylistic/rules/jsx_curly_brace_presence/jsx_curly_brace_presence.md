# jsx-curly-brace-presence

Disallow unnecessary JSX expressions when literals alone are sufficient, or enforce JSX expressions on literals in JSX children or attributes.

## Rule Details

By default, the rule warns about unnecessary curly braces in both JSX props and
children. Prop values that are JSX elements are ignored by default.

Examples of **incorrect** code for this rule:

```javascript
<App prop={'foo'} attr={"bar"}>{'Hello world'}</App>;
```

Examples of **correct** code for this rule:

```javascript
<App prop="foo" attr="bar">Hello world</App>;
```

Examples of **incorrect** code for this rule with `{ "props": "always", "children": "always" }`:

```json
{ "@stylistic/jsx-curly-brace-presence": ["error", { "props": "always", "children": "always" }] }
```

```javascript
<App>Hello world</App>;
<App prop='Hello world'>{'Hello world'}</App>;
```

Examples of **incorrect** code for this rule with `{ "props": "always", "children": "always", "propElementValues": "always" }`:

```json
{ "@stylistic/jsx-curly-brace-presence": ["error", { "props": "always", "children": "always", "propElementValues": "always" }] }
```

```javascript
<App prop=<div /> />;
```

Examples of **incorrect** code for this rule with `{ "props": "never", "children": "never", "propElementValues": "never" }`:

```json
{ "@stylistic/jsx-curly-brace-presence": ["error", { "props": "never", "children": "never", "propElementValues": "never" }] }
```

```javascript
<App prop={<div />} />;
```

## Options

The rule accepts either an options object or a single string shorthand:

```json
{ "@stylistic/jsx-curly-brace-presence": ["error", { "props": "never", "children": "never", "propElementValues": "ignore" }] }
```

```json
{ "@stylistic/jsx-curly-brace-presence": ["error", "never"] }
```

Each of `props`, `children`, and `propElementValues` accepts:

- `"always"` — enforce curly braces.
- `"never"` — disallow unnecessary curly braces.
- `"ignore"` — disable the check.

The string shorthand sets `props` and `children` to the given value;
`propElementValues` stays at its default of `"ignore"`.

Under `"never"`, an attribute string value that contains a quote character is
left wrapped — for example, `<App prop={'say "hi"'} />` and `<Foo bar={"'"} />`
are not reported. Quote-bearing children literals are still reported.

## Original Documentation

- [@stylistic/jsx-curly-brace-presence](https://eslint.style/rules/jsx-curly-brace-presence)
