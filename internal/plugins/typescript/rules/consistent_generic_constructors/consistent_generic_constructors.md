# consistent-generic-constructors

## Rule Details

Enforce specifying generic type arguments on the type annotation or the constructor of a variable declaration. When constructing a generic class, the type arguments can be placed either on the left-hand side (type annotation) or the right-hand side (constructor call). This rule enforces consistency.

The rule supports two modes: `"constructor"` (default) prefers type arguments on the constructor call, and `"type-annotation"` prefers them on the type annotation.

Examples of **incorrect** code for this rule (with default `"constructor"` option):

```typescript
const map: Map<string, number> = new Map();
const set: Set<string> = new Set();
```

Examples of **correct** code for this rule (with default `"constructor"` option):

```typescript
const map = new Map<string, number>();
const set = new Set<string>();
const map2: Map<string, number> = new Map<string, number>();
```

## Autofix

This rule automatically moves generic type arguments to the preferred location.
It preserves whitespace within the type argument list and moves comments outside
that list when removing a type annotation. Constructor calls without parentheses
receive `()` when type arguments are moved onto the constructor.

An existing type annotation is preserved in `"type-annotation"` mode, including
an annotation without generic arguments such as `const a: Foo = new Foo<string>();`.

When the current file's TypeScript Program has `compilerOptions.isolatedDeclarations`
enabled, the default `"constructor"` mode preserves type annotations and does not
report them. The value comes from the file's resolved TypeScript project, including
its inherited compiler options. This check does not require a type checker.
The `"type-annotation"` mode continues to work when the compiler option is enabled.
When no project configuration is available, this protection is not enabled.

For optional and definite class properties, inserted annotations follow the
`?` or `!` marker, including computed keys and accessor properties.

This rule does not provide suggestions.

## Differences from upstream

For a parenthesized constructor, the autofix preserves the original construction
semantics. With `"constructor"` mode, Rslint fixes
`const x: Foo<string> = new (Foo)();` to `const x = new (Foo)<string>();`.
Upstream v8.46.3 instead produces `const x = new (Foo<string>())();`, which calls
`Foo` first and then constructs its return value. Rslint inserts type arguments
after the complete parenthesized constructor expression.

Rslint also preserves the exact line terminator when moving a line comment out
of a removed annotation. Upstream v8.46.3 drops the terminator, which can comment
out constructor arguments and their side effects. LF, CRLF, CR, LS and PS are
preserved.

In `"type-annotation"` mode, Rslint fixes `class C { x? = new Foo<string>(); }`
to `class C { x?: Foo<string> = new Foo(); }`. Upstream v8.46.3 instead inserts
before the optional marker and produces the invalid type `Foo<string>?`.
Rslint likewise keeps a definite assignment marker before an inserted annotation.

Unlike upstream v8.46.3, which reads an explicit parser option, Rslint uses the
current file's Program compiler options as the source of truth for
`isolatedDeclarations`. There is no separate public parser-option override.

## Original Documentation

- [typescript-eslint: consistent-generic-constructors](https://typescript-eslint.io/rules/consistent-generic-constructors)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.46.3/packages/eslint-plugin/src/rules/consistent-generic-constructors.ts)
