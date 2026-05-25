# no-import-assign

## Rule Details

Disallows assigning to imported bindings. Imports are read-only references to values exported from other modules. Assigning to an import binding is always a mistake, as it will either throw a runtime error or silently fail.

For namespace imports (`import * as ns`), writing to any member of the namespace object is also disallowed, since namespace objects are frozen.

Examples of **incorrect** code for this rule:

```javascript
import mod from 'mod';
mod = 0;

import { named } from 'mod';
named = 0;
named++;

import * as ns from 'mod';
ns = 0;
ns.prop = 0;
ns.prop++;
```

Examples of **correct** code for this rule:

```javascript
import mod from 'mod';
mod.prop = 0; // Writing to a property of a default import is fine

import { named } from 'mod';
named.prop = 0; // Writing to a property of a named import is fine

import * as ns from 'mod';
ns.named.prop = 0; // Writing to nested properties is fine
```

## Original Documentation

- [ESLint no-import-assign](https://eslint.org/docs/latest/rules/no-import-assign)
