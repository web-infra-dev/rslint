# no-commented-out-tests

## Rule Details

Disallow commented-out Rstest test cases and `describe` blocks. Commented-out tests are easy to overlook during review and can remain disabled indefinitely. Prefer removing dead tests or using an explicit Rstest API such as `.skip` or `.todo` when a disabled test should remain visible.

The rule reconstructs consecutive line comments and parses the text inside regular line and block comments as TypeScript. It reports a comment when a line, after optional whitespace, contains a call rooted at `test`, `it`, or `describe` that ultimately registers a test or suite.

The matcher understands:

- Direct registrations and static dot or bracket member access.
- The `only`, `skip`, `todo`, `fails`, `concurrent`, and `sequential` modifiers where supported by Rstest.
- The `runIf` and `skipIf` conditional factories.
- Array and tagged-template forms of `each` and `for`, including explicit type arguments.
- Fixture APIs created with `test.extend`, including repeated extension.
- Parenthesized and optional calls.

Examples of **incorrect** code for this rule:

```typescript
// test('adds two numbers', () => {});
// it.skip('is temporarily disabled', () => {});

// describe.only.concurrent('focused suite', () => {});
// test.for([{ value: 1 }])('handles $value', ({ value }) => {});
// test.extend(fixtures).only('uses fixtures', () => {});

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

Examples of **correct** code for this rule:

```typescript
// These calls only create another API or registrar. They do not define a test.
// test.runIf(condition)
// test.each(rows)
// test.for`value | expected`
// test.extend(fixtures)

// Dynamic, unknown, and non-Rstest APIs are outside this rule's scope.
// test[modifier]('name', () => {});
// test.unknown('name', () => {});
// fit('name', () => {});
// rstest.test('name', () => {});
```

## Limitations

The code inside a comment no longer participates in the source file's symbol table. The rule therefore recognizes only the canonical `test`, `it`, and `describe` names. It does not resolve imported aliases, locally renamed APIs, variables returned from `test.extend`, or namespace and in-source forms such as `import.meta.rstest`.

Documentation comments (`/** ... */`) and calls inside Markdown fenced code blocks are ignored to avoid reporting examples as disabled tests. Regular prose must still form a standalone, syntactically valid Rstest registration before it is reported.

## References

- [Rstest `test` API](https://rstest.rs/api/runtime-api/test-api/test)
- [Rstest `describe` API](https://rstest.rs/api/runtime-api/test-api/describe)
