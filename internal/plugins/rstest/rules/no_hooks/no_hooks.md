# no-hooks

## Rule Details

Disallows lifecycle hooks when a project prefers each test to set up and clean up its own state explicitly.

The rule covers `beforeAll`, `beforeEach`, `afterEach`, and `afterAll` from globals, imports, aliases, and Playwright test objects.

## Incorrect

```ts
beforeEach(() => {
  resetDatabase();
});

test('creates a user', () => {});
```

## Correct

```ts
test('creates a user', () => {
  const database = createTestDatabase();

  expect(database.users.create({ name: 'Ada' }).name).toBe('Ada');
});
```

## Options

```json
{
  "rstest/no-hooks": [
    "error",
    {
      "allow": ["afterEach"]
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `allow` | `string[]` | `[]` | Lifecycle hooks to allow: `beforeAll`, `beforeEach`, `afterEach`, or `afterAll`. |
