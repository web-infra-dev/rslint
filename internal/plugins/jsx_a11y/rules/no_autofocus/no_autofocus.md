# no-autofocus

Enforce that the `autoFocus` prop is not enabled on JSX elements.
Programmatically moving focus on render is disorienting for sighted users
(the viewport jumps to an unexpected location) and disruptive for
assistive-technology users (screen-reader and keyboard navigation are
reset).

## Rule Details

The rule fires on a JSX attribute named `autoFocus` (case-sensitive — the
lowercase HTML attribute `autofocus` is **not** matched) whose value is
anything other than the JS boolean `false` or the literal string `"false"`.
The boolean attribute form (`<div autoFocus />`) is treated as `true` and is
reported.

Examples of **incorrect** code for this rule:

```jsx
<div autoFocus />
<div autoFocus={true} />
<div autoFocus={undefined} />
<div autoFocus="true" />
<input autoFocus />
<Foo autoFocus />
```

Examples of **correct** code for this rule:

```jsx
<div />
<div autofocus />
<input autofocus="true" />
<Foo bar />
<div autoFocus={false} />
<div autoFocus="false" />
```

## Rule Options

### `ignoreNonDOM`

Type: `boolean`. Default: `false`.

When `true`, custom components (anything not in the HTML DOM element set)
are skipped. The element name is resolved through
`settings['jsx-a11y'].polymorphicPropName` and
`settings['jsx-a11y'].components` before the DOM-set lookup, so a custom
component remapped to a DOM tag is still checked.

Examples of **correct** code with `{ "ignoreNonDOM": true }`:

```json
{ "jsx-a11y/no-autofocus": ["error", { "ignoreNonDOM": true }] }
```

```jsx
<Foo autoFocus />
<div><div autofocus /></div>
```

Examples of **incorrect** code with `{ "ignoreNonDOM": true }`:

```json
{ "jsx-a11y/no-autofocus": ["error", { "ignoreNonDOM": true }] }
```

```jsx
<input autoFocus />
```

## Differences from upstream

Unlike `eslint-plugin-jsx-a11y` v6.10.2, Rslint allows `autoFocus` values
that its static evaluator resolves to the boolean `false` or the string
`"false"`. For example, `<input autoFocus={false} />` is accepted by Rslint
but reported by upstream, which checks only whether the prop is present.
This is intentional: Rslint allows an explicitly disabled value instead of
requiring the prop to be removed. The diagnostic therefore says the prop
should not be "enabled", rather than upstream's "used".

This is not a general truthiness check: `autoFocus={undefined}`,
`autoFocus={null}` and `autoFocus={0}` are still reported. Values that the
static evaluator cannot resolve to either accepted value are also reported.
To satisfy both tools, remove the `autoFocus` prop when it is disabled.

## Original Documentation

- [eslint-plugin-jsx-a11y: no-autofocus](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/v6.10.2/docs/rules/no-autofocus.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/v6.10.2/src/rules/no-autofocus.js)
