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

## Differences from upstream

Rslint preserves syntax and expression boundaries where typescript-eslint v8.72.2 can produce an unsafe suggestion:

- **Logical expressions:** suggestions preserve the original grouping and add parentheses wherever `??` would otherwise mix with `||` or `&&`. Replacing the last `||` in `a ||= b || c || d` produces `a ||= (b || c) ?? d`. A ternary such as `a !== null ? a : b || 'z'` becomes `a ?? (b || 'z')`.
- **Computed keys:** `b.c` and `b[c]` are distinct expressions when used as keys. Rslint does not simplify `a[b.c] !== undefined ? a[b[c]] : 'fallback'`, because the accesses can return different values. Literal property access such as `a.x` and `a['x']` remains supported.
- **Assignment operators:** an `if` body using an arithmetic or bitwise assignment, or `&&=`, is not simplified to `??=`. For example, replacing `if (a === undefined) a += 'x'` would discard the addition. Simple `=`, `||=`, and `??=` assignments remain eligible.
- **Shadowed names:** a parameter or local binding named `undefined` is not treated as the global nullish value.
- **Comments:** moved line comments retain their line breaks. If a ternary or `if` rewrite would discard a comment in removed syntax, Rslint reports the diagnostic without a suggestion so the comment can be preserved manually.

Computed keys containing optional chains are compared conservatively; the rule does not treat `b?.c` and `b.c` as interchangeable key expressions. This is separate from matching optional chains along the member access being checked.

## Original Documentation

- [typescript-eslint: prefer-nullish-coalescing](https://typescript-eslint.io/rules/prefer-nullish-coalescing)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/src/rules/prefer-nullish-coalescing.ts)
