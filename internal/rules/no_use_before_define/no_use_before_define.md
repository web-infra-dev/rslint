# no-use-before-define

Disallow the use of variables before they are defined.

## Rule Details

In JavaScript, `var` declarations and function declarations are hoisted, so
using them before their declaration is legal but confusing. `let`, `const`, and
`class` declarations are not: reading them before the declaration throws at
runtime (the temporal dead zone). This rule reports a reference that appears
before the declaration it resolves to, and a reference that runs while that
declaration is still being initialized.

References to names that are not declared in the file — globals, ambient
declarations, imports from other modules — are not reported.

Examples of **incorrect** code for this rule:

```javascript
alert(a);
var a = 10;

f();
function f() {}

new C();
class C {}

var b = b;
```

Examples of **correct** code for this rule:

```javascript
var a = 10;
alert(a);

function f() {}
f();

class C {}
new C();

var b = 1;
```

Note that a reference from a different execution context still counts as
"before" when the declaration comes later in the file — the forward call in a
pair of mutually recursive functions is reported unless `functions` is turned
off:

```javascript
function isEven(n) {
  return isOdd(n - 1);
}
function isOdd(n) {
  return isEven(n - 1);
}
```

## Options

This rule takes one option, either the string `"nofunc"` or an object.

The string form `"nofunc"` is shorthand for `{ "functions": false }`.

The object form supports the following properties.

### `functions`

Whether references to function declarations are checked. Default: `true`.

Because function declarations are hoisted, calling one before its declaration is
safe, so turning this off is common in codebases that define helpers at the
bottom of a file.

Examples of **correct** code with `{ "functions": false }`:

```json
{ "no-use-before-define": ["error", { "functions": false }] }
```

```javascript
f();
function f() {}
```

### `classes`

Whether references to class declarations are checked. Default: `true`.

Turning it off only exempts references from a *different* execution context — a
function body, a method, a non-static field initializer. A reference in the same
context as the class definition is a temporal dead zone error and is still
reported.

Examples of **correct** code with `{ "classes": false }`:

```json
{ "no-use-before-define": ["error", { "classes": false }] }
```

```javascript
function make() {
  return new C();
}
class C {}
```

Examples of **incorrect** code with `{ "classes": false }`:

```json
{ "no-use-before-define": ["error", { "classes": false }] }
```

```javascript
new C();
class C {}
```

### `variables`

Whether references to `var`, `let`, and `const` declarations are checked.
Default: `true`.

As with `classes`, turning it off only exempts references from a different
execution context.

Examples of **correct** code with `{ "variables": false }`:

```json
{ "no-use-before-define": ["error", { "variables": false }] }
```

```javascript
function read() {
  return value;
}
let value = 1;
```

### `allowNamedExports`

Whether the local names in `export { ... }` are exempt. Default: `false`.

Examples of **correct** code with `{ "allowNamedExports": true }`:

```json
{ "no-use-before-define": ["error", { "allowNamedExports": true }] }
```

```javascript
export { a };
const a = 1;
```

### `enums`

Whether references to TypeScript `enum` declarations are checked. Default:
`true`.

Examples of **correct** code with `{ "enums": false }`:

```json
{ "no-use-before-define": ["error", { "enums": false }] }
```

```typescript
const value = Level.Low;

enum Level {
  Low,
}
```

### `typedefs`

Whether references to TypeScript `type` aliases, `interface` declarations, and
generic type parameters are checked. Default: `true`.

Examples of **correct** code with `{ "typedefs": false }`:

```json
{
  "no-use-before-define": [
    "error",
    { "typedefs": false, "ignoreTypeReferences": false }
  ]
}
```

```typescript
let value: Later;
type Later = string;
```

### `ignoreTypeReferences`

Whether references in type positions — a type annotation such as `let x: Foo`,
or a `typeof Foo` type query — are exempt. Default: `true`.

Because type positions are erased before the code runs, they cannot observe a
temporal dead zone, so they are ignored by default whatever `typedefs` and
`enums` say.

Examples of **incorrect** code with `{ "ignoreTypeReferences": false }`:

```json
{ "no-use-before-define": ["error", { "ignoreTypeReferences": false }] }
```

```typescript
interface Bar {
  type: typeof Foo;
}

const Foo = 2;
```

## Original Documentation

- [ESLint: no-use-before-define](https://eslint.org/docs/latest/rules/no-use-before-define)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-use-before-define.js)
