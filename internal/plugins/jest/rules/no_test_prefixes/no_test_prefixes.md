# jest/no-test-prefixes

## Rule Details

Jest lets you focus or skip tests in more than one way. You can use **`.only` and `.skip`** on the normal APIs—for example `it.only`, `test.only`, `describe.only`, `it.skip`, `test.skip`, and `describe.skip`. Alternatively, Jest supports **short `f` and `x` prefixes**: `fit`, `fdescribe`, `xit`, `xtest`, and `xdescribe`.

This rule requires the **`.only` / `.skip`** style and reports calls that use the **`f` / `x`** spellings. Replacements are suggested automatically (for example `fit` → `it.only`, `xit` → `it.skip`).

Examples of **incorrect** code for this rule:

```typescript
fit('foo');
fdescribe('foo');
xit('foo');
xtest('foo');
xdescribe('foo');
```

Examples of **correct** code for this rule:

```typescript
it.only('foo');
test.only('foo');
describe.only('foo');
it.skip('foo');
test.skip('foo');
describe.skip('foo');
```

## Original Documentation

- [jest/no-test-prefixes](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-test-prefixes.md)
