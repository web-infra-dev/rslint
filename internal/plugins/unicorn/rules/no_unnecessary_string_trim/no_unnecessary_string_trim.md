# no-unnecessary-string-trim

## Rule details

This rule replaces `trim()` with `trimStart()` before `startsWith()` and
with `trimEnd()` before `endsWith()` when trimming the other side cannot
affect the result.

A second argument can make the other side of the string observable. So can
trailing whitespace in a prefix or leading whitespace in a suffix. The rule
also skips optional chaining on the call chain, `trim()` calls with arguments,
and search values whose primitive string form cannot be determined statically.
It skips `trim()` on schemas created with a string constructor from `zod`,
including schemas stored in local variables that are never reassigned.

Examples of **incorrect** code for this rule

```javascript
value.trim().startsWith('-');
value.trim().endsWith('-');
```

Examples of **correct** code for this rule

```javascript
value.trimStart().startsWith('-');
value.trimEnd().endsWith('-');
value.trim().startsWith('-', 1);
value.trim().startsWith('foo ');
```

## Options

This rule has no options.

## Differences from upstream

Rslint skips Zod string constructors called through a default import, a named
`string` import, TypeScript's `import z = require('zod')`, or bracket access
such as `z['string']()`. It also skips schemas built with constructor options
or stored in a local variable that is never reassigned. In these cases,
Unicorn v79.0.0 can suggest replacing `trim()` with `trimStart()` or
`trimEnd()`, which would throw when building the schema.

In JavaScript, Rslint also preserves `trim()` when JSDoc identifies the
receiver as a non-string value and it is stored in a `let` variable that is
never reassigned.

With a local type alias named `z`, Unicorn v79.0.0 offers a `trimStart()`
autofix for the schema below. Rslint skips the call because `z.string()`
still uses the Zod import at runtime.

```typescript
import { z } from 'zod';

function createSchema() {
  type z = string;
  return z.string().trim().startsWith('-');
}
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unnecessary-string-trim.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-unnecessary-string-trim.js)
