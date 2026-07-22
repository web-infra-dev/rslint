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

## Differences from ESLint

- With an `/* exported publicValue */` comment, rslint still reports an
  otherwise unused top-level `publicValue`; ESLint treats it as used.

```javascript
/* exported publicValue */
var publicValue = 1;
```

## Original Documentation

- [ESLint: no-unused-vars](https://eslint.org/docs/latest/rules/no-unused-vars)
