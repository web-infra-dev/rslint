# sort-imports

## Rule Details

This rule enforces a consistent order for import declarations and for named
members within an import declaration.

Examples of **incorrect** code for this rule:

```javascript
import b from "b";
import a from "a";
import { z, a } from "values";
```

Examples of **correct** code for this rule:

```javascript
import "setup";
import * as namespace from "namespace";
import { a, z } from "values";
import value from "value";
```

The declaration order can be customized with `memberSyntaxSortOrder`.
`ignoreCase`, `ignoreDeclarationSort`, `ignoreMemberSort`, and
`allowSeparatedGroups` provide the same controls as ESLint.

## Original Documentation

- [ESLint: sort-imports](https://eslint.org/docs/latest/rules/sort-imports)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.0/lib/rules/sort-imports.js)
