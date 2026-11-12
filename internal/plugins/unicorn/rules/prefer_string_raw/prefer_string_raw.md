# prefer-string-raw

## Rule Details

Use `String.raw` for quoted strings and untagged templates that would be easier
to read with literal backslashes. The rule also flags a `String.raw` tag when
none of the template's literal parts contain a backslash.

### Incorrect

```javascript
const path = 'C:\\Users\\public';
const pattern = new RegExp('item\\.json');
const file = `C:\\temp\\run-${id}.log`;
const name = String.raw`report`;
```

### Correct

```javascript
const path = String.raw`C:\Users\public`;
const pattern = new RegExp(String.raw`item\.json`);
const file = String.raw`C:\temp\run-${id}.log`;
const name = 'report';
```

## Automatic fixes

In TypeScript, fixes preserve literal types. With type information, the rule
skips conversions that would widen a literal type. Without type information,
it fixes clearly widened bindings such as `let path = 'C:\\temp'` and offers
an editor suggestion when widening is possible. Literals directly inside
`as const` or `<const>` assertions are allowed.

Adding or removing the tag in a template with multiple substitutions can
change when JavaScript converts those substitutions to strings.

## Differences from upstream

- In JavaScript files, `--fix` can convert `const path = 'C:\\temp'` even in a
  project that also checks TypeScript. ESLint can offer an editor suggestion
  for the same conversion.
- Adding `String.raw` is left for manual review when the configured ECMAScript
  version predates tagged templates or when `String` is shadowed or disabled.
  Inside a `with` block, `String.raw` can refer to a custom tag, so tag removal
  is also skipped.
- An unnecessary tag with explicit type arguments, such as
  `` String.raw<string>`abc` ``, is reported for manual removal.
- Removing `String.raw` keeps an ordinary template when a quoted string could
  activate a directive such as `'use strict'` in a script or function.

## Original Documentation

- [eslint-plugin-unicorn: prefer-string-raw](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-string-raw.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-raw.js)
