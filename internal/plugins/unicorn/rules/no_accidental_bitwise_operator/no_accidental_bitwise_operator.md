# no-accidental-bitwise-operator

## Rule Details

Disallow bitwise operators where a logical operator was likely intended. The
rule reports `&` when an identifier is followed by a property access on the same
identifier, such as `object & object.property`. It also reports `|` and `|=` when
the right operand is a string, template literal, boolean, array, object, class,
or function expression used as an apparent fallback.

### Incorrect

```javascript
if (object & object.property) {
  use(object.property);
}

options = options | {};
input |= '';
```

### Correct

```javascript
if (object && object.property) {
  use(object.property);
}

options = options || {};
input ||= '';

const masked = flags & MASK;
const truncated = value | 0;
```

## Suggestions

Editor suggestions replace the operator with `&&`, `||`, or `||=` and add
parentheses when needed to preserve how the surrounding operators group.
Review each suggestion before applying it because logical operators
short-circuit and can change runtime behavior.

## Options

This rule has no options.

## Original Documentation

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-accidental-bitwise-operator.md)
- [Upstream rule source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-accidental-bitwise-operator.js)
