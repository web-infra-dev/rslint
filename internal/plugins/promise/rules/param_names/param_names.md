# param-names

Enforce consistent param names and ordering when creating new promises.

## Rule Details

Ensures that `new Promise()` is instantiated with the parameter names
`resolve, reject` to avoid confusion with order such as `reject, resolve`.
The Promise constructor uses the
[RevealingConstructor pattern](https://blog.domenic.me/the-revealing-constructor-pattern/).
Using the same parameter names as the language specification makes code more
uniform and easier to understand.

Examples of **incorrect** code for this rule:

```javascript
new Promise(function (reject, resolve) {}); // incorrect order
new Promise(function (ok, fail) {}); // non-standard parameter names
new Promise(function (_, reject) {}); // a simple underscore is not allowed
```

Examples of **correct** code for this rule:

```javascript
new Promise(function (resolve) {});
new Promise(function (resolve, reject) {});
new Promise(function (_resolve, _reject) {}); // unused-marker underscore prefix is allowed
```

## Options

### `resolvePattern`

Pass `{ resolvePattern: "^_?resolve$" }` to customize the first argument name
pattern. Default is `"^_?resolve$"`.

```json
{ "promise/param-names": ["error", { "resolvePattern": "^yes$" }] }
```

```javascript
new Promise(function (yes, reject) {});
```

### `rejectPattern`

Pass `{ rejectPattern: "^_?reject$" }` to customize the second argument name
pattern. Default is `"^_?reject$"`.

```json
{ "promise/param-names": ["error", { "rejectPattern": "^no$" }] }
```

```javascript
new Promise(function (resolve, no) {});
```

## Original Documentation

- [eslint-plugin-promise: param-names](https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/param-names.md)
