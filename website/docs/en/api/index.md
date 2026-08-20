# API overview

`@rslint/core` provides configuration helpers and presets for authoring flat configs, plus the `Rslint` class for programmatic linting:

```ts
import { Rslint, defineConfig, globalIgnores, js, ts } from '@rslint/core';
```

## Configuration API

Use the configuration exports to compose flat configs without constructing internal loader objects.

| API                                                  | Description                                 |
| ---------------------------------------------------- | ------------------------------------------- |
| [`defineConfig`](/api/configuration/define-config)   | Type-safe identity helper for a flat config |
| [`globalIgnores`](/api/configuration/global-ignores) | Creates a global ignore entry               |
| [`globals`](/api/configuration/globals)              | Lazily loaded runtime-global catalog        |

## Plugin API

These exports provide the configuration API for Rslint's built-in rule groups and plugins. The rules are compiled into Rslint, so the upstream ESLint plugins do not need to be installed. Each page lists the upstream project its rules are based on and maps every available Rslint preset to its upstream configuration source.

An upstream mapping describes the compatibility source, not a byte-for-byte copy. Rslint presets include the rules Rslint currently supports and may contain adaptations for its native runtime.

| API                                                   | Description                                         |
| ----------------------------------------------------- | --------------------------------------------------- |
| [`js`](/api/presets/js)                               | JavaScript                                          |
| [`ts`](/api/presets/ts)                               | TypeScript baseline, strict, and stylistic variants |
| [`reactPlugin`](/api/presets/react-plugin)            | React                                               |
| [`reactHooksPlugin`](/api/presets/react-hooks-plugin) | React Hooks                                         |
| [`importPlugin`](/api/presets/import-plugin)          | Imports                                             |
| [`promisePlugin`](/api/presets/promise-plugin)        | Promises                                            |
| [`jestPlugin`](/api/presets/jest-plugin)              | Jest                                                |
| [`rstestPlugin`](/api/presets/rstest-plugin)          | Rstest                                              |
| [`unicornPlugin`](/api/presets/unicorn-plugin)        | Unicorn                                             |
| [`jsxA11yPlugin`](/api/presets/jsx-a11y-plugin)       | JSX accessibility                                   |

## JavaScript API

Use the JavaScript API when a Node.js-compatible host needs to invoke Rslint directly. The [JavaScript API guide](/guide/js-api) covers common workflows; the reference below documents the complete class surface.

| API                     | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| [`Rslint`](/api/rslint) | ESLint-style class for linting files or in-memory source |
