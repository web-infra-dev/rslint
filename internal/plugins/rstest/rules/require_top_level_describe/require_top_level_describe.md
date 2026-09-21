# require-top-level-describe

## Rule Details

Require every test case and every lifecycle hook to be registered inside a [`describe`](https://rstest.rs/api/runtime-api/test-api/describe) block. A file whose cases all belong to a suite reports under one heading, and its [hooks](https://rstest.rs/api/runtime-api/test-api/hooks) apply to a named group instead of the whole file, which makes it clear what the setup and teardown cover. Test cases and hooks written directly at file scope are reported.

A suite counts as top level when nothing else runs it. The number of top-level suites is unlimited unless `maxNumberOfTopLevelDescribes` is set, and that option never restricts suites nested inside another suite. Every top-level suite past the limit is reported.

The rule recognizes Rstest globals and APIs imported from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, including aliases, namespace and CommonJS imports, `import.meta.rstest`, Playwright `test.describe` and `test.beforeEach`, parameterized cases and suites such as `test.each`, `test.for`, `describe.each`, and `describe.for`, conditional and chained modifiers such as `test.concurrent`, `test.todo`, and `describe.runIf`, and fixture-extended test APIs built with `test.extend`. Suite membership follows callbacks passed by reference as well as inline callbacks, so a case registered inside a function that a `describe` runs is treated as belonging to that suite. Foreign test frameworks, locally shadowed names, type-only bindings, and API factories that register nothing — `test.extend({})` on its own, or a bare `describe.each([1])` — are ignored. `onTestFinished` and `onTestFailed` run inside a test body rather than a suite, so they are not hooks for this rule. Type information is not required.

## Incorrect

```ts
import { beforeEach, test } from '@rstest/core';

beforeEach(() => resetCart());

test('places the order', () => {
  expect(checkout()).toBe('ok');
});
```

## Correct

```ts
import { beforeEach, describe, test } from '@rstest/core';

describe('checkout', () => {
  beforeEach(() => resetCart());

  test('places the order', () => {
    expect(checkout()).toBe('ok');
  });
});
```

## Options

```json
{
  "rstest/require-top-level-describe": [
    "error",
    {
      "maxNumberOfTopLevelDescribes": 1
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `maxNumberOfTopLevelDescribes` | `integer` | `none` | Set the maximum number of top-level `describe` blocks allowed in one file. |
