# no-await-expression-member

## Rule Details

Disallow accessing a property or element directly from an `await` expression.
Use destructuring or store the awaited result in a variable before accessing
its members.

Examples of **incorrect** code for this rule:

```javascript
const property = (await getObject()).property;
const secondElement = (await getArray())[1];
const data = await (await fetch('/foo')).json();
```

Examples of **correct** code for this rule:

```javascript
const { property } = await getObject();
const [, secondElement] = await getArray();
const response = await fetch('/foo');
const data = await response.json();
```

The rule has no options. It automatically fixes variable declarations that
access element `0` or `1`, or a named property matching the variable name.
Optional member access and declarations with a type annotation are reported
without an automatic fix.

## Differences from ESLint

Unlike eslint-plugin-unicorn v74.0.0, rslint reports `using` and `await using`
declarations without automatic fixes, because these declarations cannot use destructuring.

## Original Documentation

- [eslint-plugin-unicorn: no-await-expression-member](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-await-expression-member.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-await-expression-member.js)
