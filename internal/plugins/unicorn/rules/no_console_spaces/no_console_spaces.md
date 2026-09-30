# no-console-spaces

## Rule Details

Disallow redundant spaces between arguments to `console.log`, `console.debug`, `console.info`, `console.warn`, and `console.error`.

These methods already separate arguments with a space. This rule removes a single leading space from a string or template literal after the first argument, and a single trailing space before the last argument.

Examples of **incorrect** code for this rule:

```javascript
console.log('abc ', 'def');
console.log('abc', ' def');
console.warn(`value: `, value);
```

Examples of **correct** code for this rule:

```javascript
console.log('abc', 'def');
console.warn('value:', value);
console.log(' abc', 'def ');
console.log('abc  ', 'def');
console.log(' ', 'def');
console.log('abc\n', 'def');
```

The rule preserves consecutive spaces, strings containing only one space, and spaces at the outer edges of the first and last arguments. It checks literal source text, so escaped spaces such as `\u0020` are preserved. Computed methods and optional calls are ignored.

## Options

This rule has no options. It provides an automatic fix.

## Differences from upstream

When a reported trailing space is written as `\ `, rslint removes both the
backslash and the space. For example:

```javascript
// Before
console.log('abc\ ', 'def');

// After the automatic fix
console.log('abc', 'def');
```

This also applies to template literals. Unlike eslint-plugin-unicorn v76.0.0,
the fix leaves valid JavaScript instead of leaving a backslash that escapes
the closing quote or backtick. Escaped backslashes (`\\`) are preserved.

## Original Documentation

- [eslint-plugin-unicorn: no-console-spaces](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/no-console-spaces.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/no-console-spaces.js)
