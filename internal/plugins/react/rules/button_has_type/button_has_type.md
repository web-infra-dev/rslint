# react/button-has-type

## Rule Details

Forbid `<button>` elements (and `React.createElement('button', ...)` calls) without an explicit `type` attribute. The default DOM `type` for a `button` is `"submit"`, which — when used inside a `<form>` — unexpectedly submits the form. Always specify one of `"button"`, `"submit"`, or `"reset"`.

Examples of **incorrect** code for this rule:

```jsx
<button />
<button type="foo" />
<button type={foo} />
<button type={`button${foo}`} />
<button type={condition ? foo : "button"} />

React.createElement("button")
React.createElement("button", { type: foo })
React.createElement("button", { type: "foo" })
```

Examples of **correct** code for this rule:

```jsx
<button type="button" />
<button type="submit" />
<button type="reset" />
<button type={"button"} />
<button type={`button`} />
<button type={condition ? "button" : "submit"} />

React.createElement("button", { type: "button" })
React.createElement("button", { type: "submit" })
React.createElement("button", { type: "reset" })
```

## Options

The rule takes one optional argument — an object that lets you forbid specific button types. Defaults:

```json
{
  "button": true,
  "submit": true,
  "reset": true
}
```

Setting one of these flags to `false` makes that type value forbidden. For example, with `{ "reset": false }`:

```json
{ "react/button-has-type": ["error", { "reset": false }] }
```

Examples of **incorrect** code with this configuration:

```jsx
<button type="reset" />
<button type={condition ? "reset" : "button"} />
```

Examples of **correct** code with this configuration:

```jsx
<button type="button" />
<button type="submit" />
```

## Limitations

- Detects `<pragma>.createElement(...)` where `<pragma>` defaults to `React` and can be overridden via `settings.react.pragma` (e.g. `"Foo"` → `Foo.createElement(...)`). Destructured `createElement` (e.g. `import { createElement } from 'react'`) and `@jsx` comment pragmas are not supported.

## Original Documentation

- [react/button-has-type](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/button-has-type.md)
