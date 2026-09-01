# camelcase

## Rule Details

This rule enforces camelcase naming by reporting identifiers with internal
underscores. Leading and trailing underscores and all-uppercase constants are
allowed.

Examples of **incorrect** code for this rule:

```javascript
const favorite_color = "blue";
obj.favorite_color = "blue";
```

Examples of **correct** code for this rule:

```javascript
const favoriteColor = "blue";
const FAVORITE_COLOR = "blue";
const _favoriteColor = "blue";
obj.favorite_color();
```

## Options

- `"properties": "always"` (default) checks property declarations and writes;
  `"never"` permits underscored property names.
- `"ignoreDestructuring": true` permits a destructured binding when it keeps
  the source property name, while later uses are still checked.
- `"ignoreImports": true` permits a named import when its local and exported
  names are identical, while later uses are still checked.
- `"ignoreGlobals": true` skips configured and comment-declared globals;
  unresolved names are still checked.
- `"allow"` accepts exact names or JavaScript regular expression patterns.

Examples of **correct** code with property checks disabled:

```json
{ "camelcase": ["error", { "properties": "never" }] }
```

```javascript
const response = { response_code: 200 };
response.response_code = 201;
```

Examples of **correct** code with destructured source names ignored:

```json
{ "camelcase": ["error", { "ignoreDestructuring": true }] }
```

```javascript
const { response_code } = response;
```

Examples of **correct** code with an allow pattern:

```json
{ "camelcase": ["error", { "allow": ["^UNSAFE_"] }] }
```

```javascript
function UNSAFE_componentWillMount() {}
```

## Original Documentation

- [ESLint: camelcase](https://eslint.org/docs/latest/rules/camelcase)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/camelcase.js)
