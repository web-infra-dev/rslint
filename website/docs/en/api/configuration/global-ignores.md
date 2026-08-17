# globalIgnores

`globalIgnores` creates a flat-config entry whose patterns apply across the entire configuration.

## Example

```ts
import { defineConfig, globalIgnores, js } from '@rslint/core';

export default defineConfig([
  globalIgnores(['**/dist/**', '**/coverage/**']),
  js.configs.recommended,
]);
```

## Behavior

The returned entry contains only `ignores`, which makes it a global ignore rather than an entry-level ignore. The argument must be a non-empty array; passing another value or an empty array throws a `TypeError`.

See the [`ignores`](/config/ignoring-files) reference for pattern and traversal semantics.
