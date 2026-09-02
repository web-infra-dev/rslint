# no-implicit-globals

## Rule Details

It is the best practice to avoid 'polluting' the global scope with variables that are intended to be local to the script.

Global variables created from a script can produce name collisions with global variables created from another script, which will usually lead to runtime errors or unexpected behavior.

This rule disallows:

- Declarations that create one or more variables in the global scope.
- Global variable leaks.
- Redeclarations of read-only global variables and assignments to read-only global variables.

There is an explicit way to create a global variable when needed, by assigning to a property of the global object.

By default, this rule does not check `const`, `let` and `class` declarations.

### `var` and `function` declarations

This rule disallows `var` and `function` declarations at the top-level scope.

Examples of **incorrect** code for this rule:

```javascript
var foo = 1;

function bar() {}
```

Examples of **correct** code for this rule:

```javascript
// explicitly set on window
window.foo = 1;
window.bar = function () {};

// intended to be scope to this file
(function () {
  var foo = 1;

  function bar() {}
})();
```

### Global variable leaks

An assignment to an undeclared variable creates a new global variable, even inside a function. This will happen even if the code is in a function.

Examples of **incorrect** code for this rule:

```javascript
foo = 1;

Bar.prototype.baz = function () {
  a = 1; // Intended to be this.a = 1;
};
```

### Read-only global variables

This rule also disallows redeclarations of read-only global variables and assignments to read-only global variables.

A read-only global variable can be a built-in ES global (e.g. `Array`), or a global variable defined as `readonly` in the configuration file or in a `/*global */` comment.

Examples of **incorrect** code for this rule:

```javascript
/*global foo:readonly*/

foo = 1;

Array = [];
var Object;
```

### exported

You can use `/* exported variableName */` block comments to indicate that a variable is intentionally being made available for use in other scripts (for example, by loading them in the same page).

Examples of **correct** code for `/* exported variableName */`:

```javascript
/* exported global_var */

var global_var = 42;
```

## Options

This rule has an object option with one option:

- Set `"lexicalBindings"` to `true` if you want this rule to check `const`, `let` and `class` declarations as well.

### `const`, `let` and `class` declarations

Examples of **incorrect** code for this rule with `{ "lexicalBindings": true }`:

```json
{ "no-implicit-globals": ["error", { "lexicalBindings": true }] }
```

```javascript
const foo = 1;

let baz;

class Bar {}
```

Examples of **correct** code for this rule with `{ "lexicalBindings": true }`:

```json
{ "no-implicit-globals": ["error", { "lexicalBindings": true }] }
```

```javascript
{
  const foo = 1;
  let baz;
  class Bar {}
}

(function () {
  const foo = 1;
  let baz;
  class Bar {}
})();
```

## Differences from ESLint

- For a script parsed by ESLint with `languageOptions.parserOptions.ecmaFeatures.globalReturn: true`, ESLint treats the top level as a function scope and does not report its `var` or function declarations. Rslint treats the same source as a global script and reports those declarations.
- In TypeScript scripts, ESLint reports ambient `var` and function declarations and, with `lexicalBindings: true`, ambient `let`, `const`, and class declarations. Rslint does not report these declarations.
- For an overloaded global function in a TypeScript script, ESLint reports every overload signature and the implementation. Rslint reports only the implementation.
- In TypeScript assignment targets, rslint reports only runtime value targets. For `[foo as (x: T) => U] = value`, rslint reports `foo`; ESLint also reports the function-type parameter name `x`.
- Rslint recognizes value writes through nested erased assertions. For `(foo satisfies T) = value`, rslint reports `foo`, while ESLint does not.
- With `/* exported __proto__ */`, rslint suppresses the global-declaration diagnostic just as it does for other exported names. ESLint 10.9.1 still reports `__proto__` because of an upstream directive-parser bug.
- On invalid TypeScript accepted through parser recovery, such as `[foo<T>] = value` or `[foo + bar] = value`, ESLint may emit assignment diagnostics that rslint does not.

## Original Documentation

- [ESLint: no-implicit-globals](https://eslint.org/docs/latest/rules/no-implicit-globals)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/no-implicit-globals.js)
