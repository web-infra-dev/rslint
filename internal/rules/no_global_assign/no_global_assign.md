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

## Options

This rule accepts an optional object with an `exceptions` property, which is an array of global names that should be allowed to be reassigned:

```json
{
  "no-global-assign": ["error", { "exceptions": ["Object"] }]
}
```

## Original Documentation

- [ESLint no-global-assign](https://eslint.org/docs/latest/rules/no-global-assign)
