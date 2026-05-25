# no-useless-default-assignment

## Rule Details

Disallow default values that will never be used.

[Default parameters](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/Default_parameters) and [destructuring default values](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Destructuring_assignment#default_value) are only used when the parameter or property is `undefined`. If the source type guarantees a non-`undefined` value, the default is unreachable — at best dead code, at worst a misleading signal about the value's nullability.

Examples of **incorrect** code for this rule:

```typescript
function Bar({ foo = '' }: { foo: string }) {
  return foo;
}

const { foo = '' } = { foo: 'bar' };

const [foo = ''] = ['bar'];

[1, 2, 3].map((a = 42) => a + 1);

function f(a = undefined) {}

const { a = undefined } = {};

function g(p: number | undefined = undefined) {}
```

Examples of **correct** code for this rule:

```typescript
function Bar({ foo = '' }: { foo?: string }) {
  return foo;
}

const { foo = '' } = { foo: undefined };

const [foo = ''] = [undefined];

[1, 2, 3, undefined].map((a = 42) => a + 1);

function f(a?: number) {}

function g(p?: number | undefined) {}
```

## Options

### `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing`

Defaults to `false`. While `false`, the rule emits a top-of-file diagnostic on every file whose `tsconfig.json` does **not** enable `strictNullChecks` (or `strict`). Without `strictNullChecks`, TypeScript erases `undefined` and `null` from types — which makes this rule unable to tell whether a value can be `undefined`, so any per-site report would be unreliable. Set this option to `true` to opt out of that file-level diagnostic and let the rule continue to run anyway.

Examples of code for this rule with `{ "allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true }`:

```json
{ "@typescript-eslint/no-useless-default-assignment": ["error", { "allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true }] }
```

## When Not To Use It

If you use default values defensively against runtime values that bypass type checking, or for documentation purposes, you may want to disable this rule.

## Differences from ESLint

rslint reports a strict superset of what upstream `@typescript-eslint/no-useless-default-assignment` reports — every upstream diagnostic is still emitted, plus the following:

- `const`/`let`/`var` destructuring whose source is a variable reference and whose property is non-optional. Example: `declare const obj: { foo: string }; const { foo = 'd' } = obj;` — rslint reports `foo = 'd'`; upstream is silent on most property names.
- Numeric-key destructuring whose source is a variable reference. Example: `declare const obj: { 1: string }; const { 1: x = 'd' } = obj;` — rslint reports; upstream is silent.

If you only want to match upstream's behavior exactly, ignore these additional reports with an inline disable comment.

## Original Documentation

- [typescript-eslint no-useless-default-assignment](https://typescript-eslint.io/rules/no-useless-default-assignment)
