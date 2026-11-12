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

When `languageOptions.parserOptions.isolatedDeclarations` is `true`, the default
`"constructor"` mode preserves type annotations and does not report them. This
explicit parser option is independent of the compiler option in `tsconfig.json`.
The `"type-annotation"` mode continues to work when this option is enabled.

This rule does not provide suggestions.

## Differences from upstream

For a parenthesized constructor, the autofix preserves the original construction
semantics. With `"constructor"` mode, Rslint fixes
`const x: Foo<string> = new (Foo)();` to `const x = new (Foo)<string>();`.
Upstream v8.46.3 instead produces `const x = new (Foo<string>())();`, which calls
`Foo` first and then constructs its return value. Rslint inserts type arguments
after the complete parenthesized constructor expression.

## Original Documentation

- [typescript-eslint: consistent-generic-constructors](https://typescript-eslint.io/rules/consistent-generic-constructors)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.46.3/packages/eslint-plugin/src/rules/consistent-generic-constructors.ts)
