# no-importing-rstest-globals

## Rule Details

This rule disallows importing Rstest test APIs when the project enables Rstest globals. Removing redundant imports keeps test files consistent with the configured runtime environment.

It checks named imports and destructured `require()` calls from `@rstest/core` and `rstack/test`, including aliased bindings. Type-only, default, and namespace imports are ignored. Do not enable this rule together with `rstest/prefer-importing-rstest-globals`.

In a destructured `require()`, the property key decides which export a binding pulls in, so a string-literal key is treated exactly like an identifier key: `const { 'expect': expect } = require('@rstest/core')` names the same export as `const { expect } = require('@rstest/core')` and is reported. A computed key counts only when it is a static string; `const { [expect]: local } = require('@rstest/core')` reads `expect` as a value rather than naming an export, so it is left alone.

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

A binding the file exports, as in `export const { expect } = require('@rstest/core')`, is reported without a fix as well: it belongs to the module's public surface, and removing it would delete an export that consumers rely on. So is a `require()` in a loop head, such as `for (const { expect } = require('@rstest/core'); ready;)`, which no variable statement of its own owns.
