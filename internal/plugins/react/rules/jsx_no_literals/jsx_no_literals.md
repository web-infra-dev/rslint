# jsx-no-literals

## Rule Details

This rule prevents the use of unwrapped string literals as direct children of JSX elements. By default, JSX text such as `<div>foo</div>` must be wrapped in an expression container (`<div>{'foo'}</div>`). When the `noStrings` option is enabled, even wrapped string literals are forbidden — the rule then expects values to come from translation helpers, identifiers, or other non-literal sources.

Examples of **incorrect** code for this rule:

```jsx
<div>foo</div>
```

```jsx
<div>
  bar
</div>
```

Examples of **correct** code for this rule:

```jsx
<div>{'foo'}</div>
```

```jsx
<div>{translate('greeting')}</div>
```

### Options

```json
{
  "react/jsx-no-literals": [
    "error",
    {
      "noStrings": false,
      "allowedStrings": [],
      "ignoreProps": false,
      "noAttributeStrings": false,
      "restrictedAttributes": [],
      "elementOverrides": {}
    }
  ]
}
```

#### `noStrings`

When `true`, also forbids wrapped string literals (`{'foo'}`, `` {`foo`} ``) inside JSX. Defaults to `false`.

Examples of **incorrect** code for this rule with `{ "noStrings": true }`:

```json
{ "react/jsx-no-literals": ["error", { "noStrings": true }] }
```

```jsx
<Foo>{'Test'}</Foo>
```

#### `allowedStrings`

An array of literal strings that are exempt from the rule. Surrounding whitespace on each entry is trimmed before matching, and matching is performed against the raw and cooked text of the literal.

Examples of **correct** code for this rule with `{ "noStrings": true, "allowedStrings": ["&nbsp;"] }`:

```json
{ "react/jsx-no-literals": ["error", { "noStrings": true, "allowedStrings": ["&nbsp;"] }] }
```

```jsx
<div>&nbsp;</div>
```

#### `ignoreProps`

When `true`, attribute values are exempt from `noStrings`. Defaults to `false`.

#### `noAttributeStrings`

When `true`, string literals used as attribute values are reported. Defaults to `false`.

Examples of **incorrect** code for this rule with `{ "noAttributeStrings": true }`:

```json
{ "react/jsx-no-literals": ["error", { "noAttributeStrings": true }] }
```

```jsx
<img alt="logo" />
```

#### `restrictedAttributes`

An array of attribute names. Listed attributes report on any string literal value, regardless of `noStrings` or `noAttributeStrings`.

Examples of **incorrect** code for this rule with `{ "restrictedAttributes": ["className"] }`:

```json
{ "react/jsx-no-literals": ["error", { "restrictedAttributes": ["className"] }] }
```

```jsx
<div className="card" />
```

#### `elementOverrides`

A map from component name to a per-element override config. The element name must match the regex `^[A-Z][\w.]*$` — HTML element tag names are not supported and silently ignored. The override accepts the same keys as the top-level config plus:

- `allowElement` — when `true`, suppress all reports inside the element.
- `applyToNestedElements` — when `false`, only the closest matching JSX ancestor uses the override; descendant elements fall back to the base config. Defaults to `true`.

When the JSX tag is renamed at import time (`import { T as U } from 'foo'` or `const { T: U } = require('foo')`), the override looks up by the imported name, not the local alias.

Examples of **correct** code for this rule with `{ "elementOverrides": { "Trans": { "allowElement": true } } }`:

```json
{ "react/jsx-no-literals": ["error", { "elementOverrides": { "Trans": { "allowElement": true } } }] }
```

```jsx
<Trans>Welcome back</Trans>
```

## When Not To Use It

If your project does not need to enforce wrapping JSX text in expressions or has no localization workflow, this rule is unnecessary.

## Original Documentation

- [react/jsx-no-literals](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-literals.md)
