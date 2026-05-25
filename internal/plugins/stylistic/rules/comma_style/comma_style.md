# comma-style

Enforce consistent comma style.

## Rule Details

This rule enforces a consistent comma placement across comma-separated lists in JavaScript and TypeScript — array literals, object literals, variable declarations, function parameters, call arguments, import / export specifiers, sequence expressions, and the TypeScript-only lists (type / interface members, enum members, tuples, type parameters and type arguments, function / constructor signatures, import attributes).

The rule does NOT flag commas in single-line lists. A comma surrounded by linebreaks on both sides (a "lone" comma) reports under its own message regardless of the configured style.

## Options

This rule has a string option:

- `"last"` (default) — commas at the end of the current line.
- `"first"` — commas at the start of the next line.

This rule has an object option:

- `"exceptions"` — disables the rule for specific AST node types. Keys match the upstream ESTree / `@typescript-eslint/parser` node-type names so configuration is portable from ESLint Stylistic. Supported keys: `ArrayExpression`, `ArrayPattern`, `ObjectExpression`, `ObjectPattern`, `VariableDeclaration`, `FunctionDeclaration`, `FunctionExpression`, `ArrowFunctionExpression`, `CallExpression`, `NewExpression`, `ImportExpression`, `ImportDeclaration`, `ExportNamedDeclaration`, `ExportAllDeclaration`, `SequenceExpression`, `ClassDeclaration`, `ClassExpression`, `TSDeclareFunction`, `TSFunctionType`, `TSConstructorType`, `TSEmptyBodyFunctionExpression`, `TSMethodSignature`, `TSCallSignatureDeclaration`, `TSConstructSignatureDeclaration`, `TSEnumBody`, `TSTypeLiteral`, `TSInterfaceBody`, `TSInterfaceDeclaration`, `TSIndexSignature`, `TSTupleType`, `TSTypeParameterDeclaration`, `TSTypeParameterInstantiation`.

### last

Examples of **incorrect** code for this rule with the default `"last"` option:

```javascript
var foo = 1
,
bar = 2;

var foo = 1
  , bar = 2;

var foo = ["apples"
           , "oranges"];

function baz() {
  return {
    a: 1
    , b: 2
  };
}
```

Examples of **correct** code for this rule with the default `"last"` option:

```javascript
var foo = 1, bar = 2;

var foo = 1,
    bar = 2;

var foo = ["apples",
           "oranges"];

function baz() {
  return {
    a: 1,
    b: 2,
  };
}
```

### first

Examples of **incorrect** code for this rule with the `"first"` option:

```json
{ "@stylistic/comma-style": ["error", "first"] }
```

```javascript
var foo = 1,
    bar = 2;

var foo = ["apples",
           "oranges"];
```

Examples of **correct** code for this rule with the `"first"` option:

```json
{ "@stylistic/comma-style": ["error", "first"] }
```

```javascript
var foo = 1, bar = 2;

var foo = 1
    ,bar = 2;

var foo = ["apples"
          ,"oranges"];
```

### exceptions

Examples of **correct** code for this rule with `"first", { "exceptions": { "ArrayExpression": true, "ObjectExpression": true } }`:

```json
{ "@stylistic/comma-style": ["error", "first", { "exceptions": { "ArrayExpression": true, "ObjectExpression": true } }] }
```

```javascript
var ar = {fst:1,
          snd: [1,
                2]}
  , a = [];
```

## Original Documentation

- [@stylistic/comma-style](https://eslint.style/rules/comma-style)
