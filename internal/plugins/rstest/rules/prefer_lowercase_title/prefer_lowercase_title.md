# prefer-lowercase-title

Enforce lowercase test names.

Test and describe block titles should begin with a lowercase character. Uppercase initial letters are conventionally reserved for class names and named exports, not test descriptions.

## Options

This rule accepts an options object with the following properties:

### `ignore`

An array of strings specifying which functions to ignore. Valid values:

- `"describe"` — ignore all describe calls
- `"test"` — ignore all test calls
- `"it"` — ignore all it calls

### `allowedPrefixes`

An array of prefixes. A title that starts with any of these prefixes is ignored.

### `ignoreTopLevelDescribe`

A boolean, default `false`. When `true`, top-level describe blocks are not checked.

### `ignoreTodos`

A boolean, default `false`. When `true`, calls with a `.todo` member are ignored.

## Incorrect

```js
test('Foo', function () {});
describe('Bar', function () {});
```

## Correct

```js
test('foo', function () {});
describe('bar', function () {});
```

## References

- [eslint-plugin-jest source](https://github.com/jest-community/eslint-plugin-jest/blob/main/src/rules/prefer-lowercase-title.ts)
