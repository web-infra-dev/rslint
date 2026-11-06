# no-unnecessary-type-assertion

## Rule Details

This rule reports unnecessary assertions, including `as`, angle-bracket assertions, and non-null assertions (`!`).

Examples of **incorrect** code for this rule:

```typescript
const foo = 3;
const bar = foo!; // foo is already non-nullable

declare const value: number;
const num = value as number; // already a number
```

Examples of **correct** code for this rule:

```typescript
const foo: number | undefined = getValue();
const bar = foo!; // non-null assertion is meaningful

const value = someUnknown as string; // type narrowing

const x = 3 as const;
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| `typesToIgnore` | `[]` | Skip `as` and angle-bracket assertions for the listed type names. |
| `checkLiteralConstAssertions` | `false` | Also check for unnecessary `as const` assertions on literal values. |

## Differences from upstream

Rslint keeps assertions that your code relies on. This helps avoid new type errors, lost type information, or changes in behavior after running `--fix`.

Some assertions are reported without an automatic fix, so their diagnostics remain after `--fix`. This includes certain assertions split across lines with comments. The rule leaves these assertions in place to preserve your program's behavior.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unnecessary-type-assertion.ts)
