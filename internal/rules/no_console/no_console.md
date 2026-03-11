# no-console

## Rule Details

Disallow the use of `console`. In environments where `console` is not intended (such as production code), using `console` may be considered a debugging leftover.

Examples of **incorrect** code for this rule:

```javascript
console.log('message');
console.warn('warning');
console.error('error');
```

Examples of **correct** code for this rule:

```javascript
// With option { "allow": ["warn", "error"] }
console.warn('warning');
console.error('error');
```

## Options

- `allow`: An array of console method names that are allowed (e.g., `["warn", "error"]`).

## Original Documentation

https://eslint.org/docs/latest/rules/no-console
