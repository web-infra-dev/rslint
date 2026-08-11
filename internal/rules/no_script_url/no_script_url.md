# no-script-url

## Rule Details

Disallow `javascript:` URLs.

Using `javascript:` URLs is considered by some as a form of `eval`. Code passed in `javascript:` URLs has to be parsed and evaluated by the browser in the same way that `eval` is processed.

Examples of **incorrect** code for this rule:

```javascript
location.href = 'javascript:void(0)';

location.href = `javascript:void(0)`;
```

Examples of **correct** code for this rule:

```javascript
location.href = 'https://example.com';
```

## Original Documentation

- [ESLint: no-script-url](https://eslint.org/docs/latest/rules/no-script-url)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-script-url.js)

https://eslint.org/docs/latest/rules/no-script-url
