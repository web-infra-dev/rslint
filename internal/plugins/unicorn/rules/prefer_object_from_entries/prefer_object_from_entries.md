# prefer-object-from-entries

## Rule details

Prefer `Object.fromEntries()` when converting key-value pairs into an object.
The rule checks simple `reduce()` callbacks, an empty object followed by a
`for…of` loop that assigns each pair, and calls to `_.fromPairs()` or
`lodash.fromPairs()`.

Examples of **incorrect** code for this rule:

```javascript
const first = pairs.reduce(
  (object, [key, value]) => ({ ...object, [key]: value }),
  {},
);

const second = pairs.reduce(
  (object, [key, value]) => Object.assign(object, { [key]: value }),
  {},
);

const third = {};
for (const [key, value] of pairs) {
  third[key] = value;
}

const fourth = _.fromPairs(pairs);
```

Examples of **correct** code for this rule:

```javascript
const object = Object.fromEntries(pairs);
```

Automatic fixes preserve the mapped key and value. Reductions with an
accumulator used elsewhere in the callback, known non-array receivers, and
null-prototype accumulators such as `Object.create(null)` are ignored.
Non-computed `__proto__` properties are also ignored because converting them
would change how the object's prototype is set.

Comments that would be removed prevent automatic fixes. A reduction with
explicit TypeScript type arguments is reported without a fix.

## Options

### functions

Type: `string[]`\
Default: `[]`

Add names or dotted paths of functions that convert pairs into an object.
`_.fromPairs()` and `lodash.fromPairs()` are always checked.

```json
{
  "unicorn/prefer-object-from-entries": [
    "error",
    { "functions": ["getObjectFromKeyValue", "utils.fromPairs"] }
  ]
}
```

With this configuration, `utils.fromPairs(pairs)` is reported and fixed to
`Object.fromEntries(pairs)`.

## Original documentation

- [eslint-plugin-unicorn: prefer-object-from-entries](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-object-from-entries.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-object-from-entries.js)
