# no-unreadable-new-expression

## Rule Details

Disallow accessing a member directly from a `new` expression and disallow
complex expressions as constructors. Keeping construction separate from later
member access avoids JavaScript's precedence-sensitive `new` syntax.

Examples of **incorrect** code for this rule:

```javascript
const timestamp = new Date().getTime();
const value = new (factory().Constructor)();
const item = new constructors[Kind]();
```

Examples of **correct** code for this rule:

```javascript
const date = new Date();
const timestamp = date.getTime();

const { Constructor } = factory();
const value = new Constructor();

const ConstructorForKind = constructors[Kind];
const item = new ConstructorForKind();
```

Identifier constructors and static member constructors such as
`new Namespace.Constructor()` are allowed.

## Original Documentation

- [eslint-plugin-unicorn: no-unreadable-new-expression](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-unreadable-new-expression.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-unreadable-new-expression.js)
