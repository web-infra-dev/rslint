# no-delete-var

## Rule Details

Disallows the use of the `delete` operator on variables.

The purpose of the `delete` operator is to remove a property from an object. Using the `delete` operator on a variable might lead to unexpected behavior.

Examples of **incorrect** code for this rule:

```javascript
var x;
delete x;
```

Examples of **correct** code for this rule:

```javascript
var obj = { x: 1 };
delete obj.x;
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-delete-var
