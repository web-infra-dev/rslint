# no-useless-rename

Disallow renaming import, export, and destructured assignments to the same name.

## Rule Details

ES2015 allows for the renaming of references in import statements, export statements, and destructuring assignments. This gives programmers a concise syntax for performing these operations while renaming these references:

```javascript
import { foo as bar } from "baz";
export { foo as bar };
let { foo: bar } = baz;
```

With this syntax, it is possible to rename a reference to the same name. This is a completely redundant operation, as this is the same as not renaming at all.

Examples of **incorrect** code for this rule:

```javascript
import { foo as foo } from "bar";
export { foo as foo };
export { foo as foo } from "bar";
let { foo: foo } = bar;
let { 'foo': foo } = bar;
function foo({ bar: bar }) {}
({ foo: foo }) => {};
({ foo: foo } = bar);
```

Examples of **correct** code for this rule:

```javascript
import * as foo from "foo";
import { foo } from "bar";
import { foo as bar } from "baz";

export { foo };
export { foo as bar };
export { foo as bar } from "foo";

let { foo } = bar;
let { foo: bar } = baz;
let { [foo]: foo } = bar;

function foo({ bar }) {}
function foo({ bar: baz }) {}

({ foo }) => {};
({ foo: bar }) => {};
```

## Options

This rule has an object option:

- `"ignoreDestructuring": false` (default) — disallow useless renaming in destructuring patterns.
- `"ignoreImport": false` (default) — disallow useless renaming in import statements.
- `"ignoreExport": false` (default) — disallow useless renaming in export statements.

### ignoreDestructuring

Examples of **correct** code for this rule with `{ "ignoreDestructuring": true }`:

```json
{ "no-useless-rename": ["error", { "ignoreDestructuring": true }] }
```

```javascript
let { foo: foo } = bar;
function foo({ bar: bar }) {}
```

### ignoreImport

Examples of **correct** code for this rule with `{ "ignoreImport": true }`:

```json
{ "no-useless-rename": ["error", { "ignoreImport": true }] }
```

```javascript
import { foo as foo } from "bar";
```

### ignoreExport

Examples of **correct** code for this rule with `{ "ignoreExport": true }`:

```json
{ "no-useless-rename": ["error", { "ignoreExport": true }] }
```

```javascript
export { foo as foo };
export { foo as foo } from "bar";
```

## Original Documentation

- [ESLint rule](https://eslint.org/docs/latest/rules/no-useless-rename)
- [Source code](https://github.com/eslint/eslint/blob/main/lib/rules/no-useless-rename.js)
