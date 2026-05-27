# checked-requires-onchange-or-readonly

## Rule Details

Enforce using `onChange` or `readOnly` attribute when `checked` is used on an `input` element (and on `React.createElement('input', ...)` calls).

A controlled `<input>` whose `checked` value is fixed without an `onChange` handler can never be toggled by the user and triggers a React warning. This rule also forbids combining `checked` with `defaultChecked`, since an input is either controlled (`checked`) or uncontrolled (`defaultChecked`), never both.

The check applies to every `input` element carrying `checked`, regardless of its `type`.

Examples of **incorrect** code for this rule:

```jsx
<input type="checkbox" checked />
<input type="radio" checked={true} />
<input type="checkbox" checked={condition ? true : false} />
<input type="checkbox" checked defaultChecked />

React.createElement('input', { checked: false })
React.createElement('input', { checked: true, defaultChecked: true })
```

Examples of **correct** code for this rule:

```jsx
<input type="checkbox" checked onChange={() => {}} />
<input type="checkbox" checked readOnly />
<input type="checkbox" defaultChecked />

React.createElement('input', { checked: true, onChange: noop })
React.createElement('input', { checked: true, readOnly: true })
```

## Rule Options

```json
{ "react/checked-requires-onchange-or-readonly": ["error", { "ignoreMissingProperties": false, "ignoreExclusiveCheckedAttribute": false }] }
```

### ignoreMissingProperties

Default: `false`. When `true`, an `input` with `checked` but no `onChange` / `readOnly` is allowed (the `missingProperty` diagnostic is suppressed).

Examples of **correct** code for this rule with `{ "ignoreMissingProperties": true }`:

```json
{ "react/checked-requires-onchange-or-readonly": ["error", { "ignoreMissingProperties": true }] }
```

```jsx
<input type="checkbox" checked />
<input type="checkbox" checked={true} />
```

### ignoreExclusiveCheckedAttribute

Default: `false`. When `true`, combining `checked` and `defaultChecked` is allowed (the `exclusiveCheckedAttribute` diagnostic is suppressed).

Examples of **correct** code for this rule with `{ "ignoreExclusiveCheckedAttribute": true }`:

```json
{ "react/checked-requires-onchange-or-readonly": ["error", { "ignoreExclusiveCheckedAttribute": true }] }
```

```jsx
<input type="checkbox" onChange={noop} checked defaultChecked />
```

## Limitations

- Detects `<pragma>.createElement(...)` where `<pragma>` defaults to `React` and can be overridden via `settings.react.pragma` (e.g. `"Foo"` → `Foo.createElement(...)`). Destructured `createElement` (e.g. `import { createElement } from 'react'`) and `@jsx` comment pragmas are not supported.

## Original Documentation

- [react/checked-requires-onchange-or-readonly](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/checked-requires-onchange-or-readonly.md)
