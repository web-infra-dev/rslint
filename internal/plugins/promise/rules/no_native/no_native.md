# no-native

Require creating or importing a `Promise` constructor before using it.

This rule is useful for ES5 projects that intentionally use Bluebird or another
custom Promise implementation instead of relying on the runtime's native
`Promise` global.

## Invalid

```js
Promise.resolve('bad')

new Promise(function (resolve) {
  resolve('bad')
})
```

## Valid

```js
const Promise = require('bluebird')
const x = Promise.resolve('good')
```

```js
import Promise from 'bluebird'
const x = Promise.reject(new Error('good'))
```

## Differences from ESLint

`env` and `globals` options are not modeled by rslint's native Go rule tester.
The rule still treats the TypeScript default-library `Promise` symbol as native
and reports it, matching eslint-plugin-promise's behavior for ES6/native/global
Promise.

## References

- [eslint-plugin-promise: no-native](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-native.md)
