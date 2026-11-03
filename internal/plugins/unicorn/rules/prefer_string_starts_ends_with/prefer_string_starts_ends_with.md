# prefer-string-starts-ends-with

## Rule Details

Prefer `String#startsWith()` and `String#endsWith()` over simple anchored
regular expressions, `indexOf()` comparisons with zero, and string slice
comparisons.

Examples of **incorrect** code for this rule:

```javascript
/^prefix/.test(value);
/suffix$/.test(value);
'prefix-value'.indexOf('prefix') === 0;
'prefix-value'.slice(0, 6) === 'prefix';
'value-suffix'.slice(-6) === 'suffix';
```

Examples of **correct** code for this rule:

```javascript
value.startsWith('prefix');
value.endsWith('suffix');
'prefix-value'.startsWith('prefix');
'value-suffix'.endsWith('suffix');
```

The rule ignores regular expressions with case-insensitive or multiline flags
and suffix patterns with the sticky flag. It reports `indexOf()` and `slice()`
comparisons only when the receiver is known to be a string. Slice comparisons
also require a string search value and matching bounds. A dynamic suffix length
is checked only when the suffix is statically known to be nonempty.

Automatic fixes preserve comments within copied operands. Reports whose fixes
would remove other comments have no automatic fix. A reversed slice comparison
offers a suggestion when moving the receiver would change evaluation order.
For regex arguments of unknown type, suggestions also offer string conversion,
optional chaining, and nullish coalescing.

## Options

This rule has no options.

## Differences from upstream

When configured with a TypeScript project, upstream can report or skip
additional expressions based on inferred types. This rule follows upstream's
behavior without that inference. For example, the first comparison below is
left unchanged by rslint. An explicit string assertion on the call result
enables the check:

```typescript
declare const read: () => string;

read().indexOf('a') === 0; // Left unchanged
(read() as string).indexOf('a') === 0; // Reported and fixed
```

Functions with an explicit string return annotation, such as
`function read(): string { return 'abc'; }`, are checked without an assertion.

## Original Documentation

- [eslint-plugin-unicorn: prefer-string-starts-ends-with](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-string-starts-ends-with.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-starts-ends-with.js)
