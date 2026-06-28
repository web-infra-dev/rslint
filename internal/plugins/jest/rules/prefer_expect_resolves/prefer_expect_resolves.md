# prefer-expect-resolves

## Rule Details

Prefer `await expect(promise).resolves.<matcher>` over
`expect(await promise).<matcher>` when asserting on a resolved promise value.

If the promise rejects, `expect(await promise)` throws before Jest can run the
matcher, so the failure looks like an unhandled rejection rather than a test
assertion. The `.resolves` modifier keeps failures inside Jest's matcher
pipeline and mirrors `.rejects`, which has no equivalent `await`-inside-`expect`
form.

This rule reports `await` used as an argument to `expect()` (including renamed
`expect` bindings from `@jest/globals`). It is fixable.

Examples of **incorrect** code for this rule:

```js
it('passes', async () => {
  expect(await someValue()).toBe(true);
});

it('is true', async () => {
  const myPromise = Promise.resolve(true);

  expect(await myPromise).toBe(true);
});

import { expect as pleaseExpect } from '@jest/globals';

pleaseExpect(await myPromise).toBe(true);
```

Examples of **correct** code for this rule:

```js
it('passes', async () => {
  await expect(someValue()).resolves.toBe(true);
});

it('is true', async () => {
  const myPromise = Promise.resolve(true);

  await expect(myPromise).resolves.toBe(true);
});

// rejections use `.rejects` — not in scope for this rule
it('errors', async () => {
  await expect(Promise.reject(new Error('oh noes!'))).rejects.toThrowError(
    'oh noes!',
  );
});
```

## Original Documentation

- [jest/prefer-expect-resolves](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-expect-resolves.md)
