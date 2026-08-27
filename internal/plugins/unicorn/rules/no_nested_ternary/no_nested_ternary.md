# no-nested-ternary

📝 Disallow nested ternary expressions.

💼🚫 This rule is enabled in the ✅ `recommended` config.

🔧 This rule is automatically fixable by the [`--fix` CLI option](https://eslint.org/docs/latest/user-guide/command-line-interface#--fix).

Improved version of the [`no-nested-ternary`](https://eslint.org/docs/latest/rules/no-nested-ternary) ESLint rule. This rule allows cases where the nested ternary is only one level and wrapped in parens.

Unparenthesized or deeply nested ternaries make readers track multiple conditions and branches at once, so this rule permits only clearly parenthesized single-level nesting.

## Examples

```js
// ❌
const foo = i > 5 ? i < 100 ? true : false : true;

// ✅
const foo = i > 5 ? (i < 100 ? true : false) : true;
```

```js
// ❌
const foo = i > 5 ? true : (i < 100 ? true : (i < 1000 ? true : false));
```

```js
// ✅
const foo = i > 5 || i < 100 || i < 1000;
```

## Partly fixable

This rule is only fixable when the nesting is up to one level. The rule will wrap the nested ternary in parens:

```js
const foo = i > 5 ? i < 100 ? true : false : true
```

will get fixed to

```js
const foo = i > 5 ? (i < 100 ? true : false) : true
```

## Disabling ESLint `no-nested-ternary`

We recommend disabling the ESLint `no-nested-ternary` rule in favor of this one:

```js
{
	rules: {
		'no-nested-ternary': 'off',
	},
}
```

## Original Documentation

- [eslint-plugin-unicorn: no-nested-ternary](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-nested-ternary.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-nested-ternary.js)
