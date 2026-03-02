# no-empty-function

## Rule Details

Disallows empty functions. Empty functions can reduce readability because readers need to guess whether the empty body is intentional. This rule extends the base ESLint `no-empty-function` rule with TypeScript-specific support, including constructors with parameter properties, decorated functions, override methods, and various function types like async functions and generators.

Examples of **incorrect** code for this rule:

```typescript
function foo() {}

const bar = () => {};

class MyClass {
  method() {}
  constructor() {}
}
```

Examples of **correct** code for this rule:

```typescript
function foo() {
  // intentionally empty
}

const bar = () => {
  return;
};

class MyClass {
  constructor(private name: string) {}
}
```

## Original Documentation

- [typescript-eslint no-empty-function](https://typescript-eslint.io/rules/no-empty-function)
