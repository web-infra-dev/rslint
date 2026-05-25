# prefer-nullish-coalescing

Enforce using the nullish coalescing operator instead of logical assignments or chaining.

## Rule Details

The `??` nullish coalescing runtime operator allows providing a default value when dealing with `null` or `undefined`. Because the nullish coalescing operator only coalesces when the original value is `null` or `undefined`, it is much safer than relying upon logical OR operator chaining `||`, which coalesces on any falsy value.

This rule reports when `||`, `||=`, conditional expressions, and `if`-statement assignments could be replaced with the nullish-coalescing operator.

This rule requires `strictNullChecks` to be enabled in `tsconfig.json` to function correctly.

Examples of **incorrect** code for this rule:

```typescript
declare const a: string | null;
declare const b: string | null;

a || b;
a || 'fallback';
a ||= b;

a !== undefined && a !== null ? a : 'a string';
a === undefined || a === null ? 'a string' : a;

declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
if (!foo) {
  foo = makeFoo();
}
```

Examples of **correct** code for this rule:

```typescript
declare const a: string | null;
declare const b: string | null;

a ?? b;
a ?? 'fallback';
a ??= b;

declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };
foo ??= makeFoo();
```

## Options

### `ignoreConditionalTests`

Default: `true`. When `true`, ignore cases that appear inside the test of `if`/`while`/`do…while`/`for` loops or in the test of a conditional expression.

```json
{ "@typescript-eslint/prefer-nullish-coalescing": ["error", { "ignoreConditionalTests": false }] }
```

### `ignoreTernaryTests`

Default: `false`. When `true`, ignore ternary expressions that could be replaced with `??`.

### `ignoreIfStatements`

Default: `false`. When `true`, ignore `if` statements that could be replaced with `??=`.

### `ignoreMixedLogicalExpressions`

Default: `false`. When `true`, ignore `||` expressions that are part of a mixed logical expression (with `&&`).

### `ignoreBooleanCoercion`

Default: `false`. When `true`, ignore `||` arguments to the global `Boolean` constructor.

### `ignorePrimitives`

Default: `{ bigint: false, boolean: false, number: false, string: false }`. Set to `true` to ignore all listed primitives, or to a partial object to ignore individual primitive types.

```json
{ "@typescript-eslint/prefer-nullish-coalescing": ["error", { "ignorePrimitives": { "string": true } }] }
```

### `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing`

Default: `false`. By default the rule errors on every file when `strictNullChecks` is off. Setting this to `true` silences that error and lets the rule run anyway.

## Original Documentation

[`@typescript-eslint/prefer-nullish-coalescing`](https://typescript-eslint.io/rules/prefer-nullish-coalescing)
