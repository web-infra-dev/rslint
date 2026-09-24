# no-dynamic-require

## Rule Details

Disallow expressions as the first argument to `require()`, so tools can identify
module dependencies from the source code.

Examples of **incorrect** code for this rule:

```javascript
require(name);
require('../' + name);
require(`../${name}`);
require(name());
```

Examples of **correct** code for this rule:

```javascript
require('../name');
require(`../name`);
```

The rule checks direct calls named `require`, including locally declared
functions with that name. It ignores calls such as `require.resolve(name)` and
`module.require(name)`, calls without arguments, and arguments after the first.

Like upstream, the rule accepts any literal, including numbers, booleans, `null`,
bigints, and regular expressions. It also accepts template literals without
substitutions. It does not evaluate expressions, so `require('a' + 'b')` and
`` require(`${'name'}`) `` are reported.

The rule does not provide automatic fixes or suggestions.

## Options

### `esmodule`

Set `{ "esmodule": true }` to apply the same check to dynamic `import()` calls.
The default is `false`.

With this option enabled, the following code is **incorrect**:

```javascript
import(name);
import(`../${name}`);
```

The following code is **correct** with either setting:

```javascript
import('../name');
import(`../name`);
import name from '../name';
```

## Original Documentation

- [eslint-plugin-import: no-dynamic-require](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-dynamic-require.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-dynamic-require.js)
