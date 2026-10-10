# no-deprecated

Reports imports and uses of exported names marked with a JSDoc `@deprecated` tag.
It also checks namespace members, including namespaces re-exported by other modules.
A block with both `@module` and `@deprecated` marks the entire module deprecated,
so even an import with no specifiers is reported.

## Rule Details

Given `answer.js`:

```js
/** @deprecated use the new experiment */
export function multiply(six, nine) {
  return 42;
}

export function experiment() {
  return 54;
}
```

Examples of **incorrect** code:

```js
import { multiply } from './answer.js';
multiply(6, 9);

import * as answer from './answer.js';
answer.multiply(6, 9);
```

Examples of **correct** code:

```js
import { experiment } from './answer.js';
experiment();
```

Deprecated named and default imports are reported on the import specifier and
again when used. Namespace imports are reported on deprecated member names.
Local shadowing prevents reports for the shadowed binding. Computed namespace
access and destructuring are not checked, matching upstream behavior.

## Options

This rule has no options. JSDoc is enabled by default. Enable TomDoc comments
with the shared `import/docstyle` setting:

```js
{
  settings: {
    'import/docstyle': ['jsdoc', 'tomdoc'],
  },
  rules: {
    'import/no-deprecated': 'error',
  },
}
```

TomDoc marks an export with a leading `// Deprecated: reason` comment. Its first
paragraph supplies the reason. Module-level JSDoc is checked independently of
this setting.

## Autofix

This rule has no autofix or suggestions.

## Differences from upstream

- Dependency parser errors are handled by Rslint's compiler rather than emitted
  as `import/no-deprecated` rule reports. Upstream's malformed-dependency test is
  retained as an explained skip.
- A malformed unrelated documentation tag does not suppress a later
  `@deprecated`. For example, `@param {broken` before `@deprecated use newAPI`
  stops upstream's documentation parser before it sees the deprecation; Rslint
  still reports the deprecated export. Correct the malformed tag to obtain the
  same result in both tools.

## References

- [Official rule documentation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-deprecated.md)
- [Upstream implementation](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-deprecated.js)
