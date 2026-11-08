# prefer-module

## Rule details

Prefer JavaScript modules (ESM) over CommonJS.

This rule reports:

- `"use strict"` directives, because JavaScript modules are already strict.
- `return` statements outside functions.
- References to the CommonJS globals `require`, `exports`, `module`,
  `__dirname`, and `__filename`.

CommonJS names are checked when available as globals, for example through the
Node.js globals in `languageOptions.globals`. Locally declared and imported
bindings are ignored. Files ending in `.cjs` are ignored, regardless of case.

Examples of **incorrect** code:

```javascript
const {fromPairs} = require('lodash');
```

```javascript
module.exports = foo;
```

```javascript
const file = path.join(__dirname, 'foo.js');
```

Examples of **correct** code:

```javascript
import {fromPairs} from 'lodash-es';
```

```javascript
export default foo;
```

```javascript
const file = path.join(import.meta.dirname, 'foo.js');
```

## Fixes and suggestions

In `.mjs` files, automatic fixing removes `"use strict"` directives. Other files
receive a suggestion to remove the directive.

Suggestions convert top-level `require()` calls to imports and simple
`exports` or `module.exports` assignments to exports. JSON imports include a
`with {type: 'json'}` attribute. Calls inside functions, `let` declarations,
multiple variable declarations, and destructuring with defaults or rest
properties receive diagnostics without an import suggestion.

For `__dirname` and `__filename`, the rule suggests `import.meta.dirname` and
`import.meta.filename`. It also offers alternatives using
`path.dirname(url.fileURLToPath(import.meta.url))` and
`url.fileURLToPath(import.meta.url)`. Add the required `node:path` and `node:url`
imports when choosing those alternatives.

## Options

This rule has no options.

## Differences from upstream

Compared with eslint-plugin-unicorn v79.0.0, rslint keeps the diagnostic but
withholds suggestions that would produce invalid syntax or remove comments:

- Calls such as `require<string>('x')` and `require('x',)` need manual conversion
  to imports. The upstream suggestion retains the type arguments or trailing
  comma in the generated import.
- For `module /* comment */ .exports = value`, move the comment outside
  `module.exports` before applying the export suggestion. Otherwise, replacing
  the member expression would delete the comment.
- References in TypeScript types, such as `type File = typeof __filename`,
  receive no path replacement suggestions. Update the type manually when
  converting the file to ESM.
- Assignments and updates such as `__filename++` offer only the `import.meta`
  replacement. The alternative using `url.fileURLToPath()` cannot be assigned to.

## Original documentation

- [eslint-plugin-unicorn: prefer-module](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-module.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-module.js)
