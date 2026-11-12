# no-unused-vars

## Rule Details

Disallow variables that are declared or assigned but never read. Removing unused
bindings keeps scope intent clear and catches misspellings, incomplete refactors,
and discarded assignments.

Examples of **incorrect** code for this rule:

```javascript
const unused = 1;

function greet(name, punctuation) {
  return `Hello, ${name}`;
}

let result;
result = calculate();
```

Examples of **correct** code for this rule:

```javascript
const value = calculate();
consume(value);

function greet(name) {
  return `Hello, ${name}`;
}

export const publicValue = 1;
```

The rule reports a discarded binding at its last write in the binding's
variable scope. When it is safe to do so, the diagnostic includes a suggestion
that removes the unused declaration, parameter, destructuring element, class,
function, or import.

## JSX factory references

Classic JSX implicitly reads its factory even when the source contains no
explicit reference to it. With `jsx: "react"` in the selected TypeScript or
JavaScript project, this example uses the `React` import:

```jsx
import React from 'react';

export const Example = () => <div />;
```

The rule resolves JSX configuration in this order, using only explicitly
provided values:

1. File pragmas: `@jsxRuntime`, `@jsx`, and `@jsxFrag`.
2. Rslint settings: `settings.react.runtime`, `settings.react.pragma`, and
   `settings.react.fragment`.
3. Project compiler options: `jsx`, `jsxFactory`, and `jsxFragmentFactory`.
4. Default factories: `React.createElement` and `React.Fragment` (or the
   project's legacy `reactNamespace`).

Runtime is resolved first. An explicit `settings.react.runtime: "classic"`
overrides a project's `react-jsx` / `react-jsxdev` mode; `"automatic"` overrides
`jsx: "react"`. A file's `@jsxRuntime` overrides both. A file `@jsxImportSource`
also selects automatic JSX, unless that file explicitly selects classic JSX;
project `jsxImportSource` supplies automatic JSX when no file or settings runtime
has been selected.

With an automatic runtime, classic factory bindings are not marked as used,
even when factory settings or pragmas remain configured. With no explicit
runtime, a factory or fragment declaration can select classic JSX, including
`settings.react.pragma` or `settings.react.fragment` alone. Preserve-mode files
without any runtime or factory declaration retain ordinary unused-variable
checks.

For classic JSX, resolve factory and fragment independently using the order
above. An omitted, empty, or non-string setting does not replace the project's
factory. Each implicit reference resolves in the JSX expression's lexical
scope, including parameter defaults and shadowed bindings. A fragment factory
is read only when a shorthand fragment occurs.

File pragmas follow TypeScript's leading block-comment syntax. Native
`react/jsx-uses-react` is a no-op; enabling it alone does not select the runtime.

## Options

The rule accepts either `"all"` or `"local"` as a shorthand for `vars`, or one
options object:

```json
{
  "no-unused-vars": [
    "error",
    {
      "vars": "all",
      "args": "after-used",
      "caughtErrors": "all",
      "ignoreRestSiblings": false,
      "ignoreClassWithStaticInitBlock": false,
      "ignoreUsingDeclarations": false,
      "reportUsedIgnorePattern": false
    }
  ]
}
```

- `vars`: Check all variables (`"all"`, the default) or only variables in
  non-global scopes (`"local"`). Top-level ES module bindings are local and
  remain checked.
- `varsIgnorePattern`: Ignore variable names matching this JavaScript regular
  expression.
- `args`: Check all parameters (`"all"`), only parameters after the last used
  parameter (`"after-used"`, the default), or no parameters (`"none"`).
- `argsIgnorePattern`: Ignore parameter names matching this JavaScript regular
  expression.
- `caughtErrors`: Check catch-clause bindings (`"all"`, the default) or ignore
  them (`"none"`).
- `caughtErrorsIgnorePattern`: Ignore catch-clause binding names matching this
  JavaScript regular expression.
- `destructuredArrayIgnorePattern`: Ignore direct array-destructuring elements
  that match this JavaScript regular expression. Defaulted and rest elements
  continue to use their ordinary variable, parameter, or catch-clause option.
- `ignoreRestSiblings`: Ignore direct object-destructuring properties that have
  a rest sibling. Bindings nested inside those properties are still checked.
- `ignoreClassWithStaticInitBlock`: Ignore classes containing a static
  initialization block.
- `ignoreUsingDeclarations`: Ignore `using` and `await using` declarations.
- `reportUsedIgnorePattern`: Report a binding when its name matches an ignore
  pattern but the binding is actually used.

For example, this configuration allows underscore-prefixed parameters:

```json
{
  "no-unused-vars": ["error", { "args": "all", "argsIgnorePattern": "^_" }]
}
```

## The `/* exported */` comment

A script shares its globals with the other scripts loaded alongside it, where
this rule cannot see them being read. An `/* exported name */` block comment
declares that such a global is consumed elsewhere, and the rule counts the
comment itself as a use:

```javascript
/* exported publicValue */
var publicValue = 1;
```

The comment resolves each name only against the outer global scope. Whether a
file has that scope is determined by its effective
`languageOptions.sourceType`, not by the presence of `import` or `export`
syntax. With flat config, omitting `sourceType` makes `.js` and `.ts` files
modules even when they contain no module syntax, so the comment has no effect.
Set `sourceType: "script"` for a shared script.

Module bindings, bindings in a JavaScript CommonJS wrapper, block bindings, and
function locals are still reported. An exact, case-sensitive `.cjs` extension
defaults to the CommonJS wrapper. TypeScript-flavoured files configured as
`commonjs` retain a global program scope, so their top-level bindings can be
marked by the comment.

## Original Documentation

- [ESLint: no-unused-vars](https://eslint.org/docs/latest/rules/no-unused-vars)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/no-unused-vars.js)
