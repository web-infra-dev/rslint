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

## Differences from upstream

For a generic predicate whose constraint contains a function signature, the
`explicitBooleanReturnType` suggestion annotates the predicate's own parameter
list. typescript-eslint v8.67.0 instead finds the first closing parenthesis and
can insert the annotation into the constraint, producing invalid syntax:

```typescript
// Input, with allowNumber: false
[true].some(<T extends () => boolean>() => 1);

// Rslint suggestion
[true].some(<T extends () => boolean>(): boolean => 1);
```

The suggestion also preserves an existing parameter list with a trailing comma,
including comments containing parentheses. For `(x, /* ) */) => 1`, Rslint
produces `(x, /* ) */): boolean => 1`. Upstream v8.67.0 misidentifies this
as a parameter without parentheses and produces an invalid annotation.

The string-length suggestion defaults an optional chain's short-circuit result
or a potentially nullish value hidden by a type assertion to an empty string
before accessing `.length`. For example, `user?.name!` becomes
`((user?.name!) ?? "").length > 0`; the original expression is evaluated once.
Upstream v8.67.0 can suggest `(user?.name!).length > 0`, which throws when
`user` is nullish. Negated conditions use the corresponding `=== 0` check. This protection is
consistent for all authored optional chains, including receivers whose static
types are non-nullish. For `value: string | undefined`, the suggestion for
`value!` is `(value! ?? "").length > 0`. The same protection applies when an
assertion wraps a conditional expression, or occurs inside `await`. For an
optional `Promise<string>`, the suggestion for `await value!` handles the awaited
nullish result before reading its length.

Suggestions that begin with a parenthesis insert a leading semicolon when needed
inside `case` and `default` clauses, namespace bodies and ordinary blocks. This prevents a
suggestion from joining a preceding expression statement through automatic
semicolon insertion, including when `as` or `satisfies` wraps the expression.
Upstream v8.67.0 misses these boundaries.

Array-predicate checks use the callback element of a spread tuple and exclude
its `thisArg`. Empty tuples allow selection of the next argument. A possibly
empty spread admits both its element type and the following argument as possible
callbacks. No inline callback edits are offered for a spread argument.
Upstream v8.67.0 can report `conditionErrorOther` on a valid tuple containing a
boolean callback and its `thisArg`. For array-like intersection types, an
explicit required `0` property retains the first callback and nonempty status
when the type uses the standard iterator. User-declared iterators use the
compiler's yield type independently of numeric metadata and do not imply a
nonempty argument list. `Set` and `Iterable` spreads also use the yield type. Opaque
collections and unresolved mixed collection unions remain conservative.

Computed array-method names are treated as unknown when their backing objects
may have been modified by a function call. Primitive strings copied from those objects
before a later-created callback remain stable; defining that callback does not
execute its body. Function-local captures of outer mutable objects remain
unknown when repeated calls could observe earlier mutations. Fresh local objects
and primitive literals retain the single-invocation ordering. Calls before the
copy still invalidate the backing object. This avoids classifying a runtime `forEach` call
as a predicate method based on an object's initial property value.

The `noStrictNullCheck` diagnostic uses an empty range at the start of the file
(line 1, column 1). Upstream uses line 0 for this file-level diagnostic; Rslint's
API reports 1-based source locations.

## Original Documentation

- [typescript-eslint: strict-boolean-expressions](https://typescript-eslint.io/rules/strict-boolean-expressions)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/strict-boolean-expressions.ts)
