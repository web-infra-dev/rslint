# no-array-callback-reference

## Rule Details

Prevent passing function references directly to array iterator methods. An iterator passes extra arguments, such as the index and array, which can unexpectedly change a callback's behavior.

This rule checks `every`, `filter`, `find`, `findLast`, `findIndex`, `findLastIndex`, `flatMap`, `forEach`, `map`, `reduce`, `reduceRight`, and `some`. It offers editor suggestions to wrap the callback and choose the arguments to forward. There is no automatic fix.

Examples of **incorrect** code:

```javascript
array.map(callback);
array.forEach(callback);
array.reduce(reducer, initialValue);
```

Examples of **correct** code:

```javascript
array.map(element => callback(element));
array.forEach(element => {
  callback(element);
});
array.reduce((accumulator, element) => reducer(accumulator, element), initialValue);
```

Primitive conversions such as `array.map(Number)` and boolean predicates such as `array.filter(Boolean)` are allowed. TypeScript type predicates declared directly on a local callback and imported callbacks are allowed for `every`, `filter`, `find`, and `findLast`, so those methods can continue to narrow the result type.

Arrays, typed arrays, and unknown receivers are checked. Known non-array collections are ignored. Non-reducing iterator calls directly under `await` are also ignored.

## Options

### `ignore`

Type: `string[]`

Default: `[]`

Additional receiver names or dotted paths to ignore. The same names are recognized when the receiver is a call, such as `library().map(callback)`.

```javascript
{
  'unicorn/no-array-callback-reference': ['error', {
    ignore: ['Angular', 'library', 'myLib.utils'],
  }],
}
```

`Promise`, `React.Children`, `Children`, `lodash`, `underscore`, `_`, `Async`, `async`, `this`, `$`, and `jQuery` are always ignored. `Vue.filter()` and `types.map()` are also excluded.

## Differences from upstream

Suggested parameter names can differ for some irregular or capitalized array names. For example, for `People.map(callback)`, rslint suggests `(element) => callback(element)`, while upstream suggests `(Person) => callback(Person)`. The diagnostic and the arguments forwarded to the callback are the same.

## Original Documentation

- [eslint-plugin-unicorn: no-array-callback-reference](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/no-array-callback-reference.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/no-array-callback-reference.js)
