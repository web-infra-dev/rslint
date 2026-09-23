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
- `typedefs` (boolean, default `true`) - Whether to check type/interface declarations and type parameters
- `ignoreTypeReferences` (boolean, default `true`) - Whether to ignore references in type annotations
- `allowNamedExports` (boolean, default `false`) - Whether to allow references in named exports

Also accepts `"nofunc"` as a shorthand for `{ functions: false }`.

With `allowNamedExports: true`, named exports follow the other options. The
default `ignoreTypeReferences: true` ignores them, but setting it to `false`
checks them according to the referenced declaration. For example,
`export { value }; const value = 1;` is reported with
`{ allowNamedExports: true, ignoreTypeReferences: false }`.

## Original Documentation

- [typescript-eslint: no-use-before-define](https://typescript-eslint.io/rules/no-use-before-define)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.70.1/packages/eslint-plugin/src/rules/no-use-before-define.ts)
