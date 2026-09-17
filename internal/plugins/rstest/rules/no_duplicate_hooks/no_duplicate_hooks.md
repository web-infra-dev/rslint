# no-duplicate-hooks

## Rule Details

Disallow registering the same Rstest lifecycle hook more than once in the same lexical suite scope. Duplicate `beforeAll`, `beforeEach`, `afterEach`, or `afterAll` registrations can repeat setup or cleanup and make test organization harder to understand. The second and every later registration of the same hook is reported.

The file-level root and each `describe` call have independent counters. This includes modified and parameterized suites such as `describe.skip`, `describe.runIf`, `describe.each`, and `describe.for`. The rule follows lexical nesting in the source; it does not execute callbacks or helpers.

The rule recognizes Rstest globals; named, renamed, namespace, and CommonJS imports from `@rstest/core`, `rstack/test`, and `@rstest/playwright`; `import.meta.rstest`; stable local aliases; and Playwright `test.beforeEach`-style hooks. Foreign imports, locally shadowed names, dynamic computed members, and invalid hook chains are ignored. Type information is not required. See the [Rstest hooks documentation](https://rstest.rs/api/runtime-api/test-api/hooks) for lifecycle execution details.

## Incorrect

```ts
import { beforeEach, describe } from '@rstest/core';

describe('shopping cart', () => {
  beforeEach(() => resetCart());
  beforeEach(() => seedProducts());
});
```

## Correct

```ts
import { beforeEach, describe } from '@rstest/core';

describe('shopping cart', () => {
  beforeEach(() => {
    resetCart();
    seedProducts();
  });
});
```
