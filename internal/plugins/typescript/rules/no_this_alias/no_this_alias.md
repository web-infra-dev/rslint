# no-this-alias

## Rule Details

Disallow aliasing `this`.

Assigning `this` to a variable (commonly named `self` or `that`) is a legacy pattern that predates arrow functions. Arrow functions automatically capture the surrounding `this`, making `this` aliasing unnecessary.

Examples of **incorrect** code for this rule:

```typescript
const self = this;
let that = this;
const foo = this;
```

Examples of **correct** code for this rule:

```typescript
const { foo } = this; // destructuring is allowed by default
setTimeout(() => {
  this.doSomething(); // use arrow function instead of aliasing
});
```

## Original Documentation

- [typescript-eslint no-this-alias](https://typescript-eslint.io/rules/no-this-alias)
