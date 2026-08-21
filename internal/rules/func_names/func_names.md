# func-names

## Rule Details

A pattern that's becoming more common is to give function expressions names to aid in debugging. For example:

```javascript
Foo.prototype.bar = function bar() {};
```

Adding the second `bar` in the above example is optional. If you leave off the function name then when the function throws an exception you are likely to get something similar to `anonymous function` in the stack trace. If you provide the optional name for a function expression then you will get the name of the function expression in the stack trace.

This rule can enforce or disallow the use of named function expressions.

Examples of **incorrect** code for this rule with the default `"always"` option:

```javascript
Foo.prototype.bar = function () {};

const cat = {
  meow: function () {},
};

(function () {
  // ...
})();

export default function () {}
```

Examples of **correct** code for this rule with the default `"always"` option:

```javascript
Foo.prototype.bar = function bar() {};

const cat = {
  meow() {},
};

(function bar() {
  // ...
})();

export default function foo() {}
```

## Options

This rule has a string option:

- `"always"` (default) requires function expressions to have a name.
- `"as-needed"` requires function expressions to have a name, if the name isn't assigned automatically per the ECMAScript specification.
- `"never"` disallows named function expressions, except in recursive functions, where a name is needed.

This rule has an object option:

- `"generators": "always" | "as-needed" | "never"`
  - `"always"` require named generators.
  - `"as-needed"` require named generators if the name isn't assigned automatically per the ECMAScript specification.
  - `"never"` disallow named generators where possible.

When a value for `generators` is not provided the behavior for generator functions falls back to the base option.

Function expressions and function declarations in `export default` declarations must have a name under both `"always"` and `"as-needed"`.

### as-needed

ECMAScript 6 introduced a `name` property on all functions. The value of `name` is determined by evaluating the code around the function to see if a name can be inferred. For example, a function assigned to a variable will automatically have a `name` property equal to the name of the variable. The value of `name` is then used in stack traces for easier debugging.

Examples of **incorrect** code for this rule with the `"as-needed"` option:

```json
{ "func-names": ["error", "as-needed"] }
```

```javascript
Foo.prototype.bar = function () {};

(function () {
  // ...
})();

export default function () {}
```

Examples of **correct** code for this rule with the `"as-needed"` option:

```json
{ "func-names": ["error", "as-needed"] }
```

```javascript
const bar = function () {};

const cat = {
  meow: function () {},
};

class C {
  #bar = function () {};
  baz = function () {};
}

quux ??= function () {};

(function bar() {
  // ...
})();

export default function foo() {}
```

### never

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "func-names": ["error", "never"] }
```

```javascript
Foo.prototype.bar = function bar() {};

(function bar() {
  // ...
})();
```

Examples of **correct** code for this rule with the `"never"` option:

```json
{ "func-names": ["error", "never"] }
```

```javascript
Foo.prototype.bar = function () {};

(function () {
  // ...
})();
```

### generators

Examples of **incorrect** code for this rule with the `"always", { "generators": "as-needed" }` options:

```json
{ "func-names": ["error", "always", { "generators": "as-needed" }] }
```

```javascript
(function* () {
  // ...
})();
```

Examples of **correct** code for this rule with the `"always", { "generators": "as-needed" }` options:

```json
{ "func-names": ["error", "always", { "generators": "as-needed" }] }
```

```javascript
const foo = function* () {};
```

Examples of **incorrect** code for this rule with the `"always", { "generators": "never" }` options:

```json
{ "func-names": ["error", "always", { "generators": "never" }] }
```

```javascript
const foo = bar(function* baz() {});
```

Examples of **correct** code for this rule with the `"always", { "generators": "never" }` options:

```json
{ "func-names": ["error", "always", { "generators": "never" }] }
```

```javascript
const foo = bar(function* () {});
```

Examples of **incorrect** code for this rule with the `"as-needed", { "generators": "never" }` options:

```json
{ "func-names": ["error", "as-needed", { "generators": "never" }] }
```

```javascript
const foo = bar(function* baz() {});
```

Examples of **correct** code for this rule with the `"as-needed", { "generators": "never" }` options:

```json
{ "func-names": ["error", "as-needed", { "generators": "never" }] }
```

```javascript
const foo = bar(function* () {});
```

Examples of **incorrect** code for this rule with the `"never", { "generators": "always" }` options:

```json
{ "func-names": ["error", "never", { "generators": "always" }] }
```

```javascript
const foo = bar(function* () {});
```

Examples of **correct** code for this rule with the `"never", { "generators": "always" }` options:

```json
{ "func-names": ["error", "never", { "generators": "always" }] }
```

```javascript
const foo = bar(function* baz() {});
```

## Original Documentation

- [ESLint: func-names](https://eslint.org/docs/latest/rules/func-names)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/func-names.js)
