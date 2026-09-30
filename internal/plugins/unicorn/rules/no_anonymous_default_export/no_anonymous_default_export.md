# no-anonymous-default-export

## Rule Details

Disallow anonymous functions and classes as the default export. Naming default
exports makes it easier to find a module's implementation and use a consistent
name when importing it.

Examples of **incorrect** code for this rule:

```javascript
export default function () {}
export default class {}
export default () => {};

module.exports = function () {};
module.exports = class {};
exports = () => {};
```

Examples of **correct** code for this rule:

```javascript
export default function foo() {}
export default class Foo {}

const foo = () => {};
export default foo;
module.exports = foo;
```

The rule checks default exports and standalone assignments to `module.exports`
or `exports`. It allows other exported values, named functions and named classes.
Computed access such as `module['exports']` is not checked.

### Suggestions

An editor suggestion derives a name from the filename before its first dot,
using camel case for functions and an initial capital for classes. Underscores
are appended to avoid reserved names and names already used in the relevant
scopes. A filename that cannot produce an identifier receives a diagnostic
without a suggestion.

For arrow functions, the suggestion introduces a `const` declaration and exports
that identifier. It preserves comments, parentheses and the file's line ending.
Suggestions are applied manually; this rule does not provide automatic fixes.

## Options

This rule has no options.

## Differences from upstream

Compared with eslint-plugin-unicorn v76.0.0:

- For anonymous TypeScript functions with type parameters, the suggestion inserts
  the name before the type parameters. For example,
  `export default function<T>(value: T) { return value; }` becomes
  `export default function foo<T>(value: T) { return value; }` in `foo.ts`.
  Upstream places the name after `<T>`, producing invalid TypeScript.
- An arrow function assigned to `module.exports` or `exports` directly inside
  an unbraced conditional, loop or labeled statement receives a diagnostic
  without a suggestion. For example, `if (ready) exports = () => {};` is reported,
  but rslint does not suggest inserting a `const` declaration in that position.
  Add braces around the statement to make the naming suggestion available.

## Original Documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/no-anonymous-default-export.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/no-anonymous-default-export.js)
