# explicit-function-return-type

## Rule Details

Require explicit return types on functions and class methods.

Functions in TypeScript often don't need to be given an explicit return type annotation. Leaving off the return type is less code to read or write and allows the compiler to infer it from the contents of the function. However, explicit return types do make it visually more clear what type is returned by a function and can speed up TypeScript type checking performance in large codebases.

Examples of **incorrect** code for this rule:

```typescript
function test() {
  return;
}

var fn = function () {
  return 1;
};

var arrowFn = () => 'test';

class Test {
  method() {
    return;
  }
}
```

Examples of **correct** code for this rule:

```typescript
function test(): void {
  return;
}

var fn = function (): number {
  return 1;
};

var arrowFn = (): string => 'test';

class Test {
  method(): void {
    return;
  }
}
```

## Original Documentation

https://typescript-eslint.io/rules/explicit-function-return-type
