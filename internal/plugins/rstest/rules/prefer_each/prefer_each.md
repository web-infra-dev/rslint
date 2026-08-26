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

A loop is judged by what it registers, not by where it sits. A loop whose body
holds no registration is business logic or setup work and is never reported,
including inside a test callback. A loop that does register is reported even
inside a test callback, because it still registers once per iteration.

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

`eslint-plugin-jest` keeps a single flat list of registrations for the whole
file and gates it on an `inTestCaseCall` boolean that every `test(...)` sets and
every `test(...)` exit clears. Both halves leak across scopes. This port gives
each loop its own frame instead: a registration belongs to the innermost loop
whose body contains it, and a loop is reported from that frame alone. Four
observable differences follow.

The message names what the loop itself registers. Upstream's flat list still
holds the enclosing `test(...)` when a loop inside its callback is reported, so
it recommends `describe.each` for a loop that registers one test:

```ts
test('outer', () => {
  for (const row of rows) {
    test(row.name, () => {}); // `test.each` here, `describe.each` upstream
  }
});
```

A loop in a non-callback argument position is reported. Upstream is still inside
the open `test(...)` when the loop exits, so it stays silent:

```ts
test('a', () => {}, (function () {
  for (const row of rows) {
    beforeEach(() => {}); // `describe.each` here, unreported upstream
  }
  return 5;
})());
```

An outer loop keeps its own registrations when it contains another loop.
Upstream clears the shared list when the inner loop is entered, which discards
the outer loop's registration:

```ts
for (const suite of suites) {
  test(suite.name, () => {}); // outer loop is `test.each` here, unreported upstream
  for (const item of suite.items) {
    consume(item);
  }
}
```

A registration in a nested loop's header belongs to the enclosing loop, because
the header is evaluated once per iteration of that enclosing loop. Upstream
clears its shared list when the inner loop is entered, so the registration is
discarded and nothing is reported:

```ts
for (const suite of suites) {
  // outer loop is `test.each` here, unreported upstream
  for (const row of getRows(test(suite.name, () => {}))) {
    consume(row);
  }
}
```

The same applies to a nested classic `for` initializer, condition, or update
expression.

## Original Documentation

- [eslint-plugin-jest: prefer-each](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-each.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-each.ts)
