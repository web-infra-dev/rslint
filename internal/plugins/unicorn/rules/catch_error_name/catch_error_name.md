# catch-error-name

Enforce a specific parameter name in catch clauses and Promise rejection handlers.

Consistent error names make handlers easier to scan. The default name is `error`;
descriptive names ending in `error` or `Error` are accepted, as are names matching
an `ignore` pattern. An unused `_` parameter is also accepted.

## Options

- `name` (`string`, default: `"error"`) sets the expected name.
- `ignore` (`string[]`, default: `[]`) contains JavaScript regular-expression patterns to ignore.

This rule is automatically fixable. Fixes rename the declaration and all of its
references, appending underscores when necessary to avoid shadowing another name.

## Upstream

- [Documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/catch-error-name.md)
- [Source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/catch-error-name.js)
