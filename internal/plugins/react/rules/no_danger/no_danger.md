# no-danger

Disallow usage of dangerous JSX properties.

## Rule Details

Dangerous properties in React are those whose behavior is known to be a common
source of application vulnerabilities. The names of these properties make their
dangerous nature explicit. Currently only `dangerouslySetInnerHTML` is checked.

Examples of **incorrect** code for this rule:

```javascript
var React = require('react');

var Hello = <div dangerouslySetInnerHTML={{ __html: 'Hello World' }}></div>;
```

Examples of **correct** code for this rule:

```javascript
var React = require('react');

var Hello = <div>Hello World</div>;
```

## Rule Options

```json
{ "react/no-danger": ["error", { "customComponentNames": ["*Panel", "Widget"] }] }
```

### `customComponentNames`

An array of component name patterns. Defaults to `[]` (off for custom
components). Patterns use `*` as a wildcard (minimatch-style). Set to
`["*"]` to check every custom component, or list specific names such as
`["MyComponent", "Sub*"]` to opt in by name.

Examples of **incorrect** code with `{ "customComponentNames": ["*"] }`:

```json
{ "react/no-danger": ["error", { "customComponentNames": ["*"] }] }
```

```javascript
<App dangerouslySetInnerHTML={{ __html: '<span>hello</span>' }} />;
```

## When Not To Use It

Disable this rule if you know the content passed to `dangerouslySetInnerHTML`
has already been sanitized.

## Differences from ESLint

- **Deep member tag names.** For `<A.B.C dangerouslySetInnerHTML={...} />`,
  `eslint-plugin-react` flattens only one level of member access, producing
  the string `"undefined.C"` when matching against `customComponentNames`.
  rslint uses the full source-text name (`"A.B.C"`), so patterns such as
  `["A.B.C"]` or `["A.*"]` match deep member tags as intended.
- **JSX namespaced tag names.** For `<ns:name dangerouslySetInnerHTML={...} />`,
  `eslint-plugin-react` passes an Identifier AST node (not a string) into
  `minimatch`, which throws `f.split is not a function` at runtime — the
  `customComponentNames` branch crashes on namespaced tags in ESLint. rslint
  uses the string `"ns:name"`, so patterns match as expected.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-danger.md
