# no-commented-out-tests

## Rule Details

Disallow commented-out Rstest test cases and `describe` blocks. Commented-out tests are easy to overlook during review and can remain disabled indefinitely. Prefer removing dead tests or using an explicit Rstest API such as `.skip` or `.todo` when a disabled test should remain visible.

The rule examines the text inside line and block comments. It reports the entire comment when a line, after optional whitespace, contains either of these shapes:

- A call rooted at `test`, `it`, or `describe`, optionally followed by dot or string-bracket member chains and type arguments.
- A tagged-template parameterized test or `describe` block ending in `.each` or `.for`.

This covers direct calls, modifiers, chained modifiers, array-based parameterization, and tagged-template parameterization. Member chains may span multiple lines inside a block comment.

Examples of **incorrect** code for this rule:

```typescript
// test('adds two numbers', () => {});
// it.skip('is temporarily disabled', () => {});

// describe.only.concurrent('focused suite', () => {});
// test.for([{ value: 1 }])('handles $value', ({ value }) => {});

// test.each`
//   value | expected
//   ${1}  | ${2}
// `('returns $expected', ({ value, expected }) => {});

/*
describe
  .only
  .concurrent('math', () => {});
*/
```

## References

- [Rstest `test` API](https://rstest.rs/api/runtime-api/test-api/test)
- [Rstest `describe` API](https://rstest.rs/api/runtime-api/test-api/describe)
