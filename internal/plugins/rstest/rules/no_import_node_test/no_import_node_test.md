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

Changes an exact named import from `node:test` to an Rstest module when it imports only `test`, `it`, or `describe`.

The test API is reachable as both `@rstest/core` and `rstack/test`, and the fix writes whichever of the two the file already uses — in another import, a `require`, or a `/// <reference types="..." />` directive — so that the rewritten import resolves in the project it was written for. A file that names both, or names neither, gets `@rstest/core`.
