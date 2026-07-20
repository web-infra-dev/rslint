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
const second = [maybeArray].flat();
```

`[maybeArray].flat()` preserves the behavior of `[].concat(maybeArray)` when
the value may be either a single item or an array.

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

```javascript
const first = flatArray(array);
const second = utils.flat(array);
```

## Related Rules

- [prefer-array-flat-map](./prefer-array-flat-map.md)

## Original Documentation

- [eslint-plugin-unicorn: prefer-array-flat](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-array-flat.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/rules/prefer-array-flat.js)
