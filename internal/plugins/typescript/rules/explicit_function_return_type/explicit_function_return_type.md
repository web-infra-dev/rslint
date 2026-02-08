# explicit-function-return-type

## Rule Details

Require explicit return types on functions and class methods.

Examples of **incorrect** code for this rule:

```typescript
function foo() {
  return;
}

const bar = () => 'value';

class Test {
  method() {
    return 1;
  }
}
```

Examples of **correct** code for this rule:

```typescript
function foo(): void {
  return;
}

const bar = (): string => 'value';

class Test {
  method(): number {
    return 1;
  }
}
```

## Original Documentation

https://typescript-eslint.io/rules/explicit-function-return-type
