# no-shadow

Disallow variable declarations from shadowing variables declared in the outer scope.

## Rule Details

Shadowing occurs when a local variable shares the same name as a variable in its
containing scope. Inside the inner scope, the outer variable becomes
inaccessible, which can be a source of confusion.

Examples of **incorrect** code for this rule:

```javascript
var a = 3;
function b() {
  var a = 10;
}
```

Examples of **correct** code for this rule:

```javascript
var a = 3;
function b() {
  var c = 10;
}
```

## Options

### `builtinGlobals`

Shadowing a built-in global (for example `Object`, `Array`) is reported when
this option is `true`. Default: `false`.

```json
{ "no-shadow": ["error", { "builtinGlobals": true }] }
```

```javascript
function foo() {
  var Object = 0;
}
```

### `hoist`

Controls whether shadowing is reported before the outer declaration. Default:
`"functions"`.

- `"functions"`: report only before function declarations.
- `"all"`: always report, even when the outer declaration appears after the
  inner one.
- `"never"`: never report before the outer declaration.
- `"types"`: report when the outer declaration is a type (`type` or
  `interface`).
- `"functions-and-types"`: report for both outer function declarations and type
  declarations.

### `allow`

Array of names for which shadowing is allowed. Default: `[]`.

```json
{ "no-shadow": ["error", { "allow": ["done"] }] }
```

### `ignoreOnInitialization`

Ignores shadowing inside the initializer of the outer declaration when it is
called as a callback or IIFE. Default: `false`.

```json
{ "no-shadow": ["error", { "ignoreOnInitialization": true }] }
```

### `ignoreTypeValueShadow`

Ignores shadowing between a value and a type of the same name (TypeScript).
Default: `true`.

### `ignoreFunctionTypeParameterNameValueShadow`

Ignores shadowing for parameters declared inside a function type. Default:
`true`.

## Original Documentation

[https://eslint.org/docs/latest/rules/no-shadow](https://eslint.org/docs/latest/rules/no-shadow)
