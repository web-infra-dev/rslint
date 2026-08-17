# no-import-node-test

## Rule Details

This rule reports imports of Node's test runner in Rstest test files, whether through an `import` declaration or a `require()` call, and whether the module is `node:test` itself or one of its sub-paths such as `node:test/reporters`. Such imports are commonly added accidentally by editor auto-import and select Node's test runner instead of Rstest.

Named imports of `node:test` containing only `test`, `it`, and `describe` can be safely fixed by changing the module to `@rstest/core`. Everything else is still reported, but is not fixed because the two modules do not have equivalent exports.

Examples of **incorrect** code for this rule:

```typescript
import { test } from 'node:test';
import * as nodeTest from 'node:test';
import { mock } from 'node:test';
import { tap } from 'node:test/reporters';
const nodeTest = require('node:test');
```

Examples of **correct** code for this rule:

```typescript
import { test } from '@rstest/core';
import { rs } from '@rstest/core';
const assert = require('node:assert');
```
