# init-declarations

## Rule Details

Require or disallow initialization in variable declarations.

This rule applies to `var`, `let`, `const`, `using`, and `await using` declarations. It reports each declarator whose initialization does or does not match the configured mode.

Examples of **incorrect** code for this rule with `"always"` (the default):

```javascript
function foo() {
  var bar;
  let baz;
}
```

Examples of **correct** code for this rule with `"always"`:

```javascript
function foo() {
  var bar = 1;
  let baz = 2;
  const qux = 3;
}
```

Examples of **incorrect** code for this rule with `"never"`:

```json
{ "init-declarations": ["error", "never"] }
```

```javascript
function foo() {
  var bar = 1;
  let baz = 2;
  for (let i = 0; i < 1; i++) {}
}
```

Examples of **correct** code for this rule with `"never"`:

```javascript
function foo() {
  var bar;
  let baz;
  const buzz = 1;
}
```

`const`, `using`, and `await using` declarations always require an initializer, so `"never"` never flags them.

Examples of **correct** code for this rule with `{ "ignoreForLoopInit": true }`:

```json
{ "init-declarations": ["error", "never", { "ignoreForLoopInit": true }] }
```

```javascript
for (let i = 0; i < 1; i++) {}
```

A destructuring declarator (`var { a } = obj;`, `var [a] = arr;`) is never reported, in either mode.

`declare var`/`declare let`/`declare const`, and any declaration nested inside a `declare namespace`/`declare module`/ambient `.d.ts` context, is exempt in both modes.

## Options

This rule takes up to two options:

1. A string, either `"always"` (default) or `"never"`.
2. When the first option is `"never"`, an object with:
   - `ignoreForLoopInit`: when `true`, allows initialization in a `for` loop's declaration even under `"never"`. Default `false`.

## Original Documentation

- [ESLint: init-declarations](https://eslint.org/docs/latest/rules/init-declarations)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/init-declarations.js)
