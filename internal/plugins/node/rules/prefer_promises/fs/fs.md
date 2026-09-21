# prefer-promises/fs

Prefer the promise API of Node.js's `fs` module.

## Rule details

This rule reports calls to callback APIs that have a promise counterpart. It
follows CommonJS imports, ES module imports, `process.getBuiltinModule`, and local
aliases, including the `node:fs` module name.

Examples of **incorrect** code for this rule:

```javascript
const fs = require('node:fs');
fs.readFile(filePath, 'utf8', (error, content) => {});
```

Examples of **correct** code for this rule:

```javascript
const { promises: fs } = require('node:fs');
const content = await fs.readFile(filePath, 'utf8');
```

```javascript
import fs from 'node:fs/promises';
const content = await fs.readFile(filePath, 'utf8');
```

Synchronous APIs, stream APIs, and reading a method without calling it are allowed.

## Options

This rule has no options and does not provide automatic fixes or suggestions.

## Original documentation

- [eslint-plugin-n: prefer-promises/fs](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/fs.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-promises/fs.js)
