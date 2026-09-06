# rstest/prefer-called-with

## Rule Details

Prefer checking the arguments passed to a mock instead of only checking whether it was called. This rule reports `toBeCalled()`, `toHaveBeenCalled()`, and `toHaveBeenCalledOnce()`, recommending `toBeCalledWith()`, `toHaveBeenCalledWith()`, and `toHaveBeenCalledExactlyOnceWith()` respectively. Assertions negated with `.not` are exempt.

The rule recognizes global Rstest `expect`, imports and aliases from `@rstest/core`, namespace imports, `const` namespace receivers from `require` or `import.meta.rstest`, destructured imports, and Test Context `expect`. Namespace receivers declared with `let` or `var` are ignored. It also checks assertion factories such as `expect.soft` and `expect.poll`. It does not require type information, though import aliases and local shadowing are resolved when a TypeScript project is available.

Only the first matcher in an assertion chain is checked, and it must be called. Chai language chains before that matcher are allowed; Chai property assertions such as `.called` and `.calledOnce`, later matchers, and static `expect` APIs are ignored. Computed matcher names must be string or template literals without substitutions. The rule does not validate which matchers a factory allows; in particular, it does not make mock matchers available on `expect.element`.

## Incorrect

```ts
import { expect, rs, test } from '@rstest/core';

test('notifies the customer', () => {
  const notify = rs.fn();
  notify('order-confirmed');

  expect(notify).toHaveBeenCalled();
  expect(notify).toHaveBeenCalledOnce();
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';

test('notifies the customer', () => {
  const notify = rs.fn();
  notify('order-confirmed');

  expect(notify).toHaveBeenCalledWith('order-confirmed');
  expect(notify).toHaveBeenCalledExactlyOnceWith('order-confirmed');
});
```

## Autofix

The fix replaces only the matcher name, preserving the factory, modifiers, optional calls, arguments, comments, and the quote style of computed accessors. If the matcher accessor cannot be safely located, the rule reports without a fix.

The fix does not infer or insert expected arguments. Add them yourself after applying it: an empty `toHaveBeenCalledWith()` checks for a call with no arguments, and an empty `toHaveBeenCalledExactlyOnceWith()` checks for exactly one call with no arguments.
