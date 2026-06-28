# promise/no-callback-in-promise

Disallow calling `cb()` inside of a `then()` (use `util.callbackify` instead).

Mixing promise callbacks (`.then` / `.catch`) with Node.js-style error-first callbacks
(`cb`, `callback`, `next`, `done`) makes control flow harder to reason about and is a
common source of swallowed errors. This rule flags places where a callback is called or
passed inside a promise handler.

## Rule Details

Examples of **incorrect** code:

```js
a.then(cb)
a.then(() => cb())
a.then(function(err) { cb(err) })
a.catch(function(err) { callback(err) })
```

Examples of **correct** code:

```js
// callback outside a promise — fine
function thing(cb) { cb() }

// wrapped in a timeout (default: timeoutsErr is false)
whatever.then((err) => { process.nextTick(() => cb()) })
```

## Options

```json
{
  "promise/no-callback-in-promise": ["warn", {
    "exceptions": [],
    "timeoutsErr": false
  }]
}
```

### `exceptions`

An array of callback names to exclude from the check.

```js
/* eslint promise/no-callback-in-promise: ["warn", { "exceptions": ["next"] }] */
a.then(() => next())  // OK — "next" is excluded
```

### `timeoutsErr`

When `true`, passing or calling a callback inside `setTimeout`, `setImmediate`,
`requestAnimationFrame`, or `process.nextTick` that is itself inside a promise handler
is also an error. Defaults to `false`.

```js
/* eslint promise/no-callback-in-promise: ["warn", { "timeoutsErr": true }] */
whatever.then(() => { process.nextTick(() => cb()) })  // error
whatever.then(() => process.nextTick(cb))              // error
```

## Differences from ESLint

None. The rule behaves identically to the upstream `eslint-plugin-promise` implementation.
