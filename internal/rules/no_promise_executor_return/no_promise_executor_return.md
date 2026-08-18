# no-promise-executor-return

## Rule Details

The `new Promise` constructor accepts a single argument, called an _executor_. The executor's return value is ignored: it cannot be read and it does not affect the promise in any way, so returning a value from it is usually a mistake.

This rule disallows returning values from Promise executor functions. Only `return` without a value is allowed, as it's a control flow statement.

Examples of **incorrect** code for this rule:

```javascript
new Promise((resolve, reject) => {
  if (someCondition) {
    return defaultResult;
  }
  getSomething((err, result) => {
    if (err) {
      reject(err);
    } else {
      resolve(result);
    }
  });
});

new Promise((resolve, reject) =>
  getSomething((err, data) => {
    if (err) {
      reject(err);
    } else {
      resolve(data);
    }
  }),
);

new Promise(() => {
  return 1;
});

new Promise((r) => r(1));
```

Examples of **correct** code for this rule:

```javascript
// Turn the inline return into two lines
new Promise((resolve, reject) => {
  if (someCondition) {
    resolve(defaultResult);
    return;
  }
  getSomething((err, result) => {
    if (err) {
      reject(err);
    } else {
      resolve(result);
    }
  });
});

// Add curly braces
new Promise((resolve, reject) => {
  getSomething((err, data) => {
    if (err) {
      reject(err);
    } else {
      resolve(data);
    }
  });
});

new Promise((r) => {
  r(1);
});

// or just use Promise.resolve
Promise.resolve(1);
```

## Options

This rule takes one option, an object, with the following properties:

- `allowVoid`: If set to `true` (`false` by default), this rule will allow returning void values.

### allowVoid

Examples of **correct** code for this rule with the `{ "allowVoid": true }` option:

```json
{ "no-promise-executor-return": ["error", { "allowVoid": true }] }
```

```javascript
new Promise((resolve, reject) => {
  if (someCondition) {
    return void resolve(defaultResult);
  }
  getSomething((err, result) => {
    if (err) {
      reject(err);
    } else {
      resolve(result);
    }
  });
});

new Promise(
  (resolve, reject) =>
    void getSomething((err, data) => {
      if (err) {
        reject(err);
      } else {
        resolve(data);
      }
    }),
);

new Promise((r) => void r(1));
```

## Original Documentation

- [ESLint: no-promise-executor-return](https://eslint.org/docs/latest/rules/no-promise-executor-return)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-promise-executor-return.js)
