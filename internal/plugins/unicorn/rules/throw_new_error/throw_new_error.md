# throw-new-error

## Rule Details

Require `new` when creating an error.

Calling an error constructor without `new` still produces an error object, but
it is easy to mistake for an ordinary function call, and it breaks as soon as
the class is subclassed or its prototype is relied upon.

Although the rule is named for `throw`, it applies to every call expression, so
a bare call such as `const error = Error()` is reported too.

Examples of **incorrect** code for this rule:

```javascript
throw Error();
throw TypeError('message');
throw lib.CustomError();
```

Examples of **correct** code for this rule:

```javascript
throw new Error();
throw new TypeError('message');
throw new lib.CustomError();
```

Calls that `new` cannot be applied to are left alone — an optional-chained call
such as `Error?.()` or `lib?.Error()`, a computed access such as
`lib["Error"]()`, and a decorator call such as `@RegisterServiceError()`.

## Original Documentation

- [eslint-plugin-unicorn: throw-new-error](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/throw-new-error.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/throw-new-error.js)
