# no-array-constructor

## Rule Details

Disallow generic `Array` constructors.

Use of the `Array` constructor to create arrays is generally discouraged in favor of array literal notation because of the single-argument pitfall and because the `Array` global may be redefined.

The rule allows single-argument calls since they are commonly used to create arrays with a specific size.

Examples of **incorrect** code for this rule:

```javascript
new Array();
Array();
new Array(x, y);
Array(x, y);
new Array(0, 1, 2);
Array(0, 1, 2);
```

Examples of **correct** code for this rule:

```javascript
[];
[x, y];
[0, 1, 2];
new Array(500); // single argument creates array with size
Array(someOtherArray.length);
new Array<Foo>(); // TypeScript generic syntax
new Array<Foo>(1, 2, 3);
```

## Original Documentation

- [typescript-eslint no-array-constructor](https://typescript-eslint.io/rules/no-array-constructor)
- [ESLint no-array-constructor](https://eslint.org/docs/rules/no-array-constructor)
