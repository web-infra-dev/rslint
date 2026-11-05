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

Rslint keeps assertions that help TypeScript infer a return type or select a function overload. For example, the assertion below tells TypeScript that `check` accepts and returns a `string`. Removing it would make `check('hello')` return `unknown`, so the call to `toUpperCase()` would fail to compile.

```typescript
declare function makeCheck<T>(): (value: T) => T;
const check = makeCheck() as (value: string) => string;
check('hello').toUpperCase();
```

This also matters inside callbacks. Returning an untyped value `as Model` can give the caller a `Model` result. Removing the assertion can change that result to `any`.

An assertion can also widen a numeric enum. Here `as number` makes `values` a `number[]`. Removing the assertion gives it the type `Mode[]`, and TypeScript rejects `values.includes(-1)` because `-1` is outside the enum.

```typescript
enum Mode {
  Auto = 1,
  Manual = 2,
}
declare const mode: Mode;
const values = Array.of(mode as number);
values.includes(-1);
```

An assertion on an object can also keep a property narrow. In `const row = { ...{ value: 1 } } as { value: 1 }`, `row.value` has type `1`. Removing the assertion widens it to `number`. Similarly, if `tag` is a unique symbol, `const row = { tag } as { tag: typeof tag }` keeps `row.tag` specific to that symbol. Removing the assertion would widen the property to `symbol`.

With `checkLiteralConstAssertions` enabled, `--fix` can simplify `const count: 1 = 1 as const` to `const count: 1 = 1`. In `const count = 1 as const`, it keeps the assertion. A later `let copy = count` has type `1`. Removing `as const` would give `copy` the type `number`.

If `make<T>(): T` is a generic function, `(make() satisfies string) as string` simplifies to `(make() satisfies string)`. The `satisfies string` clause already gives the call its `string` type.

`--fix` may leave a redundant assertion in place when its effect on a generic call or a generator's inferred type is uncertain. If `A` and `B` are both aliases for `Mode | null`, `--fix` can remove `as B` from a value of type `A`.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-assertion](https://typescript-eslint.io/rules/no-unnecessary-type-assertion)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.68.0/packages/eslint-plugin/src/rules/no-unnecessary-type-assertion.ts)
