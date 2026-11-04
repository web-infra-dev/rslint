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

When TypeScript project information is available, the rule also checks strings
returned through callable types or stored in typed properties. Without project
information, use a direct string annotation or assertion for values whose types
cannot otherwise be determined.

## Options

This rule has no options.

## Differences from upstream

Locally declared functions named `String` are not assumed to return strings.
For example, this array comparison is allowed by rslint but reported by Unicorn
v79.0.0:

```javascript
function String() {
  return ['ab'];
}

String().at(0) === 'ab';
```

Strings containing an unpaired UTF-16 surrogate are still checked, but the quoted
value in rslint's JavaScript API diagnostic message uses replacement characters.
For example, for `"x"[0] === "\uD800a"`, rslint's message quotes `���a`, while
Unicorn preserves the unpaired surrogate followed by `a`. The diagnostic range
and comparison result are the same.

## Original Documentation

- [eslint-plugin-unicorn: no-invalid-character-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-invalid-character-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-invalid-character-comparison.js)
