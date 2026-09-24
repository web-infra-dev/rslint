# no-nodejs-modules

## Rule Details

Disallow Node.js builtin modules in code that runs without Node.js, such as a
browser application.

The rule checks static imports, re-exports, dynamic `import()` calls, and
single-argument `require()` calls with string literals. Explicit type imports
and re-exports are also checked. It reports the entire declaration or call and
does not provide automatic fixes or suggestions.

Examples of **incorrect** code for this rule:

```javascript
import fs from 'node:fs';
export { join } from 'path';
const buffer = require('buffer');
import('node:crypto');
```

Examples of **correct** code for this rule:

```javascript
import utility from 'utility';
import helper from './helper.js';
```

## Options

`allow` is an array of allowed module names. It defaults to `[]` and matches
the complete specifier exactly: allowing `fs` does not allow `node:fs` or
`fs/promises`.

```javascript
export default [
  {
    plugins: ['import'],
    rules: {
      'import/no-nodejs-modules': ['error', { allow: ['path', 'node:path'] }],
    },
  },
];
```

The rule respects `import/core-modules`, `import/internal-regex`, and
`import/resolver` settings. An internal regular-expression match takes
precedence over builtin classification. A package subpath counts as builtin
when its package root is builtin or configured as core, unless the subpath
resolves to a project file. `import/ignore` does not suppress this rule.

## Differences from upstream

`import/resolver` supports `node` and `typescript`. Other resolvers, such as
`webpack`, produce a resolver error. For projects with aliases, select
`typescript` and define the aliases in your `tsconfig.json`.

Select the TypeScript project with `languageOptions.parserOptions.project`.
For example, use `project: ['./tsconfig.app.json']` there;
`settings['import/resolver'].typescript.project` does not select a project.
If you use multiple resolvers and need a particular order, list them in an
array, such as `['typescript', 'node']`.

## When Not To Use It

Disable this rule for code that runs in Node.js and may use its builtin modules.

## Original Documentation

- [eslint-plugin-import: no-nodejs-modules](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-nodejs-modules.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-nodejs-modules.js)
