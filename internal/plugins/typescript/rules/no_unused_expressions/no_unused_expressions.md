# no-unused-expressions

## Rule Details

Disallow unused expressions. An unused expression is an expression that is evaluated but whose result is not used. This can indicate a mistake or a misunderstanding of the code.

Examples of **incorrect** code for this rule:

```typescript
0;
a;
f(), {};
a && b();
foo.bar;
foo as any;
<any>foo;
foo!;
Foo<string>;
```

Examples of **correct** code for this rule:

```typescript
a = b;
a();
new Foo();
delete foo.bar;
void 0;
'use strict';
import('./foo');
foo?.();
```

## Options

- `allowShortCircuit` (default: `false`): Allow short-circuit evaluations (e.g., `a && a()`).
- `allowTernary` (default: `false`): Allow ternary expressions (e.g., `a ? b() : c()`).
- `allowTaggedTemplates` (default: `false`): Allow tagged template literals.
- `enforceForJSX` (default: `false`): Enforce the rule for JSX elements.
- `ignoreDirectives` (default: `false`): Ignore directive prologues.

## Original Documentation

- [typescript-eslint: no-unused-expressions](https://typescript-eslint.io/rules/no-unused-expressions)
- [ESLint: no-unused-expressions](https://eslint.org/docs/latest/rules/no-unused-expressions)
