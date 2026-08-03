# prefer-node-protocol

## Rule Details

Enforces the use of the `node:` protocol when importing Node.js builtin modules.
The `node:` protocol makes it explicit that a module is a Node.js builtin and
disambiguates it from a same-named package on disk.

The rule checks module specifiers in `import` and `export` statements, dynamic
`import()`, `require()`, `process.getBuiltinModule()`, and TypeScript `import(...)`
type nodes. It reports a builtin module referenced without the `node:` prefix and
autofixes it by inserting `node:` before the module name.

A specifier is only flagged when both its bare name and its `node:`-prefixed name
are Node.js builtin modules. Names that exist only under one form are left alone
(e.g. `test` is untouched because only `node:test` is a builtin, and `punycode`
is untouched because it is no longer a builtin under either form).

Examples of **incorrect** code for this rule:

```javascript
import fs from 'fs';
export { promises } from 'fs';
export * from 'fs';
const fs = require('fs/promises');
const fs = process.getBuiltinModule('fs');
```

```typescript
type fs = import('fs');
```

Examples of **correct** code for this rule:

```javascript
import fs from 'node:fs';
const fs = require('node:fs/promises');

// Not a Node.js builtin module.
import fs from 'unknown-builtin-module';

// Only `node:test` is a builtin, bare `test` is not.
import 'test';

// No longer a builtin under either form.
import 'punycode';
```

```typescript
type fs = import('node:fs');
```

## Original Documentation

[eslint-plugin-unicorn/prefer-node-protocol](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/docs/rules/prefer-node-protocol.md)
