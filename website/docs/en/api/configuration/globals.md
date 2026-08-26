# globals

`globals` exposes the complete environment catalog from the pinned [`globals`](https://www.npmjs.com/package/globals) package. Each environment map is loaded and cached on first property access, so importing `@rslint/core` does not parse the entire catalog.

## Example

```ts
import { defineConfig, globals } from '@rslint/core';

export default defineConfig([
  {
    files: ['src/client/**'],
    languageOptions: {
      globals: globals.browser,
    },
  },
  {
    files: ['scripts/**'],
    languageOptions: {
      globals: globals.node,
    },
  },
]);
```

## Composing environments

Compose environments and project-specific names with object spreads:

```ts
languageOptions: {
  globals: {
    ...globals.browser,
    ...globals.worker,
    BUILD_ID: 'readonly',
  },
}
```

Environment values follow the upstream catalog: `false` is read-only and `true` is writable. See [`languageOptions.globals`](/config/language-options#languageoptionsglobals) for access values, overrides, and TypeScript interactions.
