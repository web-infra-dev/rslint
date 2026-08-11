# no-iterator

## Rule Details

Disallow the use of the `__iterator__` property.

The `__iterator__` property was a SpiderMonkey extension to JavaScript that could be used to create custom iterators compatible with `for...in` and `for each...in` loops. However, this property is now obsolete, so it should not be used. The standard `Symbol.iterator` property should be used instead to define the iteration protocol.

Examples of **incorrect** code for this rule:

```javascript
Foo.prototype.__iterator__ = function () {};

var a = test.__iterator__;

var a = test['__iterator__'];
```

Examples of **correct** code for this rule:

```javascript
var __iterator__ = null;

var a = test[__iterator__];

Foo.prototype[Symbol.iterator] = function () {};
```

## Original Documentation

- [ESLint: no-iterator](https://eslint.org/docs/latest/rules/no-iterator)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-iterator.js)

https://eslint.org/docs/latest/rules/no-iterator
