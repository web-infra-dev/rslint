# no-document-cookie

## Rule Details

Disallow assigning to `document.cookie` directly. Constructing cookie strings
by hand is error-prone; prefer the Cookie Store API or a cookie library.

Examples of **incorrect** code for this rule:

```javascript
document.cookie = 'name=value; Path=/; Secure';
document.cookie += '; SameSite=Lax';
```

Examples of **correct** code for this rule:

```javascript
await cookieStore.set({
  name: 'name',
  value: 'value',
  path: '/',
});

const cookies = document.cookie.split('; ');
```

The rule follows aliases of the global `document` object and its standard
global-object forms, so these assignments are also reported:

```javascript
const doc = globalThis.document;
doc.cookie = 'name=value';

window.document.cookie = 'name=value';
```

## Original Documentation

- [eslint-plugin-unicorn: no-document-cookie](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-document-cookie.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-document-cookie.js)
