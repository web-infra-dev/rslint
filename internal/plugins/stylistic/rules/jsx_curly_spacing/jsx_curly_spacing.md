# jsx-curly-spacing

Enforce or disallow spaces inside of curly braces in JSX attributes and expressions.

## Rule Details

This rule enforces a consistent style for the whitespace immediately inside JSX `{...}` braces — both when the braces wrap an attribute value and (optionally) when they wrap a child expression. It also handles the brace-spacing of `JSXSpreadAttribute` (`{...obj}`) the same way as a regular attribute brace.

By default the rule:

- Checks attribute braces (`attributes` defaults to `true`).
- Does **not** check child braces (`children` defaults to `false`).
- Disallows spaces inside the braces (`when` defaults to `"never"`).
- Allows multi-line content inside the braces (`allowMultiline` defaults to `true`).

## Options

The rule accepts either a string shorthand (`"never"` / `"always"`) or an options object. When using the shorthand, the options object may be supplied as the second argument:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "never", { "allowMultiline": false }] }
```

The options object supports:

- `when`: `"never"` (default) or `"always"` — whether spaces are required or disallowed inside braces.
- `allowMultiline`: `true` (default) or `false` — whether the contents may span multiple lines without a same-line space match.
- `spacing`: object with `"objectLiterals": "never" | "always"` — overrides the spacing rule when the contents are an immediate object literal (`{{ ... }}`).
- `attributes`: `true` (default) / `false` / `{ when?, allowMultiline?, spacing? }` — toggles or overrides the configuration for attribute braces.
- `children`: `true` / `false` (default) / `{ when?, allowMultiline?, spacing? }` — toggles or overrides the configuration for child braces.

Examples of **incorrect** code for this rule with the default `"never"` option:

```javascript
<App foo={ bar } />;
<App foo={ { bar: true, baz: true } } />;
```

Examples of **correct** code for this rule with the default `"never"` option:

```javascript
<App foo={bar} />;
<App foo={{ bar: true, baz: true }} />;
<App foo={
  bar
} />;
```

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "always"] }
```

```javascript
<App foo={bar} />;
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "always"] }
```

```javascript
<App foo={ bar } />;
```

Examples of **incorrect** code for this rule with the `{ "allowMultiline": false }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "never", { "allowMultiline": false }] }
```

```javascript
<App foo={
  bar
} />;
```

Examples of **correct** code for this rule with the `{ "allowMultiline": false }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "never", { "allowMultiline": false }] }
```

```javascript
<App foo={bar} />;
```

Examples of **incorrect** code for this rule with the `{ "spacing": { "objectLiterals": "always" } }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "never", { "spacing": { "objectLiterals": "always" } }] }
```

```javascript
<App foo={{ a: 1 }} />;
```

Examples of **correct** code for this rule with the `{ "spacing": { "objectLiterals": "always" } }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", "never", { "spacing": { "objectLiterals": "always" } }] }
```

```javascript
<App foo={ { a: 1 } } />;
```

Examples of **incorrect** code for this rule with the `{ "children": true }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", { "children": true, "when": "never" }] }
```

```javascript
<App>{ bar }</App>;
```

Examples of **correct** code for this rule with the `{ "children": true }` option:

```json
{ "@stylistic/jsx-curly-spacing": ["error", { "children": true, "when": "never" }] }
```

```javascript
<App>{bar}</App>;
```

## Original Documentation

- [@stylistic/jsx-curly-spacing](https://eslint.style/rules/jsx-curly-spacing)
