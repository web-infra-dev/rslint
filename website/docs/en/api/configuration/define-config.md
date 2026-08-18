# defineConfig

`defineConfig` is the type-safe identity helper for an Rslint flat config.

See the [Configuration overview](/config/) for the available configuration options and links to their detailed guides.

## Example

```ts
import { defineConfig, js, ts } from '@rslint/core';

export default defineConfig([
  js.configs.recommended,
  ts.configs.recommended,
  {
    files: ['src/**/*.ts'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
    },
  },
]);
```

## Behavior

The function returns the supplied config unchanged while providing TypeScript autocomplete and validation. A config is an array of entries; a preset that expands to multiple entries can be listed directly.

`defineConfig` does not load, normalize, flatten, or merge the config. Those operations happen when Rslint loads the exported array.
