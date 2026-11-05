# no-unnecessary-type-assertion

## Rule Details

Disallow type assertions that do not change the type of an expression.

Type assertions (`as` expressions, angle-bracket syntax, and non-null assertions `!`) that don't actually change the type of an expression are unnecessary and add noise to the code. This includes non-null assertions on values that are already non-nullable.

Examples of **incorrect** code for this rule:

```typescript
const foo = 3;
const bar = foo!; // foo is already non-nullable

const str = 'hello' as string; // already a string

declare const value: number;
const num = value as number; // already a number
```

Examples of **correct** code for this rule:

```typescript
const foo: number | undefined = getValue();
const bar = foo!; // non-null assertion is meaningful

const value = someUnknown as string; // type narrowing

const x = 3 as const; // const assertions are allowed
```

## Differences from upstream

Rslint applies an additional conservative check before removing assertions that may supply contextual inference for a generic call, constructor, or tagged template. The expression's type inside an assertion can already include inferences from the asserted type, including through an intervening `as const` assertion. Removing the assertion can change the inferred type or the selected overload.

```typescript
declare function makeCheck<T>(): (value: T) => T;
export const check = makeCheck() as (value: string) => string;
// Without the assertion, check accepts and returns unknown.
```

This check is conservative: it can retain a redundant assertion when a generic return type could contain type variables, even if other arguments determine those variables. The rule still checks non-generic calls, calls with explicit type arguments (including trailing defaults), and generic calls whose return type is known not to contain type variables. Assertions that may provide contextual types to function parameters are also preserved.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unnecessary-type-assertion.ts)
