# strict-boolean-expressions

## Rule Details

Disallow certain types in boolean expressions. Forbids usage of non-boolean types in the following boolean positions:

- the argument of the logical-negation operator (`!arg`)
- the test expression of a conditional (`cond ? x : y`)
- the condition of an `if`, `for`, `while`, or `do-while` statement
- the operands of the logical AND/OR operators (`&&`, `||`)
- the argument of a truthiness-assertion function (`function f(x): asserts x; f(arg)`)
- the return value of an array predicate callback (`arr.filter(cb)`, `arr.some(cb)`, …)

The `boolean` and `never` types are always allowed. Every other type reports unless the matching `allow*` option enables it.

When a non-boolean value is reported, the rule emits one or more suggestion fixes appropriate to the value's type. For example a `string` value gets `value.length > 0`, `value !== ""`, and `Boolean(value)` suggestions; a nullable number gets `value != null`, `value ?? 0`, and `Boolean(value)`; an array-predicate callback whose return type is non-boolean gets an `: boolean` return-type-annotation suggestion in addition to the standard conversion fixes.

Examples of **incorrect** code for this rule:

```typescript
declare const num: number | undefined;
if (num) {
  console.log('defined');
}

declare const str: string | null;
if (!str) {
  console.log('empty');
}

function foo(bool?: boolean) {
  if (bool) {
    bar();
  }
}

const foo = <T>(arg: T) => (arg ? 1 : 0);
```

Examples of **correct** code for this rule:

```typescript
declare const num: number | undefined;
if (num != null) {
  console.log('defined');
}

declare const str: string | null;
if (str != null && str !== '') {
  console.log('non-empty');
}

function foo(bool?: boolean) {
  if (bool ?? false) {
    bar();
  }
}

const foo = (arg: any) => (Boolean(arg) ? 1 : 0);
```

## Options

### `allowString`

Default: `true`. When `true`, allow `string` values in boolean expressions.

Examples of **incorrect** code with `{ "allowString": false }`:

```json
{ "@typescript-eslint/strict-boolean-expressions": ["error", { "allowString": false }] }
```

```typescript
declare const x: string;
if (x) {
}
```

### `allowNumber`

Default: `true`. When `true`, allow `number` and `bigint` values in boolean expressions.

Examples of **incorrect** code with `{ "allowNumber": false }`:

```json
{ "@typescript-eslint/strict-boolean-expressions": ["error", { "allowNumber": false }] }
```

```typescript
declare const x: number;
if (x) {
}
```

### `allowNullableObject`

Default: `true`. When `true`, allow nullable object values — for example `object`, `symbol`, or function types in a union with `null` or `undefined`.

### `allowNullableBoolean`

Default: `false`. When `true`, allow nullable boolean values — `boolean` in a union with `null` or `undefined`.

Examples of **correct** code with `{ "allowNullableBoolean": true }`:

```json
{ "@typescript-eslint/strict-boolean-expressions": ["error", { "allowNullableBoolean": true }] }
```

```typescript
declare const x: boolean | null;
if (x) {
}
```

### `allowNullableString`

Default: `false`. When `true`, allow nullable string values.

### `allowNullableNumber`

Default: `false`. When `true`, allow nullable number values.

### `allowNullableEnum`

Default: `false`. When `true`, allow nullable enum values.

### `allowAny`

Default: `false`. When `true`, allow `any`, `unknown`, and unconstrained generic values.

### `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing`

Default: `false`. By default the rule emits a file-level `noStrictNullCheck` diagnostic when `strictNullChecks` is off because the rule's output is unreliable without it. Set this to `true` to silence the diagnostic and run the rule anyway.

## When Not To Use It

If your codebase does not rely on JavaScript truthiness coercion in boolean positions, or you prefer the conciseness of `if (x)` over the strictness of `if (x != null)`, you can disable this rule.

## Original Documentation

- [typescript-eslint strict-boolean-expressions](https://typescript-eslint.io/rules/strict-boolean-expressions)
