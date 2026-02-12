# no-duplicate-type-constituents

## Rule Details

Disallows duplicate constituents in union or intersection types. Having the same type more than once in a union (`|`) or intersection (`&`) is redundant and can be removed without changing the type. This rule also flags explicit `undefined` on optional parameters, since the `?` modifier already implies `undefined`.

Examples of **incorrect** code for this rule:

```typescript
type Foo = string | string;

type Bar = number & number;

type Baz = 'a' | 'b' | 'a';

function fn(x?: string | undefined) {}
```

Examples of **correct** code for this rule:

```typescript
type Foo = string | number;

type Bar = { a: string } & { b: number };

type Baz = 'a' | 'b' | 'c';

function fn(x?: string) {}
```

## Original Documentation

- [typescript-eslint no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents)
