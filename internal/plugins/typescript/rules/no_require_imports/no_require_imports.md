# no-require-imports

## Rule Details

Disallow invocation of `require()`.

Prefer the newer ES6-style imports over `require()`. TypeScript projects should use `import` statements which provide better type safety and editor tooling support.

Examples of **incorrect** code for this rule:

```typescript
const fs = require('fs');
const path = require?.('path');
import foo = require('foo');
```

Examples of **correct** code for this rule:

```typescript
import fs from 'fs';
import * as path from 'path';
import { readFile } from 'fs';
```

## Original Documentation

- [typescript-eslint no-require-imports](https://typescript-eslint.io/rules/no-require-imports)
