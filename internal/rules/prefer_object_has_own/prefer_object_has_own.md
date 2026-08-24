# prefer-object-has-own

## Rule Details

`Object.prototype.hasOwnProperty.call(object, property)` is the long-standing way to ask whether an object has a property of its own without going through the object's own `hasOwnProperty`, which may be missing or redefined. `Object.hasOwn()`, introduced in ES2022, asks the same question directly.

This rule reports a call to `hasOwnProperty` reached through the global `Object`, `Object.prototype`, or an empty object literal, and offers `Object.hasOwn()` in its place.

Examples of **incorrect** code for this rule:

```javascript
Object.prototype.hasOwnProperty.call(object, 'foo');

Object.hasOwnProperty.call(object, 'foo');

({}).hasOwnProperty.call(object, 'foo');

const hasProperty = Object.prototype.hasOwnProperty.call(object, property);
```

Examples of **correct** code for this rule:

```javascript
Object.hasOwn(object, 'foo');

const hasProperty = Object.hasOwn(object, property);

object.hasOwnProperty('foo');

foo.hasOwnProperty.call(object, 'foo');

function check(Object) {
  return Object.prototype.hasOwnProperty.call(object, 'foo');
}
```

## Options

This rule has no options.

## When Not To Use It

Use this rule only where ES2022 is available: `Object.hasOwn()` does not exist in older runtimes.

## Differences from ESLint

- A TypeScript declaration that binds the name `Object` as a type only — a `type` alias, an `interface`, or a type parameter — leaves the call reported; ESLint stays silent on the whole file. A value declaration (`const Object`, `class Object`, `import Object from …`, `declare const Object`) silences the report in both.

## Original Documentation

- [ESLint: prefer-object-has-own](https://eslint.org/docs/latest/rules/prefer-object-has-own)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/prefer-object-has-own.js)
