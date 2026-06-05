# no-multiple-resolved

Disallow creating new promises where the executor can settle the promise more than once.

## Rule Details

A Promise executor that calls both `resolve` and `reject` (or calls either one twice) on the same
execution path settles the promise twice. The second settlement has no effect, but it indicates a
logic error and can mask real bugs.

Examples of **incorrect** code for this rule:

```javascript
new Promise((resolve, reject) => {
  if (error) {
    reject(error)
  }
  resolve(value) // may execute after reject
})

new Promise((resolve, reject) => {
  reject(error)
  resolve(value) // always executes after reject
})

new Promise(async (resolve, reject) => {
  try {
    const r = await foo()
    resolve()
    r() // can throw → catch may run after resolve
  } catch (error) {
    reject(error)
  }
})
```

Examples of **correct** code for this rule:

```javascript
new Promise((resolve, reject) => {
  if (error) {
    reject(error)
  } else {
    resolve(value)
  }
})

new Promise((resolve, reject) => {
  if (error) {
    reject(error)
    return
  }
  resolve(value)
})

new Promise(async (resolve, reject) => {
  try {
    const r = await foo()
    resolve(r) // last throwable expression — catch cannot run after this
  } catch (error) {
    reject(error)
  }
})
```

## Differences from ESLint

Correlated conditions across separate `if` statements (e.g. `if (err) { reject(err) }` followed
by `if (!err) { resolve(val) }`) are not recognized as mutually exclusive. This rule reports them
as potential double-resolution. Full ESLint code-path analysis handles these; our simplified
state-based analysis does not.

## Original Documentation

- [eslint-plugin-promise: no-multiple-resolved](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-multiple-resolved.md)
