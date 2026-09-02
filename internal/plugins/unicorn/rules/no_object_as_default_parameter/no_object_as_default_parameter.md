# no-object-as-default-parameter

## Rule Details

Disallow non-empty object literals as function-parameter defaults. Passing a
partial object replaces the entire default object, so individual defaulted
properties should usually be expressed with parameter destructuring instead.

Examples of **incorrect** code for this rule:

```javascript
function read(options = { cache: true, encoding: "utf8" }) {}

function connect({ secure } = { secure: true }) {}
```

Examples of **correct** code for this rule:

```javascript
function read({ cache = true, encoding = "utf8" } = {}) {}

function read(options = {}) {}

const defaults = { cache: true };
function read(options = defaults) {}
```

The rule is syntax-based. A TypeScript annotation does not exempt a non-empty
object literal default:

```typescript
function range(
  options: { minimum: number; maximum: number } = {
    minimum: 0,
    maximum: 10,
  },
) {}
```

## Original Documentation

- [eslint-plugin-unicorn: no-object-as-default-parameter](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/no-object-as-default-parameter.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-object-as-default-parameter.js)
