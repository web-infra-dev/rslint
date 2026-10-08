# dynamic-import-chunkname

## Rule Details

Reports dynamic imports that do not have a webpack magic comment naming the chunk. When a chunk is not named, the bundler generates a name that changes between builds, which prevents long-term browser caching and makes bundle size reports harder to read.

The rule checks `import()` by default. Use `importFunctions` to check calls to other functions as well.

Each comment before the first argument must be a block comment padded with spaces on both sides, such as `/* webpackChunkName: "name" */`, and must contain only webpack magic comment settings. At least one comment must name the chunk. Rspack's `rspack`-prefixed magic comments are accepted wherever the `webpack`-prefixed ones are.

Examples of **incorrect** code for this rule:

```javascript
// no leading comment
import('someModule');

// incorrectly formatted comment
import(
  /*webpackChunkName:"someModule"*/
  'someModule',
);
import(
  /* webpackChunkName : "someModule" */
  'someModule',
);

// invalid syntax for webpack comment
import(
  /* totally not webpackChunkName: "someModule" */
  'someModule',
);

// single-line comment, not a block comment
import(
  // webpackChunkName: "someModule"
  'someModule',
);

// chunk names are not needed when eager mode is set
import(
  /* webpackMode: "eager" */
  /* webpackChunkName: "someModule" */
  'someModule',
);
```

Examples of **correct** code for this rule:

```javascript
import(
  /* webpackChunkName: "someModule" */
  'someModule',
);
import(
  /* webpackChunkName: "someModule" */
  /* webpackPrefetch: true */
  'someModule',
);
import(
  /* webpackChunkName: "someModule", webpackPrefetch: true */
  'someModule',
);

// using single quotes instead of double quotes
import(
  /* webpackChunkName: 'someModule' */
  'someModule',
);

// Rspack's prefix
import(
  /* rspackChunkName: "someModule" */
  'someModule',
);
```

### Recognized magic comments

| `webpack` prefix       | `rspack` prefix       | Accepted value                                                                  |
| ---------------------- | --------------------- | ------------------------------------------------------------------------------- |
| `webpackChunkName`     | `rspackChunkName`     | any text; a chunk name must also match `webpackChunknameFormat`                 |
| `webpackPrefetch`      | `rspackPrefetch`      | `true`, `false` or an integer                                                   |
| `webpackPreload`       | `rspackPreload`       | `true`, `false` or an integer                                                   |
| `webpackIgnore`        | `rspackIgnore`        | `true` or `false`                                                               |
| `webpackInclude`       | `rspackInclude`       | a regular expression literal                                                    |
| `webpackExclude`       | `rspackExclude`       | a regular expression literal                                                    |
| `webpackMode`          | `rspackMode`          | `"lazy"`, `"lazy-once"`, `"eager"` or `"weak"`                                  |
| `webpackExports`       | `rspackExports`       | a quoted export name or an array of quoted export names                         |
| `webpackFetchPriority` | `rspackFetchPriority` | `"low"`, `"high"` or `"auto"`                                                   |

The two prefixes can be mixed in the same file and in the same comment.

### Eager mode

A chunk is not created for an import in `eager` mode, so a chunk name next to `webpackMode: "eager"` (or `rspackMode: "eager"`) is reported. The report has two suggestions: remove the chunk name, or remove the mode. A comment that becomes empty is deleted; the surrounding whitespace stays in place.

## Options

```json
{
  "import/dynamic-import-chunkname": [
    "error",
    {
      "importFunctions": ["dynamicImport"],
      "webpackChunknameFormat": "[a-zA-Z0-57-9-/_]+",
      "allowEmpty": false
    }
  ]
}
```

### `importFunctions`

An array of function names whose first argument is checked in the same way as the argument of `import()`. The default is `[]`.

### `webpackChunknameFormat`

A regular expression source that a chunk name must match. The default is `([0-9a-zA-Z-_/.]|\[(request|index)\])+`. With the configuration above, a chunk name containing the digit `6` is reported:

```javascript
import(
  /* webpackChunkName: "someModule6" */
  'someModule',
);
```

The pattern is matched between the quotes of the chunk name, so anchors such as `^` and `$` never match. If the pattern is not a valid regular expression, the rule reports nothing.

### `allowEmpty`

When `true`, imports without a leading comment, or without a chunk name, are allowed. Comments that are present must still be well formed. The default is `false`.

Examples of **correct** code with `{ "allowEmpty": true }`:

```javascript
import('someModule');

import(
  /* webpackChunkName: "someModule" */
  'someModule',
);
```

Examples of **incorrect** code with `{ "allowEmpty": true }`:

```javascript
// incorrectly formatted comment
import(
  /*webpackChunkName:"someModule"*/
  'someModule',
);
```

## When Not To Use It

If you do not care that webpack may generate chunk names that change between builds, and may invalidate browser caches and bundle size reports.

## Differences from upstream

- Every magic comment accepts the Rspack `rspack` prefix as well as the `webpack` prefix, and `webpackFetchPriority` / `rspackFetchPriority` is recognized. Upstream recognizes only the `webpack` prefix and reports `/* rspackChunkName: "someModule" */` as an invalid comment, and also `/* webpackFetchPriority: "high" */`. The diagnostics themselves keep the upstream wording, which names `webpackChunkName`.
- Upstream checks that a comment is valid by running it with `vm.runInNewContext`. rslint does not run code: it parses the same wrapper and detects syntax errors, TypeScript-only syntax such as `"a" as string`, and references to identifiers that do not exist in a fresh JavaScript context, such as `/* webpackChunkName: someModule */`. Errors that only occur while running other code are not detected. For example, with `{ "allowEmpty": true }`, `/* webpackChunkName: (() => missing)() */` is reported as invalid syntax by upstream and is not reported by rslint; without `allowEmpty`, rslint reports that no chunk name was found instead.
- A call without arguments, such as `dynamicImport()` for a function listed in `importFunctions`, is ignored. Upstream throws a `TypeError` for it.
- An invalid `webpackChunknameFormat` produces no diagnostics. Upstream fails the lint run.

## Original Documentation

- [eslint-plugin-import: dynamic-import-chunkname](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/dynamic-import-chunkname.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/dynamic-import-chunkname.js)
- [Rspack: magic comments](https://rspack.rs/api/runtime-api/module-methods)
