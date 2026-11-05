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

This check follows return inference through immediately invoked functions. It also preserves generators without an explicit return type: both yielded values and the result of a `yield` expression can receive contextual types.

The check is conservative: it can retain a redundant assertion when a generic return type could contain type variables, even if other arguments determine those variables. Contextual fixes in inferred generic calls require an unused result and a single input path, without competing properties, elements, arguments, or callback parameters. The rule still checks calls with explicit type arguments (including trailing defaults) and receivers with a shared parameter type. Overloads with a single accepted argument and identical return types can also be checked when they have no type predicates or explicit `this` parameters. Assertions that may provide contextual types to function parameters are preserved.

When deciding whether a type is unchanged, Rslint uses the compiler's type identity relation instead of treating mutual assignability as equality. This preserves distinctions between enums and their primitive base types, including within object properties and function signatures, while reporting redundant assertions between equivalent aliases. An assertion from an enum to `number` can intentionally widen the static type to allow comparisons outside the enum's members. Assertions whose receiving context safely accepts the original type can still be reported.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unnecessary-type-assertion.ts)
