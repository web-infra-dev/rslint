# no-import-node-test

## Rule Details

Disallows importing Node's test runner in an Rstest test file. Editor auto-import can otherwise select `node:test` instead of Rstest.

The rule reports static imports, `require()`, and TypeScript `import = require` forms for `node:test` and paths beneath it.

## Incorrect

```ts
import { test } from 'node:test';

test('creates a user', () => {});
```

## Correct

```ts
import { test } from '@rstest/core';

test('creates a user', () => {});
```

## Autofix

Changes an exact named import from `node:test` to `@rstest/core` when it imports only `test`, `it`, or `describe`.
