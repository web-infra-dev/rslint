# no-invalid-character-comparison

## Rule Details

Disallow comparing a single UTF-16 code unit from a string with a string longer
than one code unit. Such equality comparisons are always `false`, and inequality
comparisons are always `true`. This often catches a typo such as `'/n'` instead of
`'\n'`.

The rule checks `charAt()`, `at()`, and bracket access with a static numeric key.
It checks only values known to be strings, so arrays and user-defined methods are
ignored. Optional character access and dynamic index keys are also ignored.

Examples of **incorrect** code for this rule:

```javascript
'unicorn'.charAt(0) === 'ab';
'unicorn'.charAt(3) === '/n';
'unicorn'.at(0) !== 'ab';
'unicorn'[0] === 'ab';
'😀'.at(0) === '😀';
```

Examples of **correct** code for this rule:

```javascript
'unicorn'.charAt(0) === 'u';
'unicorn'.charAt(3) === '\n';
'unicorn'.at(0) === 'a';
'unicorn'[0] === 'a';
'unicorn'.charAt(100) === '';
```

An empty string is allowed because `charAt()` returns it for an out-of-bounds
index. Astral characters such as emoji contain two UTF-16 code units and are
reported even though they appear as a single character.

## Options

This rule has no options.

## Differences from upstream

rslint may leave a comparison unreported when a function's return type is known
only through a callable type. Unicorn can report it when TypeScript project
checking is enabled. A direct string assertion enables the check in rslint:

```typescript
declare const read: () => string;

read().charAt(0) === 'ab'; // Left unchanged
(read() as string).charAt(0) === 'ab'; // Reported
```

The same limitation applies to generic type aliases whose result depends on a
type parameter, and very deeply nested type aliases. A direct `string`
annotation or assertion enables the check.

Locally declared functions named `String` are not assumed to return strings.
For example, this array comparison is allowed by rslint but reported by Unicorn
v79.0.0:

```javascript
function String() {
  return ['ab'];
}

String().at(0) === 'ab';
```

## Original Documentation

- [eslint-plugin-unicorn: no-invalid-character-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-invalid-character-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-invalid-character-comparison.js)
