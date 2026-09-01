# no-invalid-html-attribute

## Rule Details

This rule disallows invalid values for selected HTML attributes. By default, it
validates `rel` values on HTML JSX elements and `React.createElement` calls.
It reports values that are unknown, used on the wrong element, not separated by
single spaces, or used without a required companion value.

Examples of **incorrect** code for this rule:

```javascript
<a rel="canonical" />;
<link rel="shortcut" />;
<a rel="noopener  noreferrer" />;
React.createElement('a', { rel: 'not-a-rel-value' });
```

Examples of **correct** code for this rule:

```javascript
<a rel="noopener noreferrer" />;
<link rel="canonical" />;
<link rel="shortcut icon" />;
React.createElement('form', { rel: 'external' });
```

## Options

The rule accepts one positional array listing the attributes to validate. The
only supported attribute is `rel`; omitting the option is equivalent to
selecting `rel`.

Examples of **correct** code for this rule with no selected attributes:

```json
{ "react/no-invalid-html-attribute": ["error", []] }
```

```javascript
<a rel="not-a-rel-value" />;
```

## Differences from ESLint

For `React.createElement` calls, rslint only validates `rel` values that are
string literals. Numeric, BigInt, boolean, `null`, and RegExp literals are not
sent through the string-value lookup or suggestion path:

```javascript
React.createElement('a', { rel: 1 });
React.createElement('a', { rel: 1n });
React.createElement('a', { rel: /invalid/ });
```

This intentionally differs from eslint-plugin-react v7.37.5. Its
`checkPropValidValue` branch treats every ESTree `Literal` as though
`value.value` were a string. That can produce malformed suggestions such as
`rel:  ` for a number or `rel: //` for a RegExp. rslint skips these non-string
values so suggestions never produce syntactically invalid code.

## Original Documentation

- [eslint-plugin-react: no-invalid-html-attribute](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-invalid-html-attribute.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-invalid-html-attribute.js)
