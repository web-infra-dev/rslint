# prefer-expect-resolves

## Rule Details

Prefer `await expect(promise).resolves.toBe(value)` over placing `await` inside `expect()`. The [resolves modifier](https://rstest.rs/api/runtime-api/test-api/expect#promise-matchers) describes an unexpected rejection as an assertion failure.

The rule recognizes globals, named and renamed imports, namespace imports, and CommonJS bindings from `@rstest/core`, `rstack/test`, and `@rstest/playwright`. It also recognizes direct `import.meta.rstest.expect` calls, destructured imports from `import.meta.rstest`, and test-context `expect`, including renamed destructuring and the second callback parameter of `test.for`. Reassigned or locally shadowed bindings are excluded. Copying `import.meta.rstest.expect` to a variable is not currently recognized.

Single method assertions, including Chai method assertions, are checked. Chai property assertions, chains containing multiple assertions, `expect.soft`, `expect.poll`, `expect.element`, browser locator/page matchers, and chains already using `resolves` or `rejects` are excluded. Without type information, the rule checks syntax. With type information, known non-promise values and unions with non-promise alternatives are excluded.

## Incorrect

```ts
import { expect, test } from "@rstest/core";

test("loads the account status", async () => {
  expect(await loadAccountStatus()).toBe("active");
});
```

## Correct

```ts
import { expect, test } from "@rstest/core";

test("loads the account status", async () => {
  await expect(loadAccountStatus()).resolves.toBe("active");
});
```

## Autofix

The fix removes the inner `await`, adds `.resolves` after `expect(...)`, and awaits the complete assertion. An existing outer `await` is retained. Comments and parentheses are preserved.

Automatic fixes require a known thenable type and a built-in value matcher: `toBe`, `toEqual`, `toStrictEqual`, `toBeDefined`, `toBeUndefined`, `toBeNull`, `toBeTruthy`, `toBeFalsy`, `toBeNaN`, the four numeric comparison matchers, `toBeCloseTo`, `toContain`, `toContainEqual`, `toHaveLength`, `toMatch`, `toMatchObject`, `toHaveProperty`, or `toBeInstanceOf`. The optional expect message and every matcher argument must be string, number, bigint, boolean, null, or non-interpolated template literals. Expressions such as variables and calls may observe different values when evaluated before the promise settles.

Optional chains, explicit type arguments on `expect`, custom or locally overridden matchers, and assertions whose result is assigned or passed to another expression are reported without edits. The assertion must be an expression statement, a return value, or directly awaited; parentheses are allowed. Throw, snapshot, and Chai method assertions are reported without edits.

When the promise type is unknown, the rule offers the same rewrite as a suggestion under the edit conditions above. Check that the awaited value is always a promise before applying it: JavaScript `await` also accepts plain values, while `.resolves` requires a thenable.
