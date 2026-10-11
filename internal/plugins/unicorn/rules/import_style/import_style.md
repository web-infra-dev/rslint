# import-style

Enforces specific import styles for configured modules.

## Rule Details

Some modules expose unrelated utilities that are clearer as named imports,
while modules whose exports form one cohesive API can be clearer as default
imports. This rule checks only modules listed in `styles` plus its built-in
defaults.

The four recognized import styles are:

- `unassigned`: `import 'foo'` or `require('foo')`
- `default`: `import value from 'foo'` or `const value = require('foo')`
- `namespace`: `import * as value from 'foo'` or `const value = require('foo')`
- `named`: `import { value } from 'foo'` or `const { value } = require('foo')`

Examples of **incorrect** code with the default options:

```javascript
const util = require('node:util');
import util from 'node:util';
import * as path from 'node:path';
```

Examples of **correct** code with the default options:

```javascript
const { promisify } = require('node:util');
import { promisify } from 'node:util';
import path from 'node:path';
```

## Options

### `styles`

An object mapping module names to allowed import styles. The default styles are
`default` for `chalk` and `path`, and `named` for `util`. A `node:` prefix is
ignored for matching, so configure `util` to control both `util` and
`node:util`.

Set a module to `false` to remove its restrictions. Do not set all four known
styles to `false` to ban a module; use `no-restricted-imports` for that purpose.

```json
{
  "styles": {
    "util": false,
    "path": {
      "named": true
    }
  }
}
```

### `extendDefaultStyles`

Whether `styles` extends the defaults. Defaults to `true`. Set it to `false` to
replace the defaults completely.

### `checkImport`

Whether to check static imports. Defaults to `true`.

### `checkDynamicImport`

Whether to check dynamic `import()` expressions. Defaults to `true`.

### `checkExportFrom`

Whether to check `export … from` declarations. Defaults to `false`.

### `checkRequire`

Whether to check `require()` calls. Defaults to `true`.

## Differences from upstream

JavaScript preserves object-property insertion order, so upstream lists allowed
styles in the order written in the configuration. Rslint receives normalized
rule options as unordered maps and therefore uses a deterministic order:
`named`, `namespace`, `default`, `unassigned`, followed by custom style names in
alphabetical order. For example, `{default: true, named: true}` is reported as
`named or default` instead of upstream's `default or named`. This affects only
the order of style names in the diagnostic; the accepted imports are identical.

## Original Documentation

- [eslint-plugin-unicorn: import-style](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/import-style.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/import-style.js)
