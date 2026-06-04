# catch-or-return

Enforce the use of `catch()` on un-returned promises.

## Rule Details

A promise that is used as a statement — rather than returned from a function — must
be terminated with a rejection handler. Without one, rejected promises produce
unhandled-rejection errors that are easy to miss. This rule enforces that every
such promise chain ends with `.catch()` (or a configured alternative), so the
caller cannot silently swallow errors.

Examples of **incorrect** code for this rule:

```javascript
frank().then(go)

Promise.resolve(frank)

function callPromise(promise, cb) {
  promise.then(cb)
}
```

Examples of **correct** code for this rule:

```javascript
frank().then(go).catch(doIt)

function a() {
  return frank().then(go)
}

cy.get('.myClass').then(go)
```

## Options

### `allowThen`

Pass `{ "allowThen": true }` to accept a two-argument `.then(onFulfilled, onRejected)` call as
a valid termination in place of `.catch()`.

```json
{ "promise/catch-or-return": ["error", { "allowThen": true }] }
```

```javascript
frank().then(a, b)
frank().then(go).then(zam, doIt)
```

### `allowThenStrict`

Like `allowThen`, but the first argument to `.then()` must be `null` so the handler is
exclusively for rejections.

```json
{ "promise/catch-or-return": ["error", { "allowThenStrict": true }] }
```

```javascript
frank().then(go).then(null, doIt)
```

### `allowFinally`

Pass `{ "allowFinally": true }` to allow a `.finally()` call after a valid termination.

```json
{ "promise/catch-or-return": ["error", { "allowFinally": true }] }
```

```javascript
frank().then(go).catch(doIt).finally(fn)
```

### `terminationMethod`

Replace `catch` with an alternative method name, or provide an array of accepted names.

```json
{ "promise/catch-or-return": ["error", { "terminationMethod": "done" }] }
```

```javascript
frank().then(go).done()
```

```json
{ "promise/catch-or-return": ["error", { "terminationMethod": ["catch", "asCallback", "finally"] }] }
```

```javascript
frank().then(go).catch(doIt)
frank().then(go).asCallback(fn)
frank().then(go).finally(fn)
```

## Differences from ESLint

None known.

## Original Documentation

- [eslint-plugin-promise: catch-or-return](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/catch-or-return.md)
