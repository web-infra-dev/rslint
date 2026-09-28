# no-test-return-statement

## Rule Details

Disallow `return` statements in test callbacks. Rstest awaits a promise returned from a test and ignores any other returned value, so returning is never needed to make a test wait: an `async` callback with `await` expresses the same thing and keeps every test written the same way. This rule enforces that convention; it does not report code that fails at runtime.

Only a `return` written directly in the callback's block body is reported, and only the first one. A `return` nested in another statement, such as an early `if (!supported) return;` guard, is allowed, and so is a `return` inside a nested function. An arrow function written without braces has no return statement and is not reported. Hooks and `describe` callbacks are not checked; a hook such as `beforeEach` may return a cleanup function.

Both Rstest call shapes are recognized — `test(name, fn, timeout)` and `test(name, options, fn)`. A callback passed by name is checked when it refers to a function declared in the same file and never reassigned, and only while that name is used for nothing but test callbacks: a function that is also called directly, exported, or referenced any other way is skipped, because its return value may be needed there. The rule recognizes Rstest globals and APIs imported from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, including aliases, namespace and CommonJS imports, `import.meta.rstest`, parameterized cases such as `test.each` and `test.for`, chained modifiers such as `test.concurrent`, `test.todo`, and fixture-extended test APIs built with `test.extend`. Foreign test frameworks, locally shadowed names, and type-only bindings are ignored. Type information is not required.

## Incorrect

```ts
import { expect, test } from '@rstest/core';

test('loads the profile', () => {
  return fetchProfile('ann').then((profile) => {
    expect(profile.name).toBe('Ann');
  });
});
```

## Correct

```ts
import { expect, test } from '@rstest/core';

test('loads the profile', async () => {
  const profile = await fetchProfile('ann');

  expect(profile.name).toBe('Ann');
});
```
