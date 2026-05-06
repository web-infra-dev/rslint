# jest/no-deprecated-functions

## Rule Details

Jest sometimes deprecates globals and `jest` helpers in favor of newer APIs. This rule flags **calls** to those deprecated members and reports a replacement. A **fix** rewrites the callee to the suggested API; bracket-style access is preserved (for example `jest['genMockFromModule']` becomes `jest['createMockFromModule']`).

Which names are considered deprecated **depends on the Jest version** rslint uses for the file (from `settings.jest.version` or the resolved `package.json` dependency). A symbol is only reported once your configured major version is at least the version that deprecated it:

| Deprecated | Replacement | Starting at Jest major |
|------------|-------------|------------------------|
| `jest.resetModuleRegistry` | `jest.resetModules` | 15 |
| `jest.addMatchers` | `expect.extend` | 17 |
| `require.requireMock` | `jest.requireMock` | 21 |
| `require.requireActual` | `jest.requireActual` | 21 |
| `jest.runTimersToTime` | `jest.advanceTimersByTime` | 22 |
| `jest.genMockFromModule` | `jest.createMockFromModule` | 26 |

If the resolved Jest version is too old for any of these deprecations, the rule does not report them.

Examples of **incorrect** code for this rule (assuming a Jest version where the corresponding API is deprecated):

```js
jest.resetModuleRegistry();
jest.addMatchers({});
require.requireMock('a');
require.requireActual('a');
jest.runTimersToTime(1000);
jest.genMockFromModule('m');
```

Examples of **correct** code for this rule:

```js
jest.resetModules();
expect.extend({});
jest.requireMock('a');
jest.requireActual('a');
jest.advanceTimersByTime(1000);
jest.createMockFromModule('m');
```

## Original Documentation

- [jest/no-deprecated-functions](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-deprecated-functions.md)
