# unbound-method

## Rule Details

This rule reports class and object methods that are referenced without their receiver. Calling such a reference later can give the method an unintended `this` value. The rule uses TypeScript type information to distinguish methods from bound function properties, functions declared with `this: void`, and built-in functions that can safely be referenced directly.

Rstest-aware exemptions apply when a method reference is the received value of `expect(...)` or `expect.soft(...)` and the built-in matcher inspects the value without invoking it. This includes mock, equality, ordinary snapshot, and Chai-style value assertions. Matchers that invoke the received function are not exempt: `toThrow`, `toThrowError`, Chai `throw` and `throws`, throwing snapshot matchers, and the `rejects` modifier continue to report unbound methods. The `resolves` modifier does not invoke a function subject and follows the matcher behavior.

If the file replaces a built-in matcher with `expect.extend()`, that matcher is not exempt. Equality-based matchers are likewise not exempt after the file calls `expect.addEqualityTesters()`, because the custom matcher or equality tester can execute the received function.

The rule also allows a method reference passed directly to `mocked()` on an Rstest utilities object. It recognizes the `rs` and `rstest` globals, imports and renamed imports from `@rstest/core` or `rstack/test`, module namespaces, CommonJS bindings, and `import.meta.rstest`. Unrelated or locally shadowed objects with the same names are not treated as Rstest utilities.

See the [Rstest expect API](https://rstest.rs/api/runtime-api/test-api/expect) and [mock utilities](https://rstest.rs/api/runtime-api/rstest/mock-functions) for the runtime behavior of these APIs.

## Incorrect

```ts
class Counter {
  count = 0;

  increment() {
    this.count += 1;
    return this.count;
  }
}

const counter = new Counter();

const increment = counter.increment;
expect(counter.increment).toThrow();
```

## Correct

```ts
class Counter {
  count = 0;

  increment() {
    this.count += 1;
    return this.count;
  }
}

const counter = new Counter();
rs.spyOn(counter, 'increment');

const increment = counter.increment.bind(counter);
expect(counter.increment).toHaveBeenCalled();
rs.mocked(counter.increment).mockReturnValue(2);
```

## Options

```json
{
  "rstest/unbound-method": [
    "error",
    {
      "ignoreStatic": true
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `ignoreStatic` | `boolean` | `false` | Skip checks for static methods. |
