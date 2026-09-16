# no-mixed-requires

## Rule details

Disallow mixing `require()` declarations with other variables in the same `var`, `let`, or `const` declaration. Uninitialized variables count as other declarations. Accessing a property of a required module is allowed.

Examples of **incorrect** code:

```javascript
const fs = require("fs"), count = 0;
let path = require("path"), filename;
```

Examples of **correct** code:

```javascript
const fs = require("fs"), readFile = require("fs").readFile;
const count = 0;
let filename;
```

## Options

```javascript
"node/no-mixed-requires": ["error", { grouping: false, allowCall: false }]
```

- `grouping` (default: `false`): require declarations in one statement must all load the same kind of module: core, package, file, or computed. File paths start with `/`, `./`, or `../`. Arguments other than string literals, and calls without arguments, are computed.
- `allowCall` (default: `false`): allow calling the result of `require()` directly, such as `require("diagnostics")("app")`. Calling a property, such as `require("diagnostics").create("app")`, still counts as another declaration. When combined with `grouping`, the outer call's first argument determines the group.

The legacy boolean option is also supported: `true` enables grouping and `false` disables it. Prefer the object form.

With `grouping: true`, split different module kinds into separate declarations:

```javascript
// Incorrect
const fs = require("fs"), helper = require("./helper");

// Correct
const fs = require("fs"), path = require("path");
const helper = require("./helper");
```

With `allowCall: true`, a module factory can share a declaration with other requires:

```javascript
const helper = require("helper"), debug = require("diagnostics")("app");
```

Like upstream, locally defined functions named `require` are also checked. With `grouping: true`, `require("node:fs")` and `require("fs/promises")` belong to the package group, while `require("fs")` belongs to the core group.

This rule does not provide automatic fixes or suggestions.

## Original documentation

- [eslint-plugin-n: no-mixed-requires](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-mixed-requires.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-mixed-requires.js)
