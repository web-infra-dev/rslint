# no-new-wrappers

## Rule Details

Disallows the use of `new` operators with `String`, `Number`, and `Boolean` as constructors. There are three primitive types in JavaScript that have wrapper objects: string, number, and boolean. These are represented by the constructors `String`, `Number`, and `Boolean`, respectively. Using these constructors to create new instances is generally considered bad practice because the primitive wrapper objects behave differently than their primitive counterparts in certain cases (e.g., `typeof new Boolean(false)` returns `"object"`).

Examples of **incorrect** code for this rule:

```javascript
var stringObject = new String('Hello world');
var numberObject = new Number(33);
var booleanObject = new Boolean(false);
```

Examples of **correct** code for this rule:

```javascript
var text = String(someValue);
var num = Number('33');
var bool = Boolean(someValue);
```

## Original Documentation

- [ESLint no-new-wrappers](https://eslint.org/docs/latest/rules/no-new-wrappers)
