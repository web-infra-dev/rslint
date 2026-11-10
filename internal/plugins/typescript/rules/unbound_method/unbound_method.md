# unbound-method

## Rule Details

Enforce unbound methods are called with their expected scope. Extracting a class method as a standalone variable or callback without binding it to the class instance causes `this` to be `undefined` at runtime, which is a common source of bugs. This rule reports when a method is referenced without being called, unless it is in a safe context (like a comparison, typeof check, or conditional check).

If the method does not access `this`, the rule additionally suggests annotating the method parameter with `this: void` or converting it to an arrow function.

Examples of **incorrect** code for this rule:

```typescript
class MyClass {
  method() {
    return this.value;
  }
}
const instance = new MyClass();
const unboundMethod = instance.method;
[1, 2, 3].forEach(instance.method);
```

Examples of **correct** code for this rule:

```typescript
class MyClass {
  method() {
    return this.value;
  }
}
const instance = new MyClass();
const boundMethod = instance.method.bind(instance);
instance.method();
if (instance.method) {
}
typeof instance.method;
```

## Differences from upstream

For a receiver union mixing a function-valued class field with a built-in method,
TypeScript-Go and TypeScript can visit the union constituents in a different order:

```typescript
class Foo {
  floor = function () {};
}
const { floor } = Math.random() > 0.5 ? new Foo() : Math;
```

Rslint reports `unbound` for the class field. typescript-eslint v8.28.0 reports
`unboundWithoutThisAnnotation` after visiting `Math` first. Both report the same
location. Rslint retains its function-field message because adding `this: void`
does not exempt a function-valued class field from this rule.

## Original Documentation

- [typescript-eslint: unbound-method](https://typescript-eslint.io/rules/unbound-method)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.28.0/packages/eslint-plugin/src/rules/unbound-method.ts)
