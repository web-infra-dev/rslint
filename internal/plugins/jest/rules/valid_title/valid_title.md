# valid-title

## Rule Details

Enforce valid titles on Jest `describe`, `test`, and `it` blocks. Titles should be informative strings, follow project conventions you configure, and use only allowed `printf`-style placeholders in [array-based `.each`](https://jestjs.io/docs/api#testeachtable-name-fn-timeout) titles (see [Node `util.format`](https://nodejs.org/api/util.html#utilformatformat-args)).

This rule checks that titles are:

- not empty (including empty template literals where applicable),
- string literals (unless relaxed via options),
- not accidentally prefixed with the block keyword (for example `test('test foo')`),
- free of leading or trailing whitespace (unless `ignoreSpaces` is enabled),
- using only valid `printf` specifiers for `.each` table names,
- not matching `disallowedWords` (whole-word, case-insensitive) when that option is set,
- satisfying `mustMatch` / `mustNotMatch` patterns when configured (per `describe` / `test` / `it` or a single global pattern).

Auto-fix is available for some violations (for example accidental surrounding spaces and duplicate keyword prefixes), where the implementation can rewrite the title safely.

**`emptyTitle`**

Examples of **incorrect** code:

```js
describe('', () => {});
it('', () => {});
test('', () => {});
```

Examples of **correct** code:

```js
describe('suite', () => {});
it('does something', () => {});
test('works', () => {});
```

**`titleMustBeString`**

Use string literals for titles unless `ignoreTypeOfDescribeName` or `ignoreTypeOfTestName` is `true`.

Examples of **incorrect** code:

```js
it(123, () => {});
describe(myFunction, () => {});
```

Examples of **correct** code:

```js
it('is a string', () => {});
describe('suite', () => {});
```

**`invalidEachSpecifier`**

Array-based `.each` titles may use `printf`-style segments. After a single `%`, only `p`, `s`, `d`, `i`, `f`, `j`, `o`, `#`, and `$` are accepted, and `%%` denotes a literal percent (aligned with eslint-plugin-jest; see the Jest and Node documentation linked above).

Examples of **incorrect** code:

```js
test.each([[1, 2]])('.add(%I, %I)', (a, b) => {
  expect(a + b).toBe(3);
});
```

Examples of **correct** code:

```js
test.each([[1, 2]])('.add(%i, %i)', (a, b) => {
  expect(a + b).toBe(3);
});
```

**`duplicatePrefix`**

Examples of **incorrect** code:

```js
test('test foo', () => {});
it('it foo', () => {});
describe('describe foo', () => {
  it('bar', () => {});
});
```

Examples of **correct** code:

```js
test('foo', () => {});
it('foo', () => {});
describe('foo', () => {
  it('bar', () => {});
});
```

**`accidentalSpace`**

Examples of **incorrect** code:

```js
test(' foo', () => {});
it('foo ', () => {});
```

Examples of **correct** code:

```js
test('foo', () => {});
it('foo', () => {});
```

### Options

```ts
interface Options {
  ignoreSpaces?: boolean;
  ignoreTypeOfDescribeName?: boolean;
  ignoreTypeOfTestName?: boolean;
  disallowedWords?: string[];
  mustNotMatch?: Partial<Record<'describe' | 'test' | 'it', string>> | string;
  mustMatch?: Partial<Record<'describe' | 'test' | 'it', string>> | string;
}
```

- **`ignoreSpaces`** (default `false`): skip leading/trailing space checks.
- **`ignoreTypeOfDescribeName`** / **`ignoreTypeOfTestName`** (default `false`): allow non-string first arguments for `describe` or `test`/`it` respectively.
- **`disallowedWords`**: list of words that must not appear as whole words in titles (case-insensitive).
- **`mustMatch`** / **`mustNotMatch`**: ECMAScript regular expressions as strings, either one pattern for all block kinds or an object keyed by `describe`, `test`, and `it`. You can pass a two-element array `[pattern, customMessage]` to surface `*Custom` message variants.

For full option examples and edge cases, see the upstream rule documentation below.

## Original Documentation

- [jest/valid-title](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/valid-title.md)
