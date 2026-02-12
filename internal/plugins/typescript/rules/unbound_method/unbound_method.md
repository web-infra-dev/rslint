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

## Original Documentation

- [typescript-eslint unbound-method](https://typescript-eslint.io/rules/unbound-method)
