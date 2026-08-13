# no-undef

Disallow the use of undeclared variables.

This rule reports identifiers that reference variables which have not been declared via `var`, `let`, `const`, `function`, `class`, `import`, or as a parameter.

Resolution follows ESLint scope semantics: bindings declared or imported in the current file, the standard language globals selected by `languageOptions.ecmaVersion`, and names declared through `languageOptions.globals` or a `/* global */` comment. `ecmaVersion` defaults to `"latest"`.

TypeScript's TypeChecker does not alter the result. DOM, Node, cross-file, and ambient `.d.ts` names are not implicit ESLint globals, even when TypeScript can resolve them. Declare host globals such as `console`, `window`, `process`, and `setTimeout` through `languageOptions.globals` or a `/* global */` comment. TypeScript projects normally leave this core rule disabled because `tsc` already reports undeclared names.

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

## Original Documentation

- [ESLint: no-undef](https://eslint.org/docs/latest/rules/no-undef)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-undef.js)
