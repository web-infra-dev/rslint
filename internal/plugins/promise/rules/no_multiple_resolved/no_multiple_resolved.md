# promise/no-multiple-resolved

## Rule Details

Disallows creating new promises with paths that resolve multiple times.

A Promise executor that calls both `resolve` and `reject` (or calls either one twice) on the same execution path settles the promise twice. The second settlement has no effect, but it indicates a logic error and can mask real bugs.

Examples of **incorrect** code for this rule:

```javascript
new Promise((resolve, reject) => {
  if (error) {
    reject(error)
  }
  resolve(value) // may execute after reject
})
```

```javascript
new Promise((resolve, reject) => {
  reject(error)
  resolve(value) // always executes after reject
})
```

```javascript
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
```

```javascript
new Promise((resolve, reject) => {
  if (error) {
    reject(error)
    return
  }
  resolve(value)
})
```

```javascript
new Promise(async (resolve, reject) => {
  try {
    const r = await foo()
    resolve(r) // last throwable expression → catch cannot run after this
  } catch (error) {
    reject(error)
  }
})
```

## Original Documentation

https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/no-multiple-resolved.md
