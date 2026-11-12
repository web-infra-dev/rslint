# consistent-return

## Rule Details

Require `return` statements to either always or never specify values. This is the TypeScript-enhanced version of the ESLint `consistent-return` rule. It uses type information to ignore bare `return` statements when a synchronous function has a return type containing `void`, or an async function returns a thenable whose resolved type contains `void`. This includes `number | void` and `Promise<number | void>`. A return type of `undefined` alone does not grant this exemption.

The first checked `return` determines whether later returns must specify a value. The rule reports each conflicting return statement and reachable function endings that omit a required value. With `treatUndefinedAsUnspecified: true`, `undefined` identifiers and `void` expressions count as unspecified values, as do expressions whose type is exactly `undefined`. A union type containing `undefined` does not by itself make an expression unspecified.

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

- [typescript-eslint: consistent-return](https://typescript-eslint.io/rules/consistent-return)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/consistent-return.ts)
