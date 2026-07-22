# no-commented-out-tests

## Rule Details

Disallow commenting out Rstest tests. Commented-out tests are easy to overlook during review and can remain disabled indefinitely. Prefer removing dead tests or using `.skip` / `.todo` when a disabled test should remain visible.

This rule reports line and block comments containing calls that start with `test`, `it`, or `describe`, including member forms such as `.skip`, `.only`, and `.for`.

Examples of **incorrect** code for this rule:

```typescript
// test('adds two numbers', () => {});
// it.skip('is temporarily disabled', () => {});
// describe.only('focused suite', () => {});
// test.for([{ value: 1 }])('handles $value', ({ value }) => {});

/*
describe('math', () => {});
*/
```

Examples of **correct** code for this rule:

```typescript
test('adds two numbers', () => {});
test.skip('is temporarily disabled', () => {});
test.todo('will be implemented later');

// Explain why this behavior is intentionally unsupported.
```

Rstest does not expose Jest's `f` / `x` aliases, so comments containing `fit`, `xit`, `xtest`, `fdescribe`, or `xdescribe` are not reported.
