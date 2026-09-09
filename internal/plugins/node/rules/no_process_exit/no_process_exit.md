# node/no-process-exit

## Rule Details

Disallows the use of `process.exit()`.

The `process.exit()` method in Node.js is used to force the process to exit as quickly as possible even if there are still asynchronous operations pending that have not yet completed fully, including I/O operations to `process.stdout` and `process.stderr`.

In most cases, it is not necessary to explicitly call `process.exit()`. The Node.js process will exit on its own if there is no additional work pending in the event loop. Calling `process.exit()` will force the process to exit as quickly as possible even if there are still asynchronous operations pending.

Examples of **incorrect** code for this rule:

```javascript
process.exit(1);
process.exit(0);
```

Examples of **correct** code for this rule:

```javascript
throw new Error("an error occurred");
```

## Original Documentation

- [eslint-plugin-n: no-process-exit](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-process-exit.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-process-exit.js)
