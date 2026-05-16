# init-declarations

## Rule Details

Require or disallow initialization in variable declarations. This is the
TypeScript-aware version of the ESLint `init-declarations` rule: bindings
introduced by `declare` (either on the declaration itself or on an enclosing
`declare namespace` / `declare module 'm'` / `declare global`) are skipped.

Examples of **incorrect** code for this rule:

```typescript
var foo;
var bar: string;
let baz;
namespace myLib {
  let count: number;
}
```

Examples of **correct** code for this rule:

```typescript
var foo = null;
let bar: string = 'hi';
const baz = 0;
declare const x: number;
declare namespace myLib {
  let count: number;
}
```

## Options

### `mode`

**Type:** `"always" | "never"` — **Default:** `"always"`

When set to `"always"`, every variable declaration must include an initializer.
When set to `"never"`, declarators that include an initializer are reported.
`const`, `using`, and `await using` bindings are exempt from `"never"` because
they require an initializer at parse time.

Examples of **incorrect** code with `"never"`:

```json
{ "@typescript-eslint/init-declarations": ["error", "never"] }
```

```typescript
var foo = 1;
let bar: string = 'hi';
for (var i = 0; i < 1; i++) {}
```

Examples of **correct** code with `"never"`:

```json
{ "@typescript-eslint/init-declarations": ["error", "never"] }
```

```typescript
var foo;
let bar: string;
const baz = 1;
```

### `ignoreForLoopInit`

**Type:** `boolean` — **Default:** `false`

Only meaningful when `mode` is `"never"`. When `true`, declarators in the
initializer / left slot of `for`, `for-in`, and `for-of` statements are not
reported.

Examples of **correct** code with `["never", { "ignoreForLoopInit": true }]`:

```json
{ "@typescript-eslint/init-declarations": ["error", "never", { "ignoreForLoopInit": true }] }
```

```typescript
for (var i = 0; i < 1; i++) {}
for (var key in obj) {
}
for (var item of items) {
}
```

## Original Documentation

- [typescript-eslint init-declarations](https://typescript-eslint.io/rules/init-declarations)
- [ESLint init-declarations](https://eslint.org/docs/latest/rules/init-declarations)
