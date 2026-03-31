# no-unused-vars

## Rule Details

Disallow unused variables.

Variables, functions, and function parameters that are declared but never used anywhere in the code are most likely an error due to incomplete refactoring. This rule extends the base ESLint `no-unused-vars` rule with TypeScript-specific awareness:

- Detects variables that are only used in type contexts (e.g., type annotations) and not in runtime code, reporting them as "defined but only used as a type"
- Recognizes type-level declarations (interfaces, type aliases, enums) and imports as validly used when referenced in type positions
- Handles declaration merging (e.g., interface + const with the same name)
- Respects ambient declarations (`declare module`, `.d.ts` files)
- Marks JSX factory and fragment factory imports as used when JSX elements are present (for `jsx: "preserve"` / `"react-native"` modes)

## Options

```jsonc
{
  "@typescript-eslint/no-unused-vars": ["error", {
    "vars": "all",                          // "all" | "local"
    "varsIgnorePattern": "",                 // regex pattern for vars to ignore
    "args": "after-used",                   // "after-used" | "all" | "none"
    "argsIgnorePattern": "",                // regex pattern for args to ignore
    "caughtErrors": "all",                  // "all" | "none"
    "caughtErrorsIgnorePattern": "",        // regex pattern for caught errors to ignore
    "destructuredArrayIgnorePattern": "",   // regex pattern for destructured array elements
    "ignoreRestSiblings": false,            // ignore siblings of rest properties
    "ignoreClassWithStaticInitBlock": false, // ignore classes with static init blocks
    "ignoreUsingDeclarations": false,       // ignore `using` / `await using` declarations
    "reportUsedIgnorePattern": false,       // report used vars that match ignore patterns
    "enableAutofixRemoval": {
      "imports": false                      // auto-fix to remove unused imports
    }
  }]
}
```

Examples of **incorrect** code for this rule:

```typescript
const unused = 42;

function foo(unusedParam: string) {
  return 'hello';
}

import { SomeValue } from './values';
const x: number = getSomething(); // SomeValue is never used
```

Examples of **correct** code for this rule:

```typescript
const used = 42;
console.log(used);

export function foo() {}

function bar(_unused: string, used: number) {
  return used;
}

// Type-only usage of imports is valid
import { SomeType } from './types';
const x: SomeType = getValue();

// Ignore pattern: variables starting with _ are ignored
function baz(_unused: string) {}
```

## Original Documentation

- [typescript-eslint no-unused-vars](https://typescript-eslint.io/rules/no-unused-vars)
