# prefer-todo

## Rule Details

Prefers `test.todo(...)` for a test that has no implementation. A todo test states the work is intentional instead of looking like a passing test.

The rule reports missing callbacks and inline empty callbacks. Existing `test.todo` registrations are allowed.

## Incorrect

```ts
test('exports a CSV file');

test('imports a CSV file', () => {});
```

## Correct

```ts
test.todo('exports a CSV file');

test('imports a CSV file', () => {
  expect(importCsv('name\nAda')).toEqual([{ name: 'Ada' }]);
});
```

## Autofix

Rewrites the registration to `test.todo`, replacing a `.skip` accessor at the call site and removing an empty callback argument. A registration whose skip comes from an alias, or that chains `.skip` twice, is reported without a fix, because rewriting one accessor would leave the other skip active.
