# no-absolute-path

## Rule Details

Disallow absolute module paths, which tie imports to a particular filesystem
layout. The rule checks imports, re-exports, dynamic imports, and direct
`require()` calls with string literals.

Examples of **incorrect** code for this rule:

```javascript
import value from '/project/lib/value.js';
export * from '/project/lib/value.js';
const other = require('/project/lib/other.js');
import('/project/lib/lazy.js');
```

Examples of **correct** code for this rule:

```javascript
import value from './lib/value.js';
export * from 'some-package';
const other = require('../lib/other.js');
import('./lib/lazy.js');
```

The automatic fix replaces the absolute path with a double-quoted path relative
to the importing file. Like upstream, absolute-path detection follows the host
platform: paths starting with `/` are absolute on all platforms, while Windows
drive paths such as `C:/project/lib/value.js` are only reported on Windows.
Fixes use `/` as the path separator on all platforms.

## Options

The rule accepts one object:

| Option     | Default | Meaning                                                                 |
| ---------- | ------- | ----------------------------------------------------------------------- |
| `esmodule` | `true`  | Check imports, re-exports, and dynamic imports.                           |
| `commonjs` | `true`  | Check direct `require()` calls.                                          |
| `amd`      | `false` | Check dependency arrays in two-argument `define()` and `require()` calls. |
| `ignore`   | None    | Skip paths matching any supplied JavaScript regular expression string.   |

For example, `{ "commonjs": false, "amd": true }` checks AMD dependencies while
ignoring CommonJS calls:

```javascript
define(['/project/lib/value.js'], function (value) {}); // Reported
require(['/project/lib/value.js'], function (value) {}); // Reported
const value = require('/project/lib/value.js'); // Ignored
```

For example, `{ "ignore": ["^/generated/"] }` allows absolute paths beginning
with `/generated/`. Patterns are case-sensitive and match the decoded path,
including any escaped characters. When specified, `ignore` must contain at
least one pattern, without duplicate entries.

## Differences from upstream

- Automatic fixes escape `<`, `>`, `&`, U+2028, and U+2029. For example, a path
  containing `<a>` becomes `"./\u003ca\u003e"` instead of `"./<a>"`. The fixed
  paths have the same JavaScript value and refer to the same module.
- Fixes for names beginning with a dot include the relative-path prefix. For
  example, importing `/project/.hidden.js` from `/project/index.js` becomes
  `"./.hidden.js"`. Upstream produces `".hidden.js"`, which module loaders can
  interpret as a package name and fail to load.

## Original Documentation

- [eslint-plugin-import: no-absolute-path](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-absolute-path.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-absolute-path.js)
