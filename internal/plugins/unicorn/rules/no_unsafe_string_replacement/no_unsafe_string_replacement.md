# no-unsafe-string-replacement

## Rule Details

Disallow non-literal replacement values in `String#replace()` and
`String#replaceAll()`.

String replacement patterns such as `$&`, `$1`, and `` $` `` are expanded even
when the replacement value comes from an expression. This can produce
unexpected output or security bugs. Use a literal string for static content and
a replacement function for dynamic content.

Examples of **incorrect** code for this rule:

```javascript
template.replace('{url}', htmlEscape(url));
template.replaceAll('{url}', replacement);
```

Examples of **correct** code for this rule:

```javascript
template.replace('{url}', 'https://example.com');
template.replace('{url}', () => htmlEscape(url));
template.replaceAll('{url}', String.raw`https://example.com`);
```

## Original Documentation

- [eslint-plugin-unicorn: no-unsafe-string-replacement](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-unsafe-string-replacement.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-unsafe-string-replacement.js)
