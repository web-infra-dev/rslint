# no-anonymous-default-export

## Rule Details

Require a name for default exports of arrays, arrow functions, anonymous function
and class declarations, literals, objects, and class instances. Naming these
values makes their declarations and imports easier to find.

Examples of **incorrect** code for this rule:

```javascript
export default [];
export default () => {};
export default function () {}
export default class {}
export default 123;
export default `hello ${name}`;
export default {};
export default new Service();
```

Examples of **correct** code for this rule:

```javascript
const config = {};
export default config;

export default function initialize() {}
export default class Service {}

// Function calls are allowed by default.
export default createService();
```

Named exports and re-exports are allowed. Like upstream, the rule also allows
expressions such as `export default -1`, `export default (class {})`, and
`export default ({} as Config)`.

Dynamic imports such as `export default import('module')` and optional calls such
as `export default factory?.()` are allowed even when `allowCallExpression` is
`false`.

The rule does not provide automatic fixes or suggestions.

## Options

Each option allows its corresponding default export when set to `true`.

| Option                   | Default | Allowed export                                  |
| ------------------------ | ------- | ----------------------------------------------- |
| `allowArray`             | `false` | Array expressions                               |
| `allowArrowFunction`     | `false` | Arrow functions                                 |
| `allowAnonymousClass`    | `false` | Anonymous class declarations                    |
| `allowAnonymousFunction` | `false` | Anonymous function declarations                 |
| `allowCallExpression`    | `true`  | Function calls                                  |
| `allowNew`               | `false` | Class instances created with `new`              |
| `allowLiteral`           | `false` | Literals and templates, including substitutions |
| `allowObject`            | `false` | Object expressions                              |

`allowCallExpression` defaults to `true` for compatibility with upstream. With
`{ "allowCallExpression": false }`, assign the call result to a variable before
exporting it:

```javascript
// Incorrect
export default createService();

// Correct
const service = createService();
export default service;
```

## Original Documentation

- [eslint-plugin-import: no-anonymous-default-export](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-anonymous-default-export.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-anonymous-default-export.js)
