# no-self-import

## Rule Details

Disallows a module from importing itself. A module that imports itself creates a circular dependency on itself, which is always a mistake and can cause confusing runtime behavior or errors. This applies to both ES module `import` statements and CommonJS `require()` calls.

Loading the current file through a known text or JSON attribute view is not a self module import because that request does not execute the file as JavaScript.

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

- [eslint-plugin-import: no-self-import](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-self-import.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-self-import.js)
