# no-inferrable-types

## Rule Details

Disallows explicit type declarations for variables or parameters initialized to a number, string, or boolean.

TypeScript is able to infer the types of parameters, properties, and variables from their default or initial values. There is no need to use an explicit type annotation for trivially inferred types (boolean, bigint, number, null, RegExp, string, symbol, undefined).

Examples of **incorrect** code for this rule:

```typescript
const a: bigint = 10n;
const a: bigint = -10n;
const a: bigint = BigInt(10);
const a: boolean = true;
const a: boolean = false;
const a: boolean = Boolean(null);
const a: boolean = !0;
const a: number = 10;
const a: number = +10;
const a: number = -10;
const a: number = Number('1');
const a: number = Infinity;
const a: number = NaN;
const a: null = null;
const a: RegExp = /a/;
const a: RegExp = RegExp('a');
const a: RegExp = new RegExp('a');
const a: string = 'str';
const a: string = `str`;
const a: string = String(1);
const a: symbol = Symbol('a');
const a: undefined = undefined;
const a: undefined = void 0;

function fn(a: number = 5) {}
const fn = (a: boolean = true) => {};

class Foo {
  prop: number = 5;
}
```

Examples of **correct** code for this rule:

```typescript
const a = 10n;
const a = true;
const a = 'str';
const a = null;
const a = /a/;
const a = undefined;
const a = Symbol('a');

function fn(a = 5) {}
const fn = (a = true) => {};

class Foo {
  prop = 5;
}

// Readonly properties are allowed
class Bar {
  readonly prop: number = 5;
}
```

## Options

### `ignoreParameters`

When set to `true`, ignores explicit type annotations on function parameters with default values.

```json
{
  "@typescript-eslint/no-inferrable-types": [
    "warn",
    { "ignoreParameters": true }
  ]
}
```

### `ignoreProperties`

When set to `true`, ignores explicit type annotations on class properties with initializers.

```json
{
  "@typescript-eslint/no-inferrable-types": [
    "warn",
    { "ignoreProperties": true }
  ]
}
```

## Original Documentation

https://typescript-eslint.io/rules/no-inferrable-types
