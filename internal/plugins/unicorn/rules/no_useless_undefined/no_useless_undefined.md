# no-useless-undefined

## Rule details

Disallow redundant uses of `undefined` in `let` initializers, defaults in
parameters and destructuring, `return` and `yield` expressions, arrow bodies,
and trailing arguments in JavaScript calls.

The rule preserves `var` initializers, which can reset an existing binding.
It also preserves `return undefined` when the enclosing TypeScript function
has an explicit return type other than `void` or `undefined`, including
unions such as `number | undefined`.

Examples of **incorrect** code for this rule

```javascript
let value = undefined;
const { name = undefined } = options;
const noop = () => undefined;
function* values() {
  yield undefined;
}
function read(value = undefined) {
  return undefined;
}
read(undefined);
const item = index >= 0 ? items[index] : undefined;
```

Examples of **correct** code for this rule

```javascript
let value;
const { name } = options;
const noop = () => {};
function* values() {
  yield;
}
function read(value) {
  return;
}
read();
const item = items[index];
```

## Fixes and suggestions

The rule can automatically remove redundant initializers, defaults, and
explicit `undefined` expressions. A comment in the text that would be removed
can prevent an autofix.

A conditional such as `index >= 0 ? array[index] : undefined` can offer an
editor suggestion to use `array[index]` directly. Applying it changes when the
indexed access runs, so review the suggestion before applying it.

## Options

### checkArguments

Defaults to `true`. Set it to `false` to allow trailing `undefined` arguments
in JavaScript calls. Calls in TypeScript files keep their arguments, as do
calls to functions matched by name, such as `array.push(undefined)`,
`Math.max(undefined)`, and `String(undefined)`.

### checkArrowFunctionBody

Defaults to `true`. Set it to `false` to allow expression bodies such as
`() => undefined`.

## Differences from upstream

Rslint allows a reference to a locally declared `undefined`. In
`function read(undefined) { return undefined; }`, the return value comes from
the parameter.

The rule applies its call exceptions to statically known computed method
names. For example, it keeps `items["splice"](0, undefined)` because omitting
the second argument would delete the items.

The rule also keeps an explicit `undefined` receiver in
`Reflect.get(obj, "x", undefined)` and an explicit `undefined` new target in
`Reflect.construct(Ctor, [], undefined)`. Omitting these arguments changes
which receiver a getter sees or whether construction throws.

Rslint reports the following TypeScript defaults without an automatic edit.

- An unannotated parameter such as `function f(value = undefined) {}` keeps
  its inferred type. Changing it to `value?` can introduce an implicit `any`.
- A destructured parameter such as `function f({ name }: any = undefined) {}`
  keeps its default because TypeScript rejects an optional destructuring
  parameter in a function implementation.
- A default before a required parameter, as in
  `function f(value: undefined = undefined, name: string) {}`, keeps the
  parameter order valid.
- A constructor parameter property such as
  `constructor(public value: number | undefined = undefined) {}` can still be
  omitted at call sites.

The autofix keeps grouping parentheses on destructuring assignment targets.
For example, `({ key: (target.value) = undefined } = source)` becomes
`({ key: (target.value) } = source)`.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-undefined.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-undefined.js)
