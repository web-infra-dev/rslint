# error-message

## Rule Details

Enforce passing a `message` value when creating a built-in error.

This rule enforces a `message` value to be passed in when creating an instance of a built-in `Error` object. It ensures that errors have a descriptive message associated with them, which aids in debugging.

Examples of **incorrect** code for this rule:

```javascript
throw new Error();
throw Error();
throw new Error('');
throw new Error(false);
throw new Error([]);
const foo = new TypeError();
new AggregateError(errors);
new SuppressedError(error, suppressed);
```

Examples of **correct** code for this rule:

```javascript
throw new Error('error');
throw new TypeError('error');
throw new MyCustomError();
new AggregateError(errors, 'message');
new SuppressedError(error, suppressed, 'message');
```

## Original Documentation

- [eslint-plugin-unicorn: error-message](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/error-message.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/error-message.js)
