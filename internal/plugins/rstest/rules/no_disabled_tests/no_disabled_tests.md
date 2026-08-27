# no-disabled-tests

## Rule Details

Disallows skipped tests and registrations without a callback. Disabled tests hide coverage gaps and make the suite's real status less clear.

`test.todo` and runtime control flow such as `context.skip()` are allowed. The rule recognizes Rstest globals, imports, aliases, and parameterized or fixture-based registrations.

## Incorrect

```ts
test.skip('imports a CSV file', () => {});
test('exports a CSV file');
```

## Correct

```ts
test('imports a CSV file', () => {
  expect(importCsv('name\nAda')).toEqual([{ name: 'Ada' }]);
});

test.todo('exports a CSV file');
```
