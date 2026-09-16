# prefer-to-have-been-called

## Rule Details

Use `toHaveBeenCalled()` instead of comparing a mock's call count with zero. Write `.not.toHaveBeenCalled()` to assert that a mock was never called, or `.toHaveBeenCalled()` to assert that it was called at least once. These matchers express the intent directly and provide clearer failure messages. See the [Rstest mock assertions](https://rstest.rs/api/runtime-api/test-api/expect#mock-matchers).

This rule checks `toHaveBeenCalledTimes(0)` and its `toBeCalledTimes(0)` alias, including negated assertions, dot and literal bracket access, `soft`, `poll`, and `resolves` or `rejects` chains. It recognizes global `expect`, named and renamed imports, namespace imports, CommonJS bindings from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, `import.meta.rstest`, and test-context `expect`. Bindings that are reassigned anywhere in the file are not checked. It does not require type information.

Parentheses and `as` or angle-bracket type assertions around the zero are accepted. Variables, computed expressions, unary `+0` or `-0`, and `satisfies` or non-null wrappers are not checked. Type assertions and non-null wrappers around the assertion receiver also stop recognition. Chai language chains and individual matching calls within multi-assertion chains are checked, but Chai's `callCount(0)` and property assertions are not rewritten. Static `expect` methods, `expect.element`, dynamic matcher names, and repeated `not`, `resolves`, or `rejects` modifiers are not checked.

## Incorrect

```ts
import { expect, rs, test } from '@rstest/core';

test('notifies only subscribed listeners', () => {
  const subscribed = rs.fn();
  const unsubscribed = rs.fn();
  subscribed('ready');

  expect(unsubscribed).toHaveBeenCalledTimes(0);
  expect(subscribed).not.toHaveBeenCalledTimes(0);
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';

test('notifies only subscribed listeners', () => {
  const subscribed = rs.fn();
  const unsubscribed = rs.fn();
  subscribed('ready');

  expect(unsubscribed).not.toHaveBeenCalled();
  expect(subscribed).toHaveBeenCalled();
});
```

## Autofix

The fix replaces the count matcher with `toHaveBeenCalled`, removes the zero argument and any explicit type arguments, and adds or removes `.not`. It preserves promise modifiers, accessor quotes, and optional chaining.

`expect.poll` assertions are reported without a fix because the callback can observe the assertion's negation state through `this`.

For other assertion factories, only standalone assertion statements are automatically fixed; parentheses and `await` around the assertion are allowed. Assertions used in assignments, return values (including concise arrow functions), arguments, or other expressions are reported without a fix because the returned assertion object can retain `.not` when reused.

Assertions with extra arguments, comments inside the removed argument or type-argument list, multiple matchers in one chain, or property accesses or calls following the matcher are also reported without a fix. This avoids deleting argument evaluation or comments and changing the assertion state observed later in a chain.
