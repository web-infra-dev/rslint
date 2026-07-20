# prefer-array-flat

## Rule Details

Prefer `Array#flat()` over legacy techniques that flatten arrays with identity
`flatMap` callbacks, `reduce`, `concat`, Lodash, Underscore, or configured
helper functions.

Examples of **incorrect** code for this rule:

```javascript
const first = array.flatMap(element => element);
const second = array.reduce((result, element) => result.concat(element), []);
const third = [].concat(...array);
const fourth = _.flatten(array);
```

Examples of **correct** code for this rule:

```javascript
const first = array.flat();
const second = maybeArray.flat();
```

The `functions` option adds custom flattening functions:

```json
{
  "unicorn/prefer-array-flat": [
    "error",
    { "functions": ["flatArray", "utils.flat"] }
  ]
}
```

```javascript
const first = flatArray(array);
const second = utils.flat(array);
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-array-flat](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/prefer-array-flat.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/prefer-array-flat.js)
