# prefer-snapshot-hint

## Rule Details

Prefer descriptive hints for external snapshots so their names explain what each snapshot captures. By default, hints are required when a group contains more than one external snapshot assertion. See [Rstest snapshot naming](https://rstest.rs/guide/basic/snapshot#use-descriptive-snapshot-names).

The rule checks `toMatchSnapshot`, its `matchSnapshot` alias, and `toThrowErrorMatchingSnapshot`, including Chai assertion chains, `expect.soft`, and `resolves`/`rejects`. It recognizes globals, named and renamed imports, namespace imports, CommonJS imports, `rstack/test`, `import.meta.rstest`, and test-context `expect`, including parameterized and extended tests. The Rstest assertions exposed by `@rstest/playwright` are also checked. Locally shadowed or reassigned assertion bindings are ignored. Static bracket access and optional calls are recognized. Inline snapshots, file snapshots, static `expect` methods, negated snapshots, and `expect.poll`/`expect.element` chains are excluded. An assertion wrapped with `as`, `satisfies`, or a non-null assertion before the matcher access is not inspected.

Groups follow function expressions, arrow functions, methods, and accessors, including nested helper expressions. Each test or suite registration starts a new grouping boundary. Function declarations do not create a separate group, and helper calls are not followed to count runtime assertions. A hinted snapshot still contributes to its group's count.

For `toMatchSnapshot` and `matchSnapshot`, a single string literal or template literal counts as a hint; two arguments count as properties plus a hint. A single variable or other expression is treated as properties, even if its type is `string`. `toThrowErrorMatchingSnapshot` accepts one hint argument. The rule does not validate hint contents or argument types and does not require type information.

## Incorrect

```ts
import { expect, test } from '@rstest/core';

test('cart totals', () => {
  expect(cart.subtotal).toMatchSnapshot();
  expect(cart.total).toMatchSnapshot();
});
```

## Correct

```ts
import { expect, test } from '@rstest/core';

test('cart totals', () => {
  expect(cart.subtotal).toMatchSnapshot('before tax');
  expect(cart.total).toMatchSnapshot('after tax');
});
```

## Options

```json
{
  "rstest/prefer-snapshot-hint": ["error", "always"]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| mode | string | `"multi"` | Use `"always"` to require hints for every external snapshot, or `"multi"` for groups with multiple external snapshots. |
