# no-global-assign

## Rule Details

Disallows assignments to native objects or read-only global variables. Built-in globals such as `Object`, `Array`, `String`, `Number`, `Math`, `JSON`, `undefined`, `NaN`, `Infinity`, and others should not be reassigned, as doing so can cause unexpected behavior throughout the application.

Examples of **incorrect** code for this rule:

```javascript
String = 'hello';
Array = 1;
undefined = true;
NaN++;
```

Examples of **correct** code for this rule:

```javascript
var x = String(123);
var y = new Array(1, 2, 3);

// Shadowed by local declaration
var String;
String = 'hello';

// Shadowed by function parameter
function foo(Array) {
  Array = 1;
}
```

Globals declared through [`languageOptions.globals`](/config/#languageoptionsglobals) or a `/* global */` comment carry their own access level: a `readonly` name is reported like a built-in, and a `writable` name may be reassigned — including a built-in whose declaration lifts the default. The environment maps exported as `globals` from `@rslint/core` use `false` for read-only names and `true` for writable names, so they feed the same access checks directly.

Examples of **incorrect** code with `globals: { BUILD_ID: 'readonly' }`:

```javascript
BUILD_ID = 'dev';
```

Examples of **correct** code with `globals: { Object: 'writable' }`:

```javascript
Object = {};
```

## Options

This rule accepts an optional object with an `exceptions` property, which is an array of global names that should be allowed to be reassigned:

```json
{
  "no-global-assign": ["error", { "exceptions": ["Object"] }]
}
```

## Original Documentation

- [ESLint: no-global-assign](https://eslint.org/docs/latest/rules/no-global-assign)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-global-assign.js)
