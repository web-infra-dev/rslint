# no-unnecessary-type-assertion

## Rule Details

This rule reports assertions you can remove while keeping the same type where the value is used. It checks `as` assertions, angle-bracket assertions, and `!` on values that are already non-nullable.

The rule reports these assertions.

```typescript
const foo = 3;
const bar = foo!; // foo is already non-nullable

declare const value: number;
const num = value as number; // already a number
```

The rule keeps these assertions with the default options.

```typescript
const foo: number | undefined = getValue();
const bar = foo!; // non-null assertion is meaningful

const value = someUnknown as string; // type narrowing

const x = 3 as const;
```

## Differences from upstream

Rslint keeps assertions that your code relies on. This helps avoid new type errors, lost type information, or changes in behavior after running `--fix`.

Some assertions are reported without an automatic fix, so their diagnostics remain after `--fix`. This includes certain assertions split across lines with comments. The rule leaves these assertions in place to preserve your program's behavior.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unnecessary-type-assertion.ts)
