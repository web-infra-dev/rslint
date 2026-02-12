# no-var-requires

## Rule Details

Disallow `require` statements except in import statements. In other words, the use of forms such as `var foo = require("foo")` is banned. Instead, use ES6-style imports or TypeScript's `import foo = require("foo")` syntax.

Standalone `require()` calls (as expression statements) and TypeScript `import ... = require(...)` declarations are allowed.

Examples of **incorrect** code for this rule:

```typescript
var foo = require('foo');
const foo = require('foo');
let foo = require('foo');
```

Examples of **correct** code for this rule:

```typescript
import foo from 'foo';
import foo = require('foo');
require('foo');

import { foo } from 'foo';
```

## Original Documentation

- [typescript-eslint no-var-requires](https://typescript-eslint.io/rules/no-var-requires)
