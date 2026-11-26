# no-restricted-syntax

## Rule Details

Disallows specified syntax. This is useful when a project wants to restrict language features such as `try-catch`, classes, or the `in` operator.

Restrictions use [ESTree](https://github.com/estree/estree) node names, such as `FunctionDeclaration` and `WithStatement`, or more precise [AST selectors](https://eslint.org/docs/latest/extend/selectors). Any node matching a configured selector is reported.

Examples of **incorrect** code for this rule with `["error", "FunctionExpression", "WithStatement", "BinaryExpression[operator='in']"]`, parsed as a script:

```javascript
with (me) {
  dontMess();
}

const doSomething = function () {};

foo in bar;
```

Examples of **correct** code for the same configuration:

```javascript
me.dontMess();

function doSomething() {}

foo instanceof bar;
```

## Options

The rule accepts a list of selector strings:

```json
{
  "no-restricted-syntax": [
    "error",
    "FunctionExpression",
    "WithStatement",
    "BinaryExpression[operator='in']"
  ]
}
```

Entries can also be objects with a required `selector` and an optional `message`:

```json
{
  "no-restricted-syntax": [
    "error",
    {
      "selector": "FunctionExpression",
      "message": "Function expressions are not allowed."
    },
    {
      "selector": "CallExpression[callee.name='setTimeout'][arguments.length!=2]",
      "message": "setTimeout must always be invoked with two arguments."
    }
  ]
}
```

The custom message replaces the default `Using '<selector>' is not allowed.` diagnostic. String and object entries can be mixed in one configuration.

Selectors can also target TypeScript syntax using standard [TSESTree](https://typescript-eslint.io/packages/ast-spec/) node names and attributes. For example, `TSEnumDeclaration[const=true]` selects const enums, `TSUnknownKeyword` selects `unknown` type keywords, and `TSAsExpression > TSAsExpression[typeAnnotation.type="TSUnknownKeyword"]` selects double casts such as `value as unknown as Foo`. `[const]` checks whether the attribute exists, so it matches ordinary enums as well.

## When Not To Use It

Leave this rule disabled if your project has no syntax restrictions.

## Differences from upstream

- Malformed selectors are ignored instead of rejecting the rule configuration. For example, `["error", "[", "WithStatement"]` still restricts `WithStatement` in rslint; ESLint rejects the malformed `[` selector.
- Files containing syntax errors or unsupported TypeScript constructs that cannot be projected to ESTree (such as Closure/JSDoc nullable type syntax like `?number` authored in TypeScript code) are skipped without reporting diagnostics for that file.

## Original Documentation

- [ESLint: no-restricted-syntax](https://eslint.org/docs/latest/rules/no-restricted-syntax)
- [Source code](https://github.com/eslint/eslint/blob/v10.14.0/lib/rules/no-restricted-syntax.js)
