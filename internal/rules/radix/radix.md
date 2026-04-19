# radix

## Rule Details

This rule enforces that `parseInt()` and `Number.parseInt()` are called with an
explicit radix argument. Without a radix, `parseInt` defaults to decimal for
most inputs but previously inferred `16` for strings starting with `0x` (and
some implementations inferred `8` for strings starting with `0`), which made
the behavior hard to predict. Passing an explicit radix eliminates the
ambiguity.

The rule reports three kinds of problems:

- **No arguments** — `parseInt()` is called with no arguments at all.
- **Missing radix** — `parseInt(s)` is called with only the string argument.
  A suggestion fix is offered that inserts `, 10` (or ` 10,` when a trailing
  comma is present) to pass the decimal radix explicitly.
- **Invalid radix** — the second argument is a literal (or the identifier
  `undefined`) that cannot be an integer between `2` and `36`.

Examples of **incorrect** code for this rule:

```javascript
parseInt();
parseInt("071");
parseInt("071", "abc");
parseInt("071", 37);
parseInt("071", 10.5);
Number.parseInt();
Number.parseInt("071");
```

Examples of **correct** code for this rule:

```javascript
parseInt("071", 10);
parseInt("071", 8);
parseInt("071", foo);
Number.parseInt("071", 10);
parseFloat(someValue);
```

## Options

The rule accepts a deprecated string option (`"always"` or `"as-needed"`). It
is preserved for backward compatibility and does not change the rule's
behavior — a radix argument is always required regardless of the option.

## Original Documentation

- [ESLint rule documentation](https://eslint.org/docs/latest/rules/radix)
