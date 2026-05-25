# no-restricted-syntax

## Rule Details

Disallows specified syntax. The rule accepts a list of [esquery] selectors;
any AST node matching one of the listed selectors triggers a diagnostic with
either a default or user-supplied message.

This is the catch-all rule for restricting language constructs (e.g. banning
`with`, banning `for-in`, requiring named function declarations) without
having to write a dedicated rule. The selector grammar matches ESLint's, so
selectors authored against ESLint's `no-restricted-syntax` configuration are
expected to work unchanged in rslint.

Examples of **incorrect** code for this rule:

```json
{
  "no-restricted-syntax": [
    "error",
    "FunctionExpression",
    "WithStatement"
  ]
}
```

```javascript
with (me) {
  dontMess();
}

const doSomething = function () {};
```

Examples of **correct** code for the same configuration:

```javascript
me.dontMess();

function doSomething() {}

foo instanceof bar;
```

## Options

The rule accepts an array of restriction entries. Each entry is either:

- A bare string — the esquery selector. The diagnostic message is
  `Using '<selector>' is not allowed.`.
- An object `{ "selector": <string>, "message"?: <string> }`. When `message`
  is provided it replaces the default text verbatim.

```json
{
  "no-restricted-syntax": [
    "error",
    {
      "selector": "CallExpression[callee.name='setTimeout']",
      "message": "Use the timer service instead of raw setTimeout."
    },
    "WithStatement"
  ]
}
```

### Supported selector forms

The implementation covers the subset of [esquery] used in real-world ESLint
configurations and in the upstream `no-restricted-syntax` test suite:

- ESTree node names (e.g. `Identifier`, `FunctionDeclaration`, `BinaryExpression`).
- Wildcard `*`.
- Class selectors `.field` (e.g. `Literal.key`).
- Attribute selectors with presence (`[label]`), equality
  (`[name="x"]`, `[kind='using']`), inequality (`!=`), numeric comparisons
  (`[params.length>2]`), and regex matching (`[regex.flags=/i/]`).
- Combinators `>` (direct child), descendant whitespace, `+`
  (adjacent sibling), `~` (general sibling).
- Pseudo-classes `:is()`, `:matches()`, `:not()`, `:has()`,
  `:first-child`, `:last-child`, `:nth-child(N)`, `:nth-last-child(N)`.

## Original Documentation

- [ESLint no-restricted-syntax](https://eslint.org/docs/latest/rules/no-restricted-syntax)

[esquery]: https://github.com/estools/esquery
