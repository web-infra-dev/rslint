# prefer-ending-with-an-expect

## Rule Details

Require the last statement a test case runs to be an assertion. A test that ends by calling into the code under test, changing state, or awaiting a promise has exercised something without checking the result, which is usually an unfinished test rather than a deliberate one. Only the final statement is examined; assertions earlier in the body do not satisfy the rule.

A block body contributes its last statement, an arrow function written without braces contributes its expression, and a leading `await` is looked through. `expect`, `expect.soft`, `expect.poll`, and Chai's `assert` all count as assertions, including when they reach the call site through the [test context](https://rstest.rs/api/runtime-api/test-api/test-context), a namespace import, a destructured fixture, or a renamed import. `assert` is recognized through every one of those except the test context, which Rstest does not use to provide it. Additional asserting functions can be configured with `assertFunctionNames`.

Chai asserts through property getters as well as calls, so a test ending in `expect(cart).to.be.empty` or `context.expect(ok).to.be.true` satisfies the rule. A chain that resolves no matcher still does not: `expect(cart).not` stops on a modifier and `expect(cart).toBe` never invokes its matcher, and neither asserts anything.

A destructured key is read when it names a constant, so `const { ['expect']: check } = import.meta.rstest` is recognized exactly like `const { expect: check }`. A key computed from a value that is only known at runtime, such as `const { [key]: check } = import.meta.rstest`, names no API the rule can identify; a test whose last statement asserts through such a binding is reported even though it does assert. Bind the API under a readable key to avoid this.

Both Rstest call shapes are recognized — `test(name, fn, timeout)` and `test(name, options, fn)` — so a case that passes its options object second is checked like any other. `test.todo` is exempt because the function it is given never runs, and a callback passed by reference rather than written inline is left alone. The rule recognizes Rstest globals and APIs imported from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, including aliases, namespace and CommonJS imports, `import.meta.rstest`, parameterized cases such as `test.each` and `test.for`, chained modifiers such as `test.concurrent`, and fixture-extended test APIs built with `test.extend`. Foreign test frameworks, locally shadowed names, and type-only bindings are ignored. Type information is not required.

## Incorrect

```ts
import { expect, test } from '@rstest/core';

test('lets me change the selected option', () => {
  const container = render(MySelect, { selected: 1 });

  expect(container.toHTML()).toContain('value="1" selected');

  container.setProp('selected', 2);
});
```

## Correct

```ts
import { expect, test } from '@rstest/core';

test('lets me change the selected option', () => {
  const container = render(MySelect, { selected: 1 });

  expect(container.toHTML()).toContain('value="1" selected');

  container.setProp('selected', 2);

  expect(container.toHTML()).toContain('value="2" selected');
});
```

## Options

```json
{
  "rstest/prefer-ending-with-an-expect": [
    "error",
    {
      "assertFunctionNames": ["expect", "assert", "expectSaga"],
      "additionalTestBlockFunctions": ["scenario"]
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `assertFunctionNames` | `string[]` | `["expect", "assert"]` | Function names or patterns treated as assertions. A name may use `*` to match within one dotted segment and `**` to span segments, for example `request.**.expect`. Setting it replaces the default list. |
| `additionalTestBlockFunctions` | `string[]` | `[]` | Additional functions whose callbacks are treated as test blocks. |
