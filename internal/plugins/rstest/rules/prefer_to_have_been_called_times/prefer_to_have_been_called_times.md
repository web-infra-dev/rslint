# prefer-to-have-been-called-times

## Rule Details

Use `toHaveBeenCalledTimes()` instead of measuring a mock's `mock.calls` array with `toHaveLength()`. The call-count matcher states the intent directly, and its failure message names the mock and the recorded calls instead of reporting an array length. See the [Rstest mock matchers](https://rstest.rs/api/runtime-api/test-api/expect#mock-matchers).

This rule checks any assertion whose subject is written as `<value>.mock.calls`, including bracket access with a string or template key, receivers such as `this.service.handler`, `mocks[index]` or a call result, and `expect.soft`. It recognizes global `expect`, named and renamed imports, namespace imports, CommonJS bindings from `@rstest/core`, `rstack/test` and `@rstest/playwright`, `import.meta.rstest`, and test-context `expect`. Modifiers such as `not`, `resolves` and `rejects` are left untouched. The rule does not require type information and does not check whether the subject is really an Rstest mock.

Only `expect()` and `expect.soft()` assert on the value passed to them, so those are the assertions the rule checks. `expect.poll()` takes a callback instead of a value and `expect.element()` asserts on a browser locator, so neither is checked. In a Chai chain, a matcher that replaces the assertion subject — `property`, `ownProperty`, `toContain`, `toThrow` and the like — ends the chain for this rule: a `toHaveLength` after `expect(handler.mock.calls).property(0)` measures the arguments of one recorded call, not the call count.

Optional links inside the subject, such as `expect(handler?.mock.calls)`, are not checked, because the accessors cannot be removed without changing the guard. A bracket key written as an identifier, such as `expect(handler[mock].calls)`, reads a variable rather than naming the `mock` property, so it is not checked either. Assertions on other mock state, on `mock.calls.length`, or on a single recorded call such as `mock.calls[0]` are outside the rule, and a matcher named by a variable, as in `expect(handler.mock.calls)[matcher](2)`, is not checked.

## Incorrect

```ts
import { expect, rs, test } from '@rstest/core';

test('retries the request twice', async () => {
  const request = rs.fn();

  await retry(request, { attempts: 2 });

  expect(request.mock.calls).toHaveLength(2);
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';

test('retries the request twice', async () => {
  const request = rs.fn();

  await retry(request, { attempts: 2 });

  expect(request).toHaveBeenCalledTimes(2);
});
```

## Autofix

The fix removes the `.mock.calls` accessors from the assertion's subject and renames the matcher to `toHaveBeenCalledTimes`. It keeps the matcher's arguments, the second `expect` argument, modifiers, accessor quotes, optional chaining on the matcher, and any parentheses or comments written around the subject. Explicit type arguments on `expect()` and on the matcher are removed, because they describe the subject and the matcher that the fix replaces; when a comment sits inside either list, the assertion is reported without a fix.

The fix is offered only when the expected count is a number written in source, such as `2`, `0x2` or `2 as const`. `toHaveLength` compares its argument loosely and `toHaveBeenCalledTimes` compares strictly, so rewriting `toHaveLength('1')` would turn a passing assertion into a failing one, and rewriting `not.toHaveLength('1')` would turn a failing one into a passing one. A count that is a variable, a string, a boolean or an expression is reported without a fix, as are a missing count and extra matcher arguments.

`expect()` captures the `mock.calls` array as it runs, while `toHaveBeenCalledTimes` reads `mock.calls` when the matcher runs. Anything evaluated in between can reset the mock and change the count the rewritten assertion sees, so the remaining `expect()` arguments must be literals too: `expect(handler.mock.calls, 'called once')` is fixed, `expect(handler.mock.calls, (handler.mockClear(), 'called once'))` is not.

An assertion whose subject is reached through `super`, as in `expect(super.mock.calls)`, is reported without a fix, because a bare `super` is not a value and `expect(super)` would not parse.

Only standalone assertion statements are fixed; parentheses and `await` around the assertion are allowed. An assertion used in an assignment, a return value, an argument or another expression is reported without a fix, because the assertion object it returns carries the rewritten subject into every later matcher. Chains with more than one matcher, or with a property access or call after the matcher, are reported without a fix for the same reason.
