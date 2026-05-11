# ESLint Plugin Compatibility

Rslint can load and run third-party ESLint plugins alongside its native rules, so you can keep using a plugin that rslint does not provide natively — its rules run through rslint unchanged.

> Rslint already ships native implementations of many popular rules. When a native rule exists, prefer it — it is faster and needs no extra setup. Reach for `eslintPlugins` only for plugins rslint does **not** cover natively.

## Quick start

Install the plugin, then declare it in `rslint.config.{js,mjs,ts,mts}` under `eslintPlugins`, choosing a prefix for its rules:

```ts
import { defineConfig } from '@rslint/core';
import somePlugin from 'eslint-plugin-something';

export default defineConfig([
  {
    files: ['**/*.ts', '**/*.tsx'],
    eslintPlugins: {
      // Pick any prefix that is NOT a reserved native namespace
      // (see below). Reusing a reserved name is rejected at
      // config-load time with a clear error.
      custom: somePlugin,
    },
    rules: {
      'custom/some-rule': 'error',
    },
  },
]);
```

The prefix you choose (`custom` above) becomes the namespace for that plugin's rules in the `rules` field — rename it to anything you like.

### Reserved namespaces

Rslint ships ported rules under these native namespaces, so you **cannot** use them as a key in `eslintPlugins`:

```
@stylistic, @typescript-eslint, import, jest, jsx-a11y, promise, react, react-hooks, unicorn
```

Using one of these is rejected at config-load time with a message that names the offending prefix and suggests a rename.

For the configuration schema itself, see [Configuration › eslintPlugins](/config#eslintplugins).

## Supported APIs

Plugin rules see the ESLint v10 `RuleContext` and `SourceCode` interfaces. The commonly used surfaces are available:

- `context.report` — both the descriptor form (`{ node, loc, messageId, data, fix, suggest }`) and the legacy positional form.
- `context.options`, `settings`, `id`, `filename`, `physicalFilename`, `cwd`, `languageOptions`.
- The `sourceCode` family — `getText`, `getTokens*`, `getComments*`, `scopeManager`, `getScope`, `getDeclaredVariables`, `markVariableAsUsed`, disable-directive accessors, `getNodeByRangeIndex`, `getAncestors`, and more.
- The `fixer` family, `suggest` (lazy or eager), ESQuery selectors, and ESTree + TypeScript visitor keys.

Plugins written against `espree` or `@typescript-eslint/parser` generally work as-is, since rslint presents an ESTree-shaped AST.

## Limitations

A few ESLint v10 features are out of scope. If your plugin depends on one of these, run it through ESLint directly.

| Limitation                                                                                       | Behavior                                                                                  |
| ------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------- |
| **Type-aware rules** (anything that uses the TypeScript type checker / `parserServices.program`) | Not supported — type information is not exposed to plugin rules.                          |
| **Custom parsers** (`languageOptions.parser`)                                                    | Ignored — the AST is always ESTree-shaped.                                                |
| **Processors** (`.vue`, `.svelte`, `.md`, …)                                                     | Not supported — non-JS/TS files are not linted.                                           |
| **`meta.schema` option validation**                                                              | Options are not validated up front; validate `context.options` defensively if it matters. |
| **`reportUnusedDisableDirectives`**                                                              | Not applied to plugin rules.                                                              |

If a rule throws while running, rslint records the error for that file and keeps running the other rules instead of aborting the lint.

## Troubleshooting

**My plugin rule never fires.** Check that the plugin is declared in an `eslintPlugins` block whose `files` pattern matches the file you are linting, and that the rule is set to `'error'` or `'warn'`. If rslint ships a native rule with the same name, the native rule wins.

**`context.parserServices.program is undefined`.** This is a type-aware rule — it is not supported (see [Limitations](#limitations)).

**A single file hangs or times out.** Each plugin-rule run has a per-file watchdog; when it fires, that file is marked as timed-out and the rest of the run continues. If you see this consistently on one file, please report it.
