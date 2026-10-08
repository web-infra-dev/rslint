# no-import-module-exports

Disallow static import declarations in files that also use CommonJS exports.
Mixing the two module systems can produce behavior that differs between Node.js,
bundlers, and transpilers.

Examples of **incorrect** code for this rule:

```javascript
import { value } from './value.js';
module.exports = value;

import value from './value.js';
exports.value = value;
```

Examples of **correct** code for this rule:

```javascript
import value from './value.js';
export default value;

const value = require('./value.js');
module.exports = value;
```

The package entry point selected by `package.json` is exempt. This supports
packages whose entry point imports internal ES modules and then exposes a
CommonJS export. The rule does not provide automatic fixes or suggestions.

## Options

### `exceptions`

`exceptions` is an array of minimatch glob patterns matched against the file's
absolute path. A matching file is not checked:

```json
{
  "import/no-import-module-exports": [
    "error",
    { "exceptions": ["**/*/some-file.js"] }
  ]
}
```

## Original Documentation

- [eslint-plugin-import: no-import-module-exports](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-import-module-exports.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-import-module-exports.js)
