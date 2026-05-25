# no-undef

Disallow the use of undeclared variables.

This rule reports identifiers that reference variables which have not been declared via `var`, `let`, `const`, `function`, `class`, `import`, or as a parameter.

## Options

### `typeof`

Type: `boolean`
Default: `false`

When set to `true`, `typeof` expressions will be checked for undeclared variables. By default, `typeof` of an undeclared variable does not trigger a warning, since `typeof` returns `"undefined"` for undeclared variables without throwing a ReferenceError.

## Examples

### Invalid

```js
a = 1; // 'a' is not defined.
var x = b; // 'b' is not defined.
undeclaredFunc(); // 'undeclaredFunc' is not defined.
```

With `{ "typeof": true }`:

```js
typeof x === 'string'; // 'x' is not defined.
```

### Valid

```js
var a = 1;
a;

function f() {}
f();

typeof maybeUndefined === 'string';
```
