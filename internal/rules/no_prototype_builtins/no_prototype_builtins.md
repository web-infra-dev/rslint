# no-prototype-builtins

## Rule Details

This rule disallows calling `Object.prototype` methods directly on object
instances. In particular, it flags calls to `hasOwnProperty`, `isPrototypeOf`,
and `propertyIsEnumerable` invoked as members of a target object. Such calls
can break on objects created with `Object.create(null)` (which do not inherit
from `Object.prototype`) or on objects that define shadowing properties with
the same names.

Examples of **incorrect** code for this rule:

```javascript
var hasBarProperty = foo.hasOwnProperty('bar');
var isPrototypeOfBar = foo.isPrototypeOf(bar);
var barIsEnumerable = foo.propertyIsEnumerable('bar');
```

Examples of **correct** code for this rule:

```javascript
var hasBarProperty = Object.prototype.hasOwnProperty.call(foo, 'bar');
var isPrototypeOfBar = Object.prototype.isPrototypeOf.call(foo, bar);
var barIsEnumerable = {}.propertyIsEnumerable.call(foo, 'bar');
```

## Options

This rule has no options.

## Original Documentation

- ESLint rule: https://eslint.org/docs/latest/rules/no-prototype-builtins
