# no-unsafe-call

## Rule Details

Disallow calling a value with type `any`.

Calling an `any`-typed value as a function or constructor is unsafe since there is no guarantee that the value is actually callable. This rule also flags tagged template expressions and `new` expressions with `any`-typed callee values, as well as calling values typed as the `Function` type.

Examples of **incorrect** code for this rule:

```typescript
declare const anyVal: any;
anyVal();
anyVal.foo();
new anyVal();

declare const fn: Function;
fn();
```

Examples of **correct** code for this rule:

```typescript
declare const greet: (name: string) => void;
greet('world');

declare const Cls: new () => object;
new Cls();

declare const tag: (strings: TemplateStringsArray) => string;
tag`hello`;
```

## Original Documentation

- [typescript-eslint no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call)
