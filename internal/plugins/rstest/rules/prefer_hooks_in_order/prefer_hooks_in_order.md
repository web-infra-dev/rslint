# prefer-hooks-in-order

## Rule Details

Requires lifecycle hooks to appear in the order Rstest runs them: `beforeAll`, `beforeEach`, `afterEach`, then `afterAll`. Keeping declarations in runtime order makes test setup easier to read.

The rule checks consecutive hook declarations. A non-hook call begins a new sequence, allowing separate setup groups when needed.

## Incorrect

```ts
afterAll(() => closeDatabase());
beforeAll(() => openDatabase());
```

## Correct

```ts
beforeAll(() => openDatabase());
beforeEach(() => resetDatabase());
afterEach(() => clearDatabase());
afterAll(() => closeDatabase());
```
