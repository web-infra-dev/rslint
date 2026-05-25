# object-shorthand

Require or disallow method and property shorthand syntax for object literals.

## Rule Details

ECMAScript 6 provides shorthand syntax for defining object literal methods and
properties. It is useful when an object property shares the name of a local
variable and lets you define methods using concise syntax. This rule enforces
usage of that shorthand whenever possible (the default `"always"` mode) and can
be configured to enforce the opposite (`"never"`) or to cover only one of the
two shorthand forms.

Examples of **incorrect** code for this rule (default):

```javascript
const foo = {
  w: function () {},
  x: function* () {},
  z: z,
};
```

Examples of **correct** code for this rule (default):

```javascript
const foo = {
  w() {},
  *x() {},
  z,
};
```

## Options

The first option is a string:

- `"always"` (default) — always use shorthand where possible.
- `"methods"` — enforce only method shorthand.
- `"properties"` — enforce only property shorthand.
- `"never"` — never use shorthand.
- `"consistent"` — properties within an object must be all shorthand or all
  longform.
- `"consistent-as-needed"` — all properties must be shorthand when they can be,
  otherwise all must be longform.

The second option is an object with additional flags (valid with `"always"`,
`"methods"`, or `"properties"` as noted on each):

- `avoidQuotes` — prefer longform for string literal keys (all modes above).
- `ignoreConstructors` — skip constructor-style names (only `"always"` or
  `"methods"`).
- `methodsIgnorePattern` — regular expression of method names to skip
  (only `"always"` or `"methods"`).
- `avoidExplicitReturnArrows` — report arrow functions with block bodies
  (`() => { … }`) as preferring the method shorthand
  (only `"always"` or `"methods"`). Arrows that reference `this`, `super`,
  `arguments`, or `new.target` are left alone because the conversion would
  change their binding.

## Original Documentation

- ESLint rule: https://eslint.org/docs/latest/rules/object-shorthand
- Source code: https://github.com/eslint/eslint/blob/main/lib/rules/object-shorthand.js
