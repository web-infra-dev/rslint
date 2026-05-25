# no-use-before-define

## Rule Details

Disallow the use of variables before they are defined.

This rule extends the base ESLint `no-use-before-define` rule to add support for TypeScript-specific constructs like `type`, `interface`, and `enum` declarations.

Examples of **incorrect** code for this rule:

```typescript
alert(a);
var a = 10;

f();
function f() {}

new A();
class A {}

const foo = Foo.FOO;
enum Foo { FOO }
```

Examples of **correct** code for this rule:

```typescript
var a = 10;
alert(a);

type Foo = string;
const x: Foo = "hello";

function f() {}
f();
```

## Options

- `functions` (boolean, default `true`) - Whether to check function declarations
- `classes` (boolean, default `true`) - Whether to check class declarations
- `variables` (boolean, default `true`) - Whether to check variable declarations
- `enums` (boolean, default `true`) - Whether to check enum declarations
- `typedefs` (boolean, default `true`) - Whether to check type/interface declarations
- `ignoreTypeReferences` (boolean, default `true`) - Whether to ignore references in type annotations
- `allowNamedExports` (boolean, default `false`) - Whether to allow references in named exports

Also accepts `"nofunc"` as a shorthand for `{ functions: false }`.

## Original Documentation

- https://typescript-eslint.io/rules/no-use-before-define
- https://eslint.org/docs/latest/rules/no-use-before-define
