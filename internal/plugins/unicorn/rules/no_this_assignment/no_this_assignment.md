# no-this-assignment

## Rule Details

Disallow assigning `this` directly to a variable. Use `this` directly, an arrow
function, or `Function#bind()` instead.

Examples of **incorrect** code for this rule:

```javascript
const self = this;

setTimeout(function () {
  self.run();
});
```

Examples of **correct** code for this rule:

```javascript
setTimeout(() => {
  this.run();
});

setTimeout(
  function () {
    this.run();
  }.bind(this),
);
```

## Original Documentation

- [eslint-plugin-unicorn: no-this-assignment](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-this-assignment.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-this-assignment.js)
