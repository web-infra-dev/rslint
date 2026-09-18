# prefer-optional-catch-binding

## Rule Details

Omit unused catch bindings. A reference from a nested scope counts as a use, as do writes to the caught variable.

Unused destructuring bindings are also checked. Comments surrounding the binding are preserved; a binding containing comments is reported without an automatic fix. Binding defaults are not removed.

Examples of **incorrect** code:

```javascript
try { work(); } catch (error) {}
```

Examples of **correct** code:

```javascript
try { work(); } catch {}
```

## Options

This rule has no options. It is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Original Documentation

- [eslint-plugin-unicorn: prefer-optional-catch-binding](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-optional-catch-binding.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-optional-catch-binding.js)
