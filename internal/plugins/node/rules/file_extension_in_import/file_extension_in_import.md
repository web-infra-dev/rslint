# file-extension-in-import

Enforce file extension styles in imports and re-exports.

The rule checks relative and absolute paths in `import`, `export ... from`, and
literal `import()` expressions. Package imports, Node builtins, URLs, `require()`
calls, and computed import expressions are ignored.

## Options

The first option is `"always"` (the default) or `"never"`. The second option
overrides the style for individual target file extensions.

```ts
import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    plugins: ['node'],
    rules: {
      'node/file-extension-in-import': ['error', 'always', { '.json': 'never' }],
    },
  },
]);
```

With `"always"`, an import resolving to `helper.js` needs its extension:

```js
// Incorrect
import helper from './helper';
export * from './helper';

// Correct
import helper from './helper.js';
export * from './helper.js';
```

With `"never"`, the same file is imported as `./helper`. Removing an extension
is only fixed automatically when there is one matching entry in its directory.
For example, if both `helper.js` and `helper.json` exist, the rule reports the
extension without removing it.

Directory imports with an index file are expanded by `"always"`: `./helpers`
becomes `./helpers/index.js`. If several extensions exist, the rule follows
`settings.node.tryExtensions` order, then falls back to the first directory entry.
The default search order is `.js`, `.json`, `.node`, `.mjs`, `.cjs`.

TypeScript imports use the emitted extension, such as `.js` for `.ts` and `.mjs`
for `.mts`. Overrides in the second option match the source file extension,
before this conversion. The rule reads the nearest TypeScript configuration,
including `jsx` and `allowImportingTsExtensions`.

Shared `settings.node` supports `tryExtensions`, `resolvePaths`, `tsconfigPath`,
and `typescriptExtensionMap`. The latter accepts extension pairs such as
`[['.ts', '.js'], ['.tsx', '.jsx']]`, or the presets `preserve`, `react`,
`react-jsx`, `react-jsxdev`, and `react-native`. An explicit extension map takes
precedence over TypeScript configuration. Relative lookup paths use the process
working directory.

## Differences from upstream

For `"never"`, paths containing escape sequences are reported without an
automatic fix. For example, `import './\u0061.js'` resolves like `import './a.js'`,
but upstream can remove the wrong characters from the escaped spelling. Use
the unescaped spelling or remove the extension manually.

Automatic fixes also leave non-string `import()` arguments unchanged. If a
custom extension contains a backslash, line break, or the string's quote
delimiter, the rule reports it without inserting it verbatim. For example,
mapping `.ts` to `.x'js` does not automatically modify `import './a.ts'`, which
would otherwise produce invalid JavaScript.

On Unix, directory imports whose names contain literal backslashes are not
expanded. For example, if a directory named `a\b` contains `index.js`, upstream
expands `import './a\\b'` to `import './a\\b/index.js'`; rslint leaves this
JavaScript import unchanged. Use directory names without literal backslashes.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/file-extension-in-import.md)
- [Upstream implementation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/file-extension-in-import.js)
