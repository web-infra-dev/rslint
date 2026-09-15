# max-nested-describe

## Rule Details

Limits the depth of nested [`describe`](https://rstest.rs/api/runtime-api/test-api/describe) calls. Shallow suite hierarchies keep test names, reports, and inherited suite setup easier to understand.

The rule recognizes Rstest globals and APIs imported from `@rstest/core`, `rstack/test`, and `@rstest/playwright`, including aliases, namespace and CommonJS imports, `import.meta.rstest`, Playwright `test.describe`, parameterized suites, conditional suites, chained suite modifiers, and suite callbacks passed by reference. It does not treat similarly named APIs from other test frameworks or locally shadowed functions as Rstest suites.

## Incorrect

```ts
describe('checkout', () => {
  describe('signed-in customer', () => {
    describe('saved payment method', () => {
      test('places the order', () => {});
    });
  });
});
```

## Correct

```ts
describe('checkout for a signed-in customer', () => {
  describe('saved payment method', () => {
    test('places the order', () => {});
  });
});
```

## Options

```json
{
  "rstest/max-nested-describe": [
    "error",
    {
      "max": 2
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `max` | `integer` | `5` | Set the maximum allowed suite nesting depth. |
