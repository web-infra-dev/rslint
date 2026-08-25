# no-invalid-fetch-options

## Rule Details

Disallows request bodies in `fetch()` and `new Request()` calls whose method is
`GET` or `HEAD`. The Fetch Standard throws a `TypeError` for these combinations.

Examples of **incorrect** code for this rule:

```javascript
fetch('/', { body: 'foo=bar' });

new Request('/', { method: 'HEAD', body: 'foo=bar' });
```

Examples of **correct** code for this rule:

```javascript
fetch('/', { method: 'POST', body: 'foo=bar' });

new Request('/', { method: 'HEAD' });
```

## Original Documentation

- [eslint-plugin-unicorn: no-invalid-fetch-options](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-invalid-fetch-options.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-invalid-fetch-options.js)
