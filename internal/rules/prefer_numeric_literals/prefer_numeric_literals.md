# prefer-numeric-literals

## Rule Details

This rule disallows `parseInt()` and `Number.parseInt()` when binary, octal,
or hexadecimal numeric literals can be used instead.

Examples of **incorrect** code for this rule:

```javascript
parseInt("111110111", 2) === 503;
parseInt(`767`, 8) === 503;
Number.parseInt("1F7", 16) === 255;
```

Examples of **correct** code for this rule:

```javascript
0b111110111 === 503;
0o767 === 503;
0x1F7 === 503;

parseInt(foo, 2);
parseInt("11", 10);
Number.parseInt("11", 36);
```

## Original Documentation

- [ESLint prefer-numeric-literals](https://eslint.org/docs/latest/rules/prefer-numeric-literals)
