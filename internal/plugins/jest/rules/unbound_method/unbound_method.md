# unbound-method

## Rule Details

Enforce unbound methods are called with their expected scope. This rule extends [`@typescript-eslint/unbound-method`](https://typescript-eslint.io/rules/unbound-method) with an understanding of Jest: a method may be passed unbound to `expect` or `jest.mocked`, which never call it.

Every other check of the base rule is kept. Methods passed to `expect` are still reported when the matcher calls them (`toThrow`, `toThrowError`, `toThrowErrorMatchingSnapshot`, and `toThrowErrorMatchingInlineSnapshot`), or when the `expect` chain has no called matcher.

Use this rule in place of `@typescript-eslint/unbound-method` for test files; enabling both reports the same references twice. The rule requires type information.

Examples of **incorrect** code for this rule:

```ts
class Service {
  method() {
    return this;
  }
}
const service = new Service();

expect(service.method).toThrow();
Promise.resolve().then(service.method);
const method = service.method;
```

Examples of **correct** code for this rule:

```ts
class Service {
  method() {
    return this;
  }
}
const service = new Service();

expect(service.method).toHaveBeenCalledTimes(1);
expect(service.method).not.toHaveBeenCalled();
jest.mocked(service.method).mockImplementation(() => service);
```

## Options

This rule accepts the same options as [`@typescript-eslint/unbound-method`](https://typescript-eslint.io/rules/unbound-method#options), such as `ignoreStatic`.

```json
{ "jest/unbound-method": ["error", { "ignoreStatic": true }] }
```

## Original Documentation

- [eslint-plugin-jest: unbound-method](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/unbound-method.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/unbound-method.ts)
