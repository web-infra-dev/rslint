# prefer-importing-jest-globals

## Rule Details

Prefer explicit imports from `@jest/globals` instead of relying on Jest's
injected globals. That keeps Jest APIs imported consistently across the codebase
and helps migrations when
[`injectGlobals`](https://jestjs.io/docs/configuration#injectglobals-boolean)
cannot be enabled (for example some ESM setups).

Examples of **incorrect** code for this rule:

```javascript
describe('foo', () => {
  it('accepts this input', () => {
    expect(true).toBeDefined();
  });
});
```

Examples of **correct** code for this rule:

```javascript
import { describe, expect, it } from '@jest/globals';

describe('foo', () => {
  it('accepts this input', () => {
    expect(true).toBeDefined();
  });
});
```

## Options

```ts
interface Options {
  types?: Array<'hook' | 'describe' | 'test' | 'expect' | 'jest' | 'unknown'>;
}
```

### `types`

A list of Jest global kinds to enforce explicit imports for. By default all Jest
globals are enforced. Restricting `types` is useful during incremental
migrations — for example enforcing only the `jest` helper when adopting ESM.

Examples of **incorrect** code with `{ "types": ["jest"] }`:

```json
{ "jest/prefer-importing-jest-globals": ["error", { "types": ["jest"] }] }
```

```javascript
jest.useFakeTimers();
```

Examples of **correct** code with `{ "types": ["jest"] }`:

```json
{ "jest/prefer-importing-jest-globals": ["error", { "types": ["jest"] }] }
```

```javascript
import { jest } from '@jest/globals';

jest.useFakeTimers();

// other globals may still be injected
describe('suite', () => {
  test('foo', () => {
    expect(true).toBeDefined();
  });
});
```

## Original Documentation

- [jest/prefer-importing-jest-globals](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-importing-jest-globals.md)
