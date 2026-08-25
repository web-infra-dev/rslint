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

This rule uses type information to decide whether the receiver is a string, and
is skipped on files where type information is unavailable.

## Differences from the original rule

The original rule classifies the receiver syntactically, from its type
annotation, before it consults type information. This rule always uses type
information, and reports unless the type shows the receiver is not a string.
Where the two classifiers disagree, this rule follows the type:

- A receiver typed as a no-substitution template literal type (`` `x` ``) is
  reported. The original rule treats every literal type whose literal is not a
  plain string literal as a non-string, but such a type is a string.
- A receiver whose type reduces to `never`, such as `string & number`, is not
  reported. The receiver is uninhabited, so there is nothing to report on.
- A receiver whose numeric or bigint literal type is inferred rather than
  annotated, such as `const value = 1`, `(1)`, or `1 | 2` after narrowing, is
  not reported. The original rule reports these because it has no annotation to
  classify and leaves such types unknown; the inferred type is the same type as
  the annotated one, which is provably not a string. Calling `replace` on those
  receivers is a TypeScript error in the first place.

## Original Documentation

- [eslint-plugin-unicorn: no-unsafe-string-replacement](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-unsafe-string-replacement.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-unsafe-string-replacement.js)
