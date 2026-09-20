# prefer-ending-with-an-expect

## Rule Details

Require the last statement a test case runs to be an assertion. A test that ends by calling into the code under test, changing state, or awaiting a promise has exercised something without checking the result, which is usually an unfinished test rather than a deliberate one. Only the final statement is examined; assertions earlier in the body do not satisfy the rule.

A block body contributes its last statement, an arrow function written without braces contributes its expression, and a leading `await` is looked through. `expect`, `expect.soft`, `expect.poll`, and Chai's `assert` all count as assertions, including when they reach the call site through the [test context](https://rstest.rs/api/runtime-api/test-api/test-context), a namespace import, a destructured fixture, or a renamed import. Additional asserting functions can be configured with `assertFunctionNames`.

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
