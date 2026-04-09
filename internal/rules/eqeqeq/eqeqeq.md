# eqeqeq

## Rule Details

Requires the use of `===` and `!==` instead of `==` and `!=`. Using strict equality operators helps avoid unexpected type coercion in comparisons.

### Options

This rule supports three modes:

- `"always"` (default): Always require `===` and `!==`. Supports a `null` sub-option:
  - `"always"` (default): Enforce strict equality for null comparisons too
  - `"ignore"`: Skip enforcement for null comparisons
  - `"never"`: Enforce `==`/`!=` for null comparisons and `===`/`!==` for everything else
- `"smart"`: Allow `==` for typeof comparisons, same-type literal comparisons, and null checks
- `"allow-null"`: Shorthand for `["always", {"null": "ignore"}]`

Examples of **incorrect** code for this rule:

```javascript
a == b;
a != b;
typeof a == 'number'; // in "always" mode
```

Examples of **correct** code for this rule:

```javascript
a === b;
a !== b;
typeof a === 'number';

// With "smart" option:
typeof a == 'number';
null == a;
'hello' == 'world';

// With "allow-null" option:
a == null;
null != a;
```

## Original Documentation

- [ESLint eqeqeq](https://eslint.org/docs/latest/rules/eqeqeq)
