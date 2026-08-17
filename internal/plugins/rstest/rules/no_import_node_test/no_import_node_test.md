# no-import-node-test

## Rule Details

This rule reports static imports from `node:test` in Rstest test files. Such imports are commonly added accidentally by editor auto-import and select Node's test runner instead of Rstest.

Named imports containing only `test`, `it`, and `describe` can be safely fixed by changing the module to `@rstest/core`. Other imports are still reported, but are not fixed because the two modules do not have equivalent exports.

Examples of **incorrect** code for this rule:

```typescript
import { test } from 'node:test';
import * as nodeTest from 'node:test';
import { mock } from 'node:test';
```

Examples of **correct** code for this rule:

```typescript
import { test } from '@rstest/core';
import { test } from 'node:test/reporters';
const nodeTest = require('node:test');
```
