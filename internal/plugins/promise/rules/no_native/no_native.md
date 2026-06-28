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

In TypeScript, a `Promise` reference is checked against its own namespace: value
references (including `typeof Promise`) require a local value declaration, and
type references require a local type declaration. The runtime's native `Promise`
is always reported.

## Original Documentation

- [eslint-plugin-promise: no-native](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-native.md)
