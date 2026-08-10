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

### Known differences from ESLint

The following differences may affect an existing ESLint
`no-restricted-syntax` configuration when it is used with rslint:

- ESLint rejects the configuration when a selector has invalid syntax. rslint
  currently skips that selector and continues with the remaining selectors, so
  the restriction expressed by the invalid selector is not enforced.
- Selectors for standard JavaScript nodes and commonly used TypeScript shapes
  are supported. A selector that targets a TS-ESTree-only node or field may not
  match yet. For example, `ClassBody > MethodDefinition` does not match because
  tsgo does not expose `ClassBody` as a separate AST node. Such selectors can
  therefore produce fewer diagnostics than the same ESLint configuration.
- ESLint accepts esquery's `!` subject marker in selector listener names, for
  example `!IfStatement > BlockStatement`. rslint does not currently support
  this marker and skips selectors that use it.

## Original Documentation

- [ESLint no-restricted-syntax](https://eslint.org/docs/latest/rules/no-restricted-syntax)

[esquery]: https://github.com/estools/esquery
