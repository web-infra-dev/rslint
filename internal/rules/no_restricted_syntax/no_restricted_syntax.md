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

The implementation follows the esquery 1.7 selector forms used by ESLint
10.9.1, including the complete upstream `no-restricted-syntax` test suite:

- ESTree node names (e.g. `Identifier`, `FunctionDeclaration`,
  `BinaryExpression`) and supported TS-ESTree names such as
  `TSEnumDeclaration`. ESTree-only wrapper shapes such as `ClassBody`,
  `JSXEmptyExpression`, and `MethodDefinition.value` are exposed as virtual
  facades over the tsgo AST.
- Wildcard `*`.
- Field selectors, including nested fields (e.g. `Literal.key` and
  `.body.declarations.init`).
- Attribute selectors with presence (`[label]`), equality
  (`[name="x"]`, `[kind='using']`), inequality (`!=`), numeric comparisons
  (`[params.length>2]`), numeric path segments (`[arguments.0.type='Literal']`),
  `type(...)`, and regex matching (`[regex.flags=/i/]`). Attribute paths may
  inspect ESLint's `parent` link. BigInt metadata and class-method function
  fields are available through selectors such as `Literal[bigint]` and
  `MethodDefinition[value.body.body.length=0]`.
- Combinators `>` (direct child), descendant whitespace, `+`
  (adjacent sibling), `~` (general sibling), including decorator selectors
  such as `Decorator > CallExpression[callee.name='sealed']`.
- The `!` subject marker, including its reverse sibling/adjacent matching
  behavior, and ESLint's `:exit` event suffix.
- Pseudo-classes `:is()`, `:matches()`, `:not()`, `:has()`,
  `:first-child`, `:last-child`, `:nth-child(N)`, `:nth-last-child(N)`, and
  the semantic classes `:statement`, `:expression`, `:declaration`,
  `:function`, and `:pattern`.

## AST representation note

tsgo does not allocate separate nodes for every ESTree wrapper. Direct
selectors and structural relationships for those wrappers are modeled, with
ESLint-compatible ranges. A universe selector such as `*`, however, walks the
physical tsgo tree and therefore does not emit an additional diagnostic for
each virtual wrapper around the same physical node.

## Original Documentation

- [ESLint: no-restricted-syntax](https://eslint.org/docs/latest/rules/no-restricted-syntax)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/no-restricted-syntax.js)

[esquery]: https://github.com/estools/esquery
