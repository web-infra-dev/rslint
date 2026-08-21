# prefer-hooks-in-order

## Rule Details

Prefer Rstest lifecycle hooks in the same order Rstest runs them:

1. `beforeAll`
2. `beforeEach`
3. `afterEach`
4. `afterAll`

This rule checks each consecutive run of hook calls and reports a hook that
appears after another hook from a later runtime stage. A normal non-hook call
starts a new run. Comments, blank lines, and statements that do not contain a
call do not.

Examples of **incorrect** code for this rule:

```ts
afterAll(() => {});
beforeAll(() => {});

import { test } from "@rstest/playwright";

test.afterEach(() => {});
test.beforeEach(() => {});
```

Examples of **correct** code for this rule:

```ts
beforeAll(() => {});
beforeEach(() => {});
afterEach(() => {});
afterAll(() => {});

afterAll(() => {});
doSomething();
beforeAll(() => {});
```

The rule follows Rstest hooks imported from `@rstest/core`, globals,
CommonJS/namespace forms, `import.meta.rstest`, same-file aliases, and
`@rstest/playwright` member hooks such as `test.beforeEach()` and
`test.extend({}).afterAll()`.

## Differences from ESLint

- Nested hook callbacks do not accidentally end the surrounding hook run. A
  later outer hook is still compared against the outer hook sequence after an
  inner hook callback exits.

## Original Documentation

- [eslint-plugin-jest: prefer-hooks-in-order](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-hooks-in-order.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-hooks-in-order.ts)
