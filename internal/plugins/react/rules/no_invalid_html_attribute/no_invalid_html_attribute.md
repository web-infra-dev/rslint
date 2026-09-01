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

## Original Documentation

- [eslint-plugin-react: no-invalid-html-attribute](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-invalid-html-attribute.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-invalid-html-attribute.js)
