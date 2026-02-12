# consistent-return

## Rule Details

Require `return` statements to either always or never specify values. This is the TypeScript-enhanced version of the ESLint `consistent-return` rule. It uses type information to allow valid return patterns for functions with `void`, `undefined`, or `Promise<void>` return types.

A function with inconsistent return statements (some returning a value and some not) is typically a mistake.

Examples of **incorrect** code for this rule:

```typescript
function foo(flag: boolean): string | undefined {
  if (flag) {
    return 'hello';
  }
  return;
}

function bar(x: number) {
  if (x > 0) {
    return x;
  }
  return;
}
```

Examples of **correct** code for this rule:

```typescript
function foo(flag: boolean): string | undefined {
  if (flag) {
    return 'hello';
  }
  return undefined;
}

function bar(): void {
  if (Math.random() > 0.5) {
    return;
  }
  console.log('done');
}
```

## Original Documentation

- [typescript-eslint consistent-return](https://typescript-eslint.io/rules/consistent-return)
