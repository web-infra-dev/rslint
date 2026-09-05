# prefer-importing-rstest-globals

## Rule Details

This rule requires test files to import every Rstest API they use when Rstest globals are disabled. Explicit imports make each runtime dependency visible and prevent test APIs from being read from an unconfigured global environment.

It recognizes all APIs provided by Rstest's `globals` setting, including `expect`, `assert`, `rs`, `rstest`, and the lifecycle hooks. Imports and destructured `require()` calls from either `@rstest/core` or `rstack/test` satisfy the rule. Local value declarations with the same names are ignored. Do not enable this rule together with `rstest/no-importing-rstest-globals`.

Only a runtime read of the API counts. A type position such as `const value: expect = input` never reaches the value, a write such as `expect = value` assigns to the global rather than reading it, and a label such as `test: for (...) {}` lives in its own namespace. None of them is reported.

Intrinsic JSX tags such as `<test />` do not read a variable and are ignored. Expressions inside JSX and the object of a member tag such as `<test.Component />` still count as runtime reads.

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

The autofix adds the missing names to an existing named import or destructured `require()` from `@rstest/core` or `rstack/test`. Otherwise, it inserts a new `@rstest/core` import, or a `require()` declaration for CommonJS files, after the directive prologue when that prologue makes the file strict, and before the first statement otherwise. Names added to an existing declaration are appended after its last specifier or binding, in sorted order, so a default binding, existing aliases, and comments stay untouched.

Three situations are reported without a fix:

- A name the file also assigns to, as in `expect = customExpect`. An import binding is read-only, so importing the name would break the assignment.
- A name the file already binds with a type-only import, whatever the module and whatever the import form. Adding a value binding of the same name would declare it twice; turning the existing binding into a value import is the edit to make instead.
- A destructured `require()` that ends in a rest element, as in `const { ...core } = require('@rstest/core')`. Naming one more property there would take it out of the rest object and break the code reading it from there.
