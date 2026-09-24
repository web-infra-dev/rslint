# max-dependencies

## Rule Details

Limit the number of distinct dependencies in a file. Static imports, re-exports,
dynamic `import()` calls, and direct `require()` calls count when their module
path is a string literal. Repeated paths count once, even across these forms.
Paths are compared by their string values without resolving them to files.

The rule reports once on the last module path when the limit is exceeded. That
path can be a duplicate or an ignored type import. The rule does not provide
automatic fixes or suggestions.

With `{ "max": 2 }`, this code is **incorrect**:

```javascript
import a from './a';
const b = require('./b');
import c from './c';
```

This code is **correct**:

```javascript
import a from './a';
const anotherA = require('./a');
import { x, y, z } from './foo';
```

Template literals and computed paths such as `require(name)` do not count.
TypeScript import assignments (`import name = require('module')`) and import
types (`type T = import('module').T`) do not count either.

## Options

### `max`

The maximum number of distinct dependencies. With no options, the limit is `10`.
Like eslint-plugin-import 2.32.0, passing an options object without `max` disables
the check. Specify `max` when configuring `ignoreTypeImports`.

### `ignoreTypeImports`

Set this to `true` to exclude declarations using `import type`. The default is
`false`. Inline type specifiers (`import { type T } from 'module'`) and type
re-exports still count.

With `{ "max": 2, "ignoreTypeImports": true }`, this code is **correct**:

```typescript
import a from './a';
import b from './b';
import type c from './c';
```

## Differences from upstream

Setting `{ "max": -1 }` for a file containing only `export const value = 1;`
makes eslint-plugin-import stop with an error. rslint reports no problem for
that file. Use `{ "max": 0 }` to forbid dependencies.

## Original Documentation

- [eslint-plugin-import: max-dependencies](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/max-dependencies.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/max-dependencies.js)
