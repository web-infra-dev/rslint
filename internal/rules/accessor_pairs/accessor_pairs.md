# accessor-pairs

Enforce getter and setter pairs in objects and classes.

## Rule Details

By default the rule flags setters declared without a matching getter. It can
optionally flag getters declared without a matching setter, and also extends
to class bodies and TypeScript type-literal / interface members.

Examples of **incorrect** code for this rule:

```javascript
const obj = {
  set a(value) {
    this.val = value;
  },
};

const obj2 = { d: 1 };
Object.defineProperty(obj2, 'c', {
  set: function (value) {
    this.val = value;
  },
});
```

Examples of **correct** code for this rule:

```javascript
const obj = {
  set a(value) {
    this.val = value;
  },
  get a() {
    return this.val;
  },
};
```

## Options

All options are boolean with the following defaults:

- `setWithoutGet` (default `true`): report setters without a matching getter.
- `getWithoutSet` (default `false`): report getters without a matching setter.
- `enforceForClassMembers` (default `true`): apply to class declarations and expressions.
- `enforceForTSTypes` (default `false`): apply to TypeScript type literals and interfaces.

## Original Documentation

- ESLint rule: https://eslint.org/docs/latest/rules/accessor-pairs
