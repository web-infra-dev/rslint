# no-meaningless-void-operator

## Rule Details

Disallows the `void` operator when its argument is already of type `void` or `undefined`. The `void` operator is intended to convey that a return value is deliberately being ignored. Using it on an expression that already evaluates to `void` or `undefined` is redundant and misleading. Optionally, the rule can also check for `void` on `never`-typed expressions.

Examples of **incorrect** code for this rule:

```typescript
void undefined;

void console.log('hello');

function foo(): void {}
void foo();
```

Examples of **correct** code for this rule:

```typescript
void Promise.resolve();

void someAsyncFunction();

console.log('hello');

foo();
```

## Original Documentation

- [typescript-eslint no-meaningless-void-operator](https://typescript-eslint.io/rules/no-meaningless-void-operator)
