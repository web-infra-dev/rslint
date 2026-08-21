# prefer-each

## Rule Details

Prefer `.each` over wrapping Rstest `test`, `it`, `describe`, or hooks in
native `for`, `for...in`, and `for...of` loops. Parameterized registrations are
clearer in reporters and keep related cases grouped under one declaration.

Examples of **incorrect** code:

```ts
for (const row of rows) {
  test(row.name, () => {
    expect(runCase(row)).toBe(true);
  });
}

for (const row of rows) {
  beforeEach(() => setup(row));
  test(row.name, () => {
    expect(doSomething()).toBe(row.expected);
  });
}
```

Examples of **correct** code:

```ts
test.each(rows)('$name', (row) => {
  expect(runCase(row)).toBe(true);
});

describe.each(rows)('$name', (row) => {
  beforeEach(() => setup(row));
  test('works', () => {
    expect(doSomething()).toBe(row.expected);
  });
});

test('business loop', () => {
  for (const row of rows) {
    consume(row);
  }
});
```

Loops inside a real test callback are ignored. Those loops usually represent
business logic or setup work, not repeated registrations.

This rule only reports. It does not autofix, and it never recommends `.for`.
Rstest's `.for` callback receives `(row, context)`, which is not a mechanical
replacement for a hand-written loop.

## Rstest specifics

Rstest resolves same-file aliases and supports multiple API entry points, so
the rule follows final registrations from:

- globals
- `@rstest/core` named imports and aliases
- `require('@rstest/core')`
- namespace imports
- `import.meta.rstest`
- same-file `const` aliases
- `@rstest/playwright`

When exactly one test registration appears in the loop, the suggestion keeps
the original API family:

- `it(...)` suggests `it.each`
- `test(...)` suggests `test.each`
- Playwright test registrations also suggest `test.each`

Any loop containing `describe`, hooks, or multiple registrations suggests
`describe.each`.

## Differences from ESLint

Unlike `eslint-plugin-jest`, the Rstest port does not report a manual loop that
appears inside a real test callback, even if that callback earlier registered a
nested test.

This avoids a false positive on shapes like:

```ts
test('outer', () => {
  test('inner', () => {});

  for (const row of rows) {
    consume(row); // not reported
  }
});
```

## Original Documentation

- [eslint-plugin-jest: prefer-each](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-each.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-each.ts)
