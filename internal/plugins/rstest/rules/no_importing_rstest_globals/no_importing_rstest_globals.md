# no-importing-rstest-globals

## Rule Details

This rule disallows importing Rstest test APIs when the project enables Rstest globals. Removing redundant imports keeps test files consistent with the configured runtime environment.

It checks named imports and destructured `require()` calls from `@rstest/core` and `rstack/test`, including aliased bindings. Type-only, default, and namespace imports are ignored. Do not enable this rule together with `rstest/prefer-importing-rstest-globals`.

## Incorrect

```javascript
import { expect, test } from '@rstest/core';

test('formats a user name', () => {
  expect(formatUserName(user)).toBe('Ada Lovelace');
});
```

## Correct

```javascript
test('formats a user name', () => {
  expect(formatUserName(user)).toBe('Ada Lovelace');
});
```

## Autofix

The autofix removes each redundant import specifier or destructured property. It removes the whole import or variable statement only when every binding it declares is itself safe to remove; otherwise the surviving bindings keep the declaration.

An aliased binding, a binding used outside an invocation, or a destructuring pattern containing a default value or rest element is reported without a fix because removing it could leave a reference undefined or change destructuring behavior.
