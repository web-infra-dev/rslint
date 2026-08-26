# warn-todo

## Rule Details

Disallows `.todo` on Rstest `test`, `it`, and `describe` registrations. A todo registration names work that has not been written yet and never runs, so one left in a committed suite is a test that can never fail.

Every `.todo` registration is reported, including one combined with other modifiers such as `test.only.todo(...)` or `test.todo.each(...)`. The rule recognizes Rstest globals, `@rstest/core` imports, namespace and `import.meta.rstest` members, same-file aliases, and `@rstest/playwright` registrations; a `test` or `describe` from any other source is left alone. Only the `.todo` accessor registers a todo, so a `todo` property in a test options object is not reported.

This rule is the counterpart of `prefer-todo`, which steers empty tests toward `.todo`. Enable at most one of the two.

## Incorrect

```ts
test.todo('exports a CSV file');

describe.todo('CSV import');
```

## Correct

```ts
test('exports a CSV file', () => {
  expect(exportCsv([{ name: 'Ada' }])).toBe('name\nAda');
});

describe('CSV import', () => {
  test('parses a header row', () => {
    expect(importCsv('name\nAda')).toEqual([{ name: 'Ada' }]);
  });
});
```
