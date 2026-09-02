# prefer-array-flat

## Rule Details

Prefer `Array#flat()` over legacy techniques that flatten arrays with identity
`flatMap` callbacks, `reduce`, `concat`, Lodash, Underscore, or configured
helper functions.

Examples of **incorrect** code for this rule:

```javascript
const first = array.flatMap(element => element);
const second = array.reduce((result, element) => result.concat(element), []);
const third = array.reduce((result, element) => [...result, ...element], []);
const fourth = [].concat(...array);
const fifth = [].concat.apply([], array);
const sixth = Array.prototype.concat.call([], ...array);
const seventh = _.flatten(array);
const eighth = lodash.flatten(array);
const ninth = underscore.flatten(array);
```

Examples of **correct** code for this rule:

```javascript
const first = array.flat();
const second = [].concat(maybeArray);
const third = Array.prototype.concat.call([], maybeArray);
```

Plain concat normalization does not consume an array of concat arguments, so
this rule intentionally leaves it to `prefer-spread`.

A matching `reduce()` is also ignored when its receiver is known not to be an
array, including a typed array, because the proposed `.flat()` replacement
does not exist on that receiver. Unknown receivers are still reported.

## Options

### functions

Type: `string[]`\
Default: `[]`

Adds custom flattening functions. `_.flatten()`, `lodash.flatten()`, and
`underscore.flatten()` are always checked.

```json
{
  "unicorn/prefer-array-flat": [
    "error",
    { "functions": ["flatArray", "utils.flat"] }
  ]
}
```

Examples of **incorrect** code for this rule with this option:

```javascript
const first = flatArray(array);
const second = utils.flat(array);
```

## Related Rules

- [prefer-array-flat-map](./prefer-array-flat-map.md)

## Original Documentation

- [eslint-plugin-unicorn: prefer-array-flat](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-array-flat.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-array-flat.js)
