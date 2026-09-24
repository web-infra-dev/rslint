# group-exports

## Rule Details

Keep named exports together so a module's public API is easier to find. This rule
reports every declaration in a group that contains multiple exports:

- Local named exports form one group.
- Named re-exports form a separate group for each source string.
- Type exports are grouped separately from value exports.
- CommonJS assignments to `module.exports`, `module.exports.name`, and
  `exports.name` form one group for the file.

Default export declarations and star exports are ignored. An inline type
specifier, such as `export { type Config }`, belongs to the value export group;
use `export type { Config }` for a type export group.

Examples of **incorrect** code for this rule:

```javascript
export const first = true;
export const second = true;
```

```javascript
export { first } from './values';
export { second } from './values';
```

```javascript
module.exports = {};
exports.first = true;
```

Examples of **correct** code for this rule:

```typescript
const first = true;
const second = true;
type Config = { enabled: boolean };

export { first, second };
export type { Config };
```

```javascript
export { first, second } from './values';
export { default as service } from './service';
```

```javascript
module.exports = { first: true, second: true };
```

Deeper assignments such as `module.exports.first.enabled = true` modify an
existing export and do not count as new exports. Like upstream, CommonJS checks
use syntax and also apply to locally declared `module` or `exports` identifiers.

The rule does not provide automatic fixes or suggestions.

## Options

This rule has no options.

## When Not To Use It

Disable this rule if your project prefers exports beside their declarations or
allows CommonJS exports to be assigned in several places.

## Original Documentation

- [eslint-plugin-import: group-exports](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/group-exports.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/group-exports.js)
