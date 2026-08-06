# no-undef

Disallow the use of undeclared variables.

This rule reports identifiers that reference variables which have not been declared via `var`, `let`, `const`, `function`, `class`, `import`, or as a parameter.

Resolution is lexical scope analysis: declared bindings, standard ECMAScript built-ins (`Array`, `Promise`, `JSON`, `undefined`, and similar), and anything declared through `languageOptions.globals` or a `/* global */` comment. On a file with type information available, names known only to the type checker — DOM/Node globals such as `console` or `window`, or anything declared in a `.d.ts` file — are recognized too.

Enable this rule for files that aren't type-checked. On a type-checked file, `tsc` already reports references to undeclared names. On files without type information, declare the host globals your code relies on (`console`, `process`, `setTimeout`, and similar) through `languageOptions.globals` or a `/* global */` comment, the same way ESLint's own `no-undef` expects.

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
