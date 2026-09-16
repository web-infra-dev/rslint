# handle-callback-err

Requires callbacks to reference their error parameter.

## Rule details

The rule checks the first parameter binding of each function, arrow function, or method. If its name matches the configured error name and it has no references, the rule reports the function.

Examples of **incorrect** code:

```javascript
function loadData(err, data) {
  processData(data);
}
```

Examples of **correct** code:

```javascript
function loadData(err, data) {
  if (err) {
    console.error(err);
    return;
  }
  processData(data);
}
```

A reference inside a nested function counts. A different variable with the same name does not. The rule checks references, so assigning to the parameter or giving it a default value also counts; it does not prove that the error is handled at runtime. Destructuring uses the first bound name, skipping empty patterns.

## Options

The single option is a string. It defaults to `"err"`; an empty string also uses this default. Other names replace the default.

```javascript
export default [
  {
    plugins: ["node"],
    rules: {
      "node/handle-callback-err": ["error", "^(err|error)$"],
    },
  },
];
```

A string beginning with `^` is a JavaScript regular expression with the Unicode (`u`) flag. For example, `"^.+Error$"` matches `connectionError`. Strings without a leading `^` match a parameter name exactly.

Invalid patterns such as `"^["` cause a configuration error before linting starts.

This rule has no automatic fix or suggestions.

## Differences from upstream

Use Unicode category abbreviations such as `\p{L}` instead of long property names such as `\p{Letter}`. For example, rslint rejects `"^\\p{Letter}+$"` as an unsupported pattern, while upstream accepts it and reports `function f(错误) {}`. Changing the option to `"^\\p{L}+$"` makes both report it.

## Original documentation

- [eslint-plugin-n: handle-callback-err](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/handle-callback-err.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/handle-callback-err.js)
