# no-restricted-jest-methods

## Rule Details

Disallow specific `jest` methods. Use this rule to ban particular `jest.*` calls that your team prefers to avoid, such as spies, mocks, or timer helpers, and optionally provide custom messages explaining the preferred alternative.

Restrictions are matched against the **first member** of a `jest` call chain. For example, banning `fn` reports `jest.fn()` and `jest["fn"]()`, but not bare `jest` or `jest()` without a method name.

By default, no `jest` methods are restricted.

Examples of **incorrect** code for this rule with the following configuration:

```json
{
  "jest/no-restricted-jest-methods": [
    "error",
    {
      "fn": null,
      "mock": "Do not use mocks",
      "advanceTimersByTime": null
    }
  ]
}
```

```javascript
jest.useFakeTimers();
it('calls the callback after 1 second via advanceTimersByTime', () => {
  // ...

  jest.advanceTimersByTime(1000);

  // ...
});

test('plays video', () => {
  const spy = jest.spyOn(video, 'play');

  // ...
});
```

## Options

- First argument (required to enable the rule): object whose keys are restricted `jest` method names and whose values are custom messages.
  - Keys are method names such as `fn`, `mock`, `spyOn`, or `advanceTimersByTime`.
  - Values are either a string (shown as the diagnostic message) or `null` (uses the default message: ``Use of `{method}` is disallowed``).

Examples of **incorrect** code with `{ "mock": "Do not use mocks" }`:

```javascript
jest.mock();
jest['mock']();
```

Examples of **incorrect** code with `{ "fn": null }`:

```javascript
jest.fn();
jest['fn']();
```

Examples of **incorrect** code with `{ "advanceTimersByTime": null }`:

```javascript
import { jest } from '@jest/globals';

jest.advanceTimersByTime(1000);
```

## Original Documentation

- [jest/no-restricted-jest-methods](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-restricted-jest-methods.md)
