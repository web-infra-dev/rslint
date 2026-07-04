# no-thenable

## Rule Details

Disallow adding a `then` property to objects or classes, and disallow exporting
a binding named `then`.

Objects with a `then` property can be treated as thenables when they are
awaited. Modules with a named `then` export can also behave unexpectedly with
dynamic `import()`.

Computed property names are also checked when they can be statically resolved to `then`.

Examples of **incorrect** code for this rule:

```javascript
const value = {
  then() {},
};

class Value {
  then() {}
}

foo.then = () => {};

export function then() {}
```

Examples of **correct** code for this rule:

```javascript
const value = {
  success() {},
};

class Value {
  success() {}
}

foo.success = () => {};

export function success() {}
```

## Original Documentation

- [eslint-plugin-unicorn: no-thenable](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-thenable.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/no-thenable.js)
