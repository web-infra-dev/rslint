# no-hooks

## Rule Details

Disallow Rstest setup and teardown hooks. This rule reports the four lifecycle
hooks Rstest exposes: `beforeAll`, `beforeEach`, `afterEach`, and `afterAll`.

Examples of **incorrect** code for this rule:

```typescript
import { beforeEach, test } from '@rstest/core';
import { test as playwrightTest } from '@rstest/playwright';

beforeEach(() => {
  resetDatabase();
});

playwrightTest.beforeAll(async () => {
  await seedFixtures();
});

test('reads state', () => {});
```

Examples of **correct** code for this rule:

```typescript
import { test } from '@rstest/core';
import { beforeEach } from 'vitest';

test('reads state', () => {
  const database = createDatabase();
  expect(database.ready()).toBe(true);
});

beforeEach(() => {});
```

## Options

```json
{
  "rstest/no-hooks": ["error", { "allow": ["afterEach", "afterAll"] }]
}
```

### `allow`

This array option controls which Rstest hooks are allowed. Supported values
are:

- `"beforeAll"`
- `"beforeEach"`
- `"afterAll"`
- `"afterEach"`

By default, no hook is allowed.

Examples of **incorrect** code for the `{ "allow": ["afterEach"] }` option:

```json
{ "rstest/no-hooks": ["error", { "allow": ["afterEach"] }] }
```

```typescript
import { afterEach, beforeEach } from '@rstest/core';

beforeEach(() => {
  setupDatabase();
});

afterEach(() => {
  resetModules();
});
```

Examples of **correct** code for the `{ "allow": ["afterEach"] }` option:

```json
{ "rstest/no-hooks": ["error", { "allow": ["afterEach"] }] }
```

```typescript
import { afterEach, test } from '@rstest/core';

afterEach(() => {
  resetModules();
});

test('reads state', () => {
  const database = setupDatabase();
  expect(database.ready()).toBe(true);
});
```

## Differences from ESLint

- `@rstest/playwright` member hooks such as `test.beforeEach()` and
  `test.extend({}).afterAll()` are reported because they are real Rstest hook
  registrations.
- `@rstest/core` lookalikes such as `test.beforeEach()`, invalid chains such as
  `test.skip.beforeEach()`, and execution-time APIs like `onTestFinished()` are
  not reported.

## Original Documentation

- [eslint-plugin-jest: no-hooks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/no-hooks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/no-hooks.ts)
