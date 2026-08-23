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

Hooks written inside a hook callback form their own run at that nesting level.
They are not compared against the hook that encloses them, and they do not end
the run that surrounds it:

```ts
// reported: `beforeEach` comes after `afterAll` in the same run
afterAll(() => {});
beforeEach(() => {});

// not reported: the inner run is ordered correctly on its own terms
afterAll(() => {
  beforeEach(() => {});
  afterEach(() => {});
});
```

## Differences from ESLint

The upstream rule tracks "inside a hook" as a single boolean and ignores every
call made while it is set. That conflates two separate scopes, and this rule
scopes each run to its own nesting level instead. Two consequences, both of
which this rule reports differently from upstream on hook-inside-hook shapes:

- A nested hook callback no longer ends the surrounding run. After a nested hook
  exits, an ordinary call resets upstream's run, so a later outer hook starts
  fresh and an inversion can go unreported. Here the surrounding run survives
  the callback, so
  `afterAll(() => { beforeEach(() => {}); doSomething(); }); beforeAll(() => {});`
  is reported.

- A nested hook is never compared against a hook from the enclosing run.
  Upstream lets the enclosing run's index leak in, which both reports correctly
  ordered nested runs and attributes reports to a hook from another nesting
  level. Here `afterAll(() => { beforeEach(() => {}); afterEach(() => {}); })`
  is accepted, and a genuinely inverted nested run is reported against the hook
  that precedes it at its own level.

## Original Documentation

- [eslint-plugin-jest: prefer-hooks-in-order](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-hooks-in-order.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-hooks-in-order.ts)
