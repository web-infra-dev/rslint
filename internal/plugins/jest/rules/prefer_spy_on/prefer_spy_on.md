# prefer-spy-on

## Rule Details

When a function is mocked by assigning `jest.fn()` to an object property, the
original implementation is lost and must be restored manually during cleanup.
`jest.spyOn()` wraps the existing method as a mock while preserving its behavior,
and Jest can restore it with `jest.restoreAllMocks()`, `mockFn.mockRestore()`, or
the `restoreMocks` config option.

The mock returned by `jest.spyOn()` behaves like the original function until you
change it with `mockImplementation()`, `mockReturnValue()`, or another mock API.

Examples of **incorrect** code for this rule:

```js
Date.now = jest.fn();
Date.now = jest.fn(() => 10);
obj.a = jest.fn();
window.fetch = jest.fn(() => ({})).mockReturnValue('ok');
```

Examples of **correct** code for this rule:

```js
jest.spyOn(Date, 'now');
jest.spyOn(Date, 'now').mockImplementation(() => 10);
jest.spyOn(obj, 'a').mockImplementation();
jest.spyOn(window, 'fetch').mockReturnValue('ok');
```

## Original Documentation

- [jest/prefer-spy-on](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-spy-on.md)
