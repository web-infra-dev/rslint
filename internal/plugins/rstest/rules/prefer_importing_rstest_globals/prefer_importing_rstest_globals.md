# prefer-importing-rstest-globals

## Rule Details

This rule requires test files to import every Rstest API they use when Rstest globals are disabled. Explicit imports make each runtime dependency visible and prevent test APIs from being read from an unconfigured global environment.

It recognizes all APIs provided by Rstest's `globals` setting, including `expect`, `assert`, `rs`, `rstest`, and the lifecycle hooks. Imports and destructured `require()` calls from either `@rstest/core` or `rstack/test` satisfy the rule. Local value declarations with the same names are ignored. Do not enable this rule together with `rstest/no-importing-rstest-globals`.

## Incorrect

```javascript
describe('user service', () => {
  test('reads a user', () => {
    expect(getUser('1')).toEqual(user);
  });
});
```

## Correct

```javascript
import { describe, expect, test } from '@rstest/core';

describe('user service', () => {
  test('reads a user', () => {
    expect(getUser('1')).toEqual(user);
  });
});
```

## Autofix

The autofix adds the missing names to an existing named import or destructured `require()` from `@rstest/core` or `rstack/test`. Otherwise, it inserts a new `@rstest/core` import, or a `require()` declaration for CommonJS files, after an initial `"use strict"` directive or before the first statement. Names added to an existing named import are appended after its last specifier, in sorted order, so a default binding, existing aliases, and comments stay untouched.
