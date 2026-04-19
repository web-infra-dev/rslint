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

When `true`, the rule reports redeclaring names that shadow ECMAScript built-in globals such as `Object`, `Array`, or `Number`.

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

## Differences from ESLint

- `builtinGlobals` covers every global visible through the project's `tsconfig.json` `lib` setting — including DOM (`top`, `self`, `HTMLElement`, …) and ES-extension names (`Promise`, `WeakRef`, …). ESLint only flags globals the active `env` / `globals` options declare.
- A file counts as a module when it has a top-level `import` or `export`. Declarations in a module do not collide with lib globals, so `var Object = 0;` inside a module is not reported.
- `type Foo = ...` is not reported as shadowing a lib interface named `Foo` — type aliases and interfaces live in different declaration spaces, and the collision is not a real one.

## Original Documentation

- <https://typescript-eslint.io/rules/no-redeclare>
- <https://eslint.org/docs/latest/rules/no-redeclare>
