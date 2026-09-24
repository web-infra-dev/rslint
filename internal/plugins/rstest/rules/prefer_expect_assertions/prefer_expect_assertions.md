# prefer-expect-assertions

## Rule Details

Require every test to declare how many assertions it expects, by starting with [`expect.assertions(n)`](https://rstest.rs/api/runtime-api/test-api/expect#expectassertions) or [`expect.hasAssertions()`](https://rstest.rs/api/runtime-api/test-api/expect#expecthasassertions). Assertions inside callbacks, loops, and promise handlers may never run, and a test whose assertions are all skipped still passes. Declaring the count makes Rstest fail such a test.

The declaration must be the first statement of the test callback, after any directives such as `'use strict'`, or the body of an arrow function written without braces. A declaration inside a nested block, such as the branch of an `if`, on one side of `&&`, `||`, `??`, or `?:`, after `?.`, or in a destructuring default does not count. The rule also reports `expect.hasAssertions()` called with arguments and `expect.assertions()` called with anything other than a single integer number literal. A declaration that a [`beforeEach`](https://rstest.rs/api/runtime-api/test-api/hooks#beforeeach) callback makes on every run covers every test in the same `describe` block and its nested blocks, wherever the hook appears in the block. It must be a top-level statement of the callback, not in any of those skippable positions, and no earlier statement may `return` or `throw`. The callback can be inline or a function declared in the same file and never reassigned, and it can use the `expect` of the [test context](https://rstest.rs/api/runtime-api/test-api/test#testcontext) it receives. Declarations in `afterEach`, `beforeAll`, and `afterAll` do not count, because Rstest checks the assertion count before `afterEach` runs and resets it before each test. A declaration made later, such as in a timer started by the hook, does not count either.

In a [concurrent test](https://rstest.rs/api/runtime-api/test-api/test#testconcurrent), declare the count on the test context's `expect`. Tests that run at the same time share the imported `expect`, so a declaration made through it, directly or by a `beforeEach` hook, is not guaranteed to apply to the test that made it. The rule does not check which `expect` a concurrent test declares with.

Only tests whose callback is a function written at the registration are checked; a callback passed by reference is not. Both call shapes, `test(name, fn, timeout)` and `test(name, options, fn)`, are recognized, as are `test.each`, `test.for`, chained modifiers, and test APIs built with `test.extend`. The rule recognizes Rstest globals and APIs imported from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, including aliases, namespace and CommonJS imports, `import.meta.rstest`, and the test context's `expect`, destructured or not. Foreign test frameworks, locally shadowed names, and type-only bindings are ignored. Type information is not required.

## Incorrect

```ts
import { expect, test } from '@rstest/core';

test('notifies every subscriber', () => {
  const events = createEventBus();

  events.subscribe((payload) => {
    expect(payload).toEqual({ type: 'saved' });
  });

  events.publish({ type: 'saved' });
});
```

## Correct

```ts
import { expect, test } from '@rstest/core';

test('notifies every subscriber', () => {
  expect.assertions(1);
  const events = createEventBus();

  events.subscribe((payload) => {
    expect(payload).toEqual({ type: 'saved' });
  });

  events.publish({ type: 'saved' });
});
```

## Options

```json
{
  "rstest/prefer-expect-assertions": [
    "error",
    {
      "onlyFunctionsWithAsyncKeyword": true,
      "onlyFunctionsWithExpectInCallback": true,
      "disallowHasAssertions": true
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `onlyFunctionsWithAsyncKeyword` | `boolean` | `false` | Check tests whose callback is declared `async`. |
| `onlyFunctionsWithExpectInLoop` | `boolean` | `false` | Check tests that call `expect` inside a `for`, `for...in`, `for...of`, `while`, or `do...while` loop of the test callback. |
| `onlyFunctionsWithExpectInCallback` | `boolean` | `false` | Check tests that call `expect` inside a function nested in the test callback. |
| `disallowHasAssertions` | `boolean` | `false` | Report `expect.hasAssertions()` and suggest `expect.assertions(n)` instead. |

With none of the `onlyFunctions*` options enabled, every test is checked. When any of them is enabled, a test is checked if it matches at least one enabled option. A loop that encloses the test registration rather than the `expect` call does not count.

## Suggestions

For a missing declaration, the rule suggests inserting `expect.hasAssertions();` or `expect.assertions();` at the start of the test callback, after any directives. The number of assertions has to be filled in. When `disallowHasAssertions` is enabled, only `expect.assertions();` is suggested. The inserted call uses the same `expect` as the test: a destructured test-context `expect`, then the test context of a concurrent test, then the `expect` the file imports, including an alias or a namespace member, then the test context's `expect`. A `require` binding that the file reassigns is not used. The global `expect` is used only in files that import nothing from Rstest and bind no `expect` of their own, including through destructuring. No suggestion is offered when none of these is available, or when a function around the test declares a variable with the same name. An arrow function written without braces gets no suggestion.

Extra arguments to `expect.hasAssertions()` or `expect.assertions()` have a suggestion that removes them. With `disallowHasAssertions`, `expect.hasAssertions()` has a suggestion that renames it to `expect.assertions()`, keeping the quotes of a bracket access.
