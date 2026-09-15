# prefer-string-trim-start-end

## Rule Details

Prefer `String#trimStart()` and `String#trimEnd()` over the legacy
`String#trimLeft()` and `String#trimRight()` aliases. The preferred names are
direction-independent and therefore clearer for both left-to-right and
right-to-left text.

Examples of **incorrect** code for this rule:

```javascript
const leading = value.trimLeft();
const trailing = value.trimRight();
```

Examples of **correct** code for this rule:

```javascript
const leading = value.trimStart();
const trailing = value.trimEnd();
```

Optional access is checked, while an optional call is left unchanged:

```javascript
value?.trimLeft(); // Reported and fixed to value?.trimStart()
value.trimLeft?.(); // Allowed
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-string-trim-start-end](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-string-trim-start-end.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-string-trim-start-end.js)
