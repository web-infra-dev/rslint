# no-unused-vars

## Rule Details

Disallow unused variables.

Variables, functions, and function parameters that are declared but never used anywhere in the code are most likely an error due to incomplete refactoring. This rule extends the base ESLint `no-unused-vars` rule with TypeScript-specific awareness, such as detecting variables that are only used in type contexts (e.g., type annotations, type assertions) and not in runtime code.

Examples of **incorrect** code for this rule:

```typescript
const unused = 42;

function foo(unusedParam: string) {
  return 'hello';
}

import { SomeType } from './types';
const x: SomeType = getValue(); // x is only used as a type
```

Examples of **correct** code for this rule:

```typescript
const used = 42;
console.log(used);

export function foo() {}

function bar(_unused: string, used: number) {
  return used;
}
```

## Original Documentation

- [typescript-eslint no-unused-vars](https://typescript-eslint.io/rules/no-unused-vars)
