# no-nested-ternary

## Rule Details

Improved version of the core ESLint [`no-nested-ternary`](https://eslint.org/docs/latest/rules/no-nested-ternary) rule. It allows cases where the nested ternary is only one level and wrapped in parentheses.

Unparenthesized or deeply nested ternaries force readers to track multiple conditions and branches at once, so this rule permits only clearly parenthesized single-level nesting.

Examples of **incorrect** code for this rule:

```javascript
const foo = i > 5 ? i < 100 ? true : false : true;
```

```javascript
const foo = i > 5 ? true : (i < 100 ? true : (i < 1000 ? true : false));
```

```javascript
const foo = i > 5 ? true : (i < 100 ? (i > 50 ? false : true) : false);
```

Examples of **correct** code for this rule:

```javascript
const foo = i > 5 ? true : (i < 100 ? true : false);
```

```javascript
const foo = i > 5 ? (i < 100 ? true : false) : true;
```

```javascript
const foo = i > 5 ? (i < 100 ? true : false) : (i < 100 ? true : false);
```

```javascript
const foo = i > 5 || i < 100 || i < 1000;
```

## Original Documentation

- [eslint-plugin-unicorn: no-nested-ternary](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-nested-ternary.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-nested-ternary.js)
