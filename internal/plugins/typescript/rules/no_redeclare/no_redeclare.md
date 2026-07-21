# no-redeclare

## Rule Details

This rule disallows redeclaring variables. It extends ESLint's base `no-redeclare` to understand TypeScript constructs such as type aliases, interfaces, namespaces, and declaration merging.

Examples of **incorrect** code for this rule:

```javascript
var a = 3;
var a = 10;
```

```javascript
function a() {}
function a() {}
```

```typescript
type T = 1;
type T = 2;
```

Examples of **correct** code for this rule:

```javascript
var a = 3;
var b = function () {
  var a = 10;
};
```

```typescript
interface A {
  prop1: 1;
}
interface A {
  prop2: 2;
}
```

```typescript
class Foo {}
namespace Foo {}
```

## Options

### `builtinGlobals` (default: `true`)

When `true`, the rule reports redeclaring ECMAScript built-in globals and names
provided by TypeScript's active lib type definitions, such as `Object`,
`Promise`, or `HTMLElement`. Configured `languageOptions.globals` also
participate as built-ins. Active `/* global */` directives participate as
declarations in either mode; a final `:off` setting removes that inline global.
Turning off a value global does not remove a same-named TypeScript type global.

```json
{ "@typescript-eslint/no-redeclare": ["error", { "builtinGlobals": true }] }
```

```javascript
var Object = 0;
```

### `ignoreDeclarationMerge` (default: `true`)

When `true`, the rule ignores redeclarations that are legal TypeScript declaration merges:

- `interface` + `interface`
- `namespace` + `namespace`
- `class` + `interface` / `class` + `namespace` / `class` + `interface` + `namespace` (at most one class)
- `function` + `namespace` (at most one function)
- `enum` + `namespace` (at most one enum)

```json
{ "@typescript-eslint/no-redeclare": ["error", { "ignoreDeclarationMerge": true }] }
```

```typescript
function A() {}
namespace A {}
```

## Original Documentation

- [https://typescript-eslint.io/rules/no-redeclare](https://typescript-eslint.io/rules/no-redeclare)
- [https://eslint.org/docs/latest/rules/no-redeclare](https://eslint.org/docs/latest/rules/no-redeclare)
