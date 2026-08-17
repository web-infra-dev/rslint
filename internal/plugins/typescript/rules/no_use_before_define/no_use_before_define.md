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

## Differences from ESLint

rslint also ships the core `no-use-before-define` rule, which gained TypeScript
support in ESLint 10 and takes the same options. Enable one or the other — the
two disagree in a few places, because this rule extends an older version of the
core rule:

- Code that reads a class binding while the class itself is still being defined
  — `class C extends C {}`, `class C { [C](){} }`, `const C = class { static x = C }` —
  is reported by the core rule and not by this one.
- A class field initializer or static block is an ordinary separate scope here,
  so `classes`, `variables`, and `enums` exempt references from inside one. The
  core rule treats static initializers as part of the surrounding code, and
  still reports them.
- `ignoreTypeReferences` covers every type position here — including
  `implements` clauses, qualified type names, and the exported name of
  `export = X` / `export default X`. The core rule only exempts a bare type
  annotation and a `typeof` query.
- A reference from the parameter list of a function type (`type F = (x: Foo) => void`)
  is never reported here.
- A reference that resolves to a string-literal enum member (`enum E { b = a, "a" = 1 }`)
  is not reported here, because such a member declares no identifier. The core
  rule measures it from the literal and reports it.

## Original Documentation

- [typescript-eslint: no-use-before-define](https://typescript-eslint.io/rules/no-use-before-define)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-use-before-define.ts)
