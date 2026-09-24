# group-exports

## Rule Details

Keep named exports together so a module's public API is easier to find. This rule
reports every declaration in a group that contains multiple exports:

- Local named exports form one group.
- Named re-exports form a separate group for each source string.
- Type exports are grouped separately from value exports.
- CommonJS assignments to `module.exports`, `module.exports.name`, and
  `exports.name` form one group for the file.

Named export groups are separate for each namespace or declared module body.
CommonJS checks also recognize literal access such as `module["exports"]` and
computed export names such as `exports[name]` or `module.exports[name]`.

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
existing export and do not count as new exports. Assignments to locally declared
`module` or `exports` bindings are also ignored.

The rule does not provide automatic fixes or suggestions.

## Options

This rule has no options.

## When Not To Use It

Disable this rule if your project prefers exports beside their declarations or
allows CommonJS exports to be assigned in several places.

## Differences from upstream

Compared with eslint-plugin-import 2.32.0:

- Exports in different namespace or module bodies do not trigger each other.
  For example, `namespace A { export const a = 1 }` and
  `namespace B { export const b = 2 }` can coexist without a report. Multiple
  exports within either body are still reported.
- Repeated re-exports from `"__proto__"` are reported just like any other source.
  Upstream misses them.
- Dynamic access such as `module[exports]` and private members such as
  `module.#exports` are not treated as CommonJS exports. Literal access such as
  `module["exports"]` or `` module[`exports`] `` does count. Upstream reports the
  first two forms but misses the literal forms.
- Assignments such as `getBox().exports.a = 1` do not count as CommonJS exports.
  Upstream can report them along with other assignments to the returned object.
  CommonJS targets must start at the `module` or `exports` identifier.
- Local bindings named `module` or `exports` are ignored. For example,
  `function setup(exports) { exports.a = 1; exports.b = 2; }` does not export from
  the file and is not reported. Upstream reports both assignments.

## Original Documentation

- [eslint-plugin-import: group-exports](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/group-exports.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/group-exports.js)
