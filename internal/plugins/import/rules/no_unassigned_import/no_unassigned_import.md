# no-unassigned-import

## Rule Details

Disallow imports whose result is not assigned or otherwise used. A bare `import`
or `require()` can load a module only for its side effects, making dependencies
harder to understand and test.

Examples of **incorrect** code:

```javascript
import 'should';
require('should');
import {} from 'module';
```

Examples of **correct** code:

```javascript
import value from 'module';
const other = require('other-module');
consume(require('module'));
require('module').initialize();
```

The rule checks standalone `require()` calls with exactly one string literal
argument. It does not check dynamic `import()`, dynamic `require()` arguments,
or calls whose result is used by another expression. It provides no automatic
fixes or suggestions because removing an import can remove side effects.

## Options

### `allow`

An array of glob patterns for imports that may be used only for side effects.
The default is `[]`, which allows no imports solely for their side effects.

```json
{
  "import/no-unassigned-import": [
    "error",
    { "allow": ["**/*.css", "polyfill", "src/setup/**"] }
  ]
}
```

Package names match as written. Import paths beginning with `.` or `/` match
their absolute paths, resolved from the importing file. Patterns also match
relative to the linter's working directory, usually the project root.

For example, in `src/app.js`, `import './styles/app.css'` is allowed by
`src/styles/**` or `**/*.css`, but not by `styles/*.css`.

The upstream schema also accepts `devDependencies`, `optionalDependencies`, and
`peerDependencies` as booleans or arrays. These options have no effect on this
rule, matching upstream behavior.

## Original Documentation

- [eslint-plugin-import: no-unassigned-import](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-unassigned-import.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unassigned-import.js)
