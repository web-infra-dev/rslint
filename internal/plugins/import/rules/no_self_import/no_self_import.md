# import/no-self-import

## Rule Details

Disallows a module from importing itself. A module that imports itself creates a circular dependency on itself, which is always a mistake and can cause confusing runtime behavior or errors. This applies to both ES module `import` statements and CommonJS `require()` calls.

Examples of **incorrect** code for this rule:

```javascript
// in file "foo.js"
import foo from './foo';

// in file "index.js"
const index = require('./index');
```

Examples of **correct** code for this rule:

```javascript
// in file "foo.js"
import bar from './bar';

// in file "index.js"
const utils = require('./utils');
```

## Original Documentation

- [import/no-self-import](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-self-import.md)
