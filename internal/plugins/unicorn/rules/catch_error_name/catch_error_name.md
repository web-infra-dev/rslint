# catch-error-name

## Rule Details

Enforce a specific parameter name in catch clauses and Promise rejection handlers.

Consistent error names make handlers easier to scan. The default name is `error`;
descriptive names ending in `error` or `Error` are accepted, as are names matching
an `ignore` pattern. An unused `_` parameter is also accepted. The rule can
automatically rename the declaration and all references, appending underscores
when necessary to avoid shadowing.

Examples of **incorrect** code for this rule:

```javascript
try {} catch (badName) {}
promise.catch(badName => {});
promise.then(undefined, badName => {});
```

Examples of **correct** code for this rule:

```javascript
try {} catch (error) {}
promise.catch(error => {});
promise.then(undefined, error => {});
try {} catch (fsError) {}
```

## Options

- `name` (`string`, default: `"error"`) sets the expected name.
- `ignore` (`string[]`, default: `[]`) contains JavaScript regular-expression patterns to ignore.

## Differences from ESLint

When an unused handler parameter would be renamed to the same name as a direct
lexical declaration in the handler body, rslint appends an underscore. ESLint
currently emits a fix that creates a duplicate binding and invalid syntax in
this case.

## Original Documentation

- [eslint-plugin-unicorn: catch-error-name](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/catch-error-name.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/catch-error-name.js)
