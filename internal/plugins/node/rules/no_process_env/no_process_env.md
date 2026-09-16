# no-process-env

## Rule details

Disallows `process.env` and `process["env"]` accesses. Keeping environment-dependent settings in a configuration module makes them easier to manage throughout a project.

Examples of **incorrect** code for this rule:

```javascript
if (process.env.NODE_ENV === "development") {
  enableDebugging();
}
```

Examples of **correct** code for this rule:

```javascript
const config = require("./config");

if (config.env === "development") {
  enableDebugging();
}
```

## Options

`allowedVariables` is an array of environment variable names that may be accessed. It defaults to an empty array.

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/no-process-env": ["error", { allowedVariables: ["NODE_ENV"] }],
    },
  },
];
```

With this configuration, `process.env.NODE_ENV` and `process["env"]["NODE_ENV"]` are allowed. Reading `process.env` itself or accessing other variable names still reports an error. Computed variable names such as `process.env[name]` are not exempted.

This rule has no automatic fix or suggestions.

## Differences from upstream

JavaScript files containing optional access to a private field, such as `process?.#env`, currently report a syntax error instead of this rule's diagnostics. Non-optional private-field access (`process.#env`) is checked normally.

## Original documentation

- [eslint-plugin-n: no-process-env](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-process-env.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-process-env.js)
