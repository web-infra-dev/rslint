# prefer-to-have-been-called-times

## Rule Details

Use `toHaveBeenCalledTimes()` instead of measuring a mock's `mock.calls` array with `toHaveLength()`. The call-count matcher states the intent directly, and its failure message names the mock and the recorded calls instead of reporting an array length. See the [Rstest mock matchers](https://rstest.rs/api/runtime-api/test-api/expect#mock-matchers).

This rule checks any assertion whose subject is written as `<value>.mock.calls`, including bracket access with a string or template key, receivers such as `this.service.handler`, `mocks[index]` or a call result, and `expect.soft`. It recognizes global `expect`, named and renamed imports, namespace imports, CommonJS bindings from `@rstest/core`, `rstack/test` and `@rstest/playwright`, `import.meta.rstest`, and test-context `expect`. Modifiers such as `not`, `resolves` and `rejects` are left untouched. The rule does not require type information and does not check whether the subject is really an Rstest mock.

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

The fix removes the `.mock.calls` accessors from the assertion's subject and renames the matcher to `toHaveBeenCalledTimes`. It keeps the matcher's arguments, the second `expect` argument, modifiers, accessor quotes, optional chaining on the matcher, and any parentheses or comments written around the subject.

Only standalone assertion statements are fixed; parentheses and `await` around the assertion are allowed. An assertion used in an assignment, a return value, an argument or another expression is reported without a fix, because the assertion object it returns carries the rewritten subject into every later matcher. Chains with more than one matcher, or with a property access or call after the matcher, are reported without a fix for the same reason.

`expect.poll` assertions are reported without a fix, because the fix would hand the mock itself to `expect.poll`, which repeatedly calls its argument.
