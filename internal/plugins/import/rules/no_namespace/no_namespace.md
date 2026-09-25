# no-namespace

## Rule Details

Disallow namespace (`*`) imports in favor of default or named imports. This also
applies to TypeScript's `import type * as` declarations. Namespace re-exports and
TypeScript `import name = require('module')` declarations are not reported.

Examples of **incorrect** code:

```javascript
import * as helpers from './helpers';
import defaultExport, * as utilities from './utilities';
```

Examples of **correct** code:

```javascript
import defaultExport from './helpers';
import { format, parse } from './utilities';
```

When every reference accesses a named member, the rule can replace the namespace
import and its references with named imports:

```javascript
// Before
import * as helpers from './helpers';
helpers.format(value);
helpers['parse'](text);

// After
import { format, parse } from './helpers';
format(value);
parse(text);
```

Conflicting names receive an alias such as `helpers_format` or
`helpers_format_1`. Unused imports and namespaces used as objects, JSX tag names,
or in type annotations such as `type T = helpers.Type` are reported without a
fix. The rule does not provide suggestions.

## Options

`ignore` is an array of unique module glob patterns, defaulting to `[]`. Patterns
without `/` match the module's basename, so `*.svg` also matches
`./assets/icon.svg`. Matching is case-sensitive and supports braces and extended
globs.

With `{ "ignore": ["*.svg"] }`, this is allowed:

```javascript
import * as icon from './assets/icon.svg';
```

## Differences from upstream

Diagnostics follow eslint-plugin-import 2.32.0. Some automatic fixes differ:

- **Dynamic keys are not autofixed.** Dynamic keys such as `helpers[key]`, and
  uses of the namespace as a key such as `object[helpers]`, are reported without
  a fix. Upstream can incorrectly replace them with an unrelated named import.
  Choose the intended export manually.
- **Fixes preserve comments and valid bindings.** A fix is withheld when it
  would remove comments or generate an invalid local binding, such as
  `helpers.default`, `helpers['x-y']`, or `helpers[0]`. Rewrite these imports
  manually, preserving comments and using an explicit alias where needed.
- **Writes and deletions are not autofixed.** For example, `helpers.format = value`
  and `delete helpers.format` are reported without a fix. Replacing these targets
  with named imports would assign to an imported binding or introduce invalid
  syntax. Update the operation manually.
- **Aliases avoid name conflicts.** This includes all enclosing scopes and other
  generated imports. For example, a nested use of `helpers.format` still gets an
  alias when `format` is declared at the module level; upstream can miss that
  conflict.
- **Fixes preserve required spaces.** Replacing `return(helpers).format` keeps
  the separating space required for `return format`. Upstream can join the two
  tokens.
- **Exports named `constructor` or `__proto__` can be fixed.** Upstream's fix can
  fail on these names.

## Original Documentation

- [eslint-plugin-import: no-namespace](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-namespace.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-namespace.js)
