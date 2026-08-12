# no-restricted-syntax

## Rule Details

Disallows specified syntax. The rule accepts a list of [esquery] selectors;
any AST node matching one of the listed selectors triggers a diagnostic with
either a default or user-supplied message.

This is the catch-all rule for restricting language constructs (e.g. banning
`with`, banning `for-in`, requiring named function declarations) without
having to write a dedicated rule. The core selector grammar follows ESLint's
esquery-based implementation.

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
- Field selectors, including nested fields (e.g. `Literal.key` and
  `.body.declarations.init`).
- Attribute selectors with presence (`[label]`), equality
  (`[name="x"]`, `[kind='using']`), inequality (`!=`), numeric comparisons
  (`[params.length>2]`), numeric path segments (`[arguments.0.type='Literal']`),
  `type(...)`, and regex matching (`[regex.flags=/i/]`). Attribute paths may
  inspect ESLint's `parent` link.
- Combinators `>` (direct child), descendant whitespace, `+`
  (adjacent sibling), `~` (general sibling).
- Pseudo-classes `:is()`, `:matches()`, `:not()`, `:has()`,
  `:first-child`, `:last-child`, `:nth-child(N)`, `:nth-last-child(N)`, and
  the semantic classes `:statement`, `:expression`, `:declaration`,
  `:function`, and `:pattern`.

## Differences from ESLint

- ESLint rejects the configuration when a selector has invalid syntax. rslint
  ignores only that selector and continues with the remaining selectors, so
  its restriction is not enforced.
- Selectors that target TS-ESTree-only nodes or fields may report fewer
  diagnostics than ESLint. For example, `ClassBody > MethodDefinition`
  produces no diagnostics in rslint because tsgo has no separate `ClassBody`
  node.
- rslint ignores selectors using the `!` subject marker, such as
  `!IfStatement > BlockStatement`, while ESLint accepts them.

## Original Documentation

- [ESLint: no-restricted-syntax](https://eslint.org/docs/latest/rules/no-restricted-syntax)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-restricted-syntax.js)

[esquery]: https://github.com/estools/esquery
