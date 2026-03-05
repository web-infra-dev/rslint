# no-async-promise-executor

## Rule Details

Disallows passing an `async` function as the executor to `new Promise()`. Using an async executor is usually a mistake because if the async executor throws an error, the error will be lost and will not cause the newly-constructed Promise to reject. Additionally, if a Promise executor uses `await`, this is usually a sign that it is not actually necessary to use the `new Promise` constructor.

Examples of **incorrect** code for this rule:

```javascript
const result = new Promise(async (resolve, reject) => {
  resolve(await foo);
});

const result = new Promise(async function (resolve, reject) {
  resolve(await foo);
});
```

Examples of **correct** code for this rule:

```javascript
const result = new Promise((resolve, reject) => {
  resolve(foo);
});

const result = new Promise(function (resolve, reject) {
  readFile('foo.txt', function (err, data) {
    if (err) reject(err);
    else resolve(data);
  });
});
```

## Original Documentation

- [ESLint no-async-promise-executor](https://eslint.org/docs/latest/rules/no-async-promise-executor)
