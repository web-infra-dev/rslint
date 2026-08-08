# valid-title

## Rule Details

Enforce valid titles on Rstest `describe`, `test`, and `it` blocks. Titles should be informative strings, follow project conventions you configure, and use only the `printf`-style placeholders that Rstest actually expands in array-based `.each` / `.for` titles.

This rule checks that titles are:

- not empty (including empty template literals),
- string literals (unless relaxed via options),
- not accidentally prefixed with the block keyword (for example `test('test foo')`),
- free of leading or trailing whitespace (unless `ignoreSpaces` is enabled),
- using only valid `printf` specifiers for array-based `.each` / `.for` table names,
- not matching `disallowedWords` (whole-word, case-insensitive) when that option is set,
- satisfying `mustMatch` / `mustNotMatch` patterns when configured (per `describe` / `test` / `it` or a single global pattern).

Auto-fix is available for accidental surrounding spaces and duplicate keyword prefixes.

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

Titles of array-based `.each` and `.for` registrations are formatted at runtime. After a single `%`, only `s`, `d`, `j`, `i`, `f`, `o`, `O`, `c`, `#`, and `$` are accepted, and `%%` denotes a literal percent.

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

Tagged-template tables are not checked, because they interpolate with `$name` rather than `printf` specifiers and a bare `%` in them is literal text:

```js
test.each`
  a    | b
  ${1} | ${2}
`('returns $a and $b', ({ a, b }) => {});
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

<!-- cspell:ignore sdjifo -->

### Differences from `jest/valid-title`

The rule is otherwise a port of `jest/valid-title`, with three differences that follow from how Rstest itself behaves.

**The set of valid `printf` specifiers is Rstest's, not Jest's.** Rstest formats parameterized titles with Node's `util.format` semantics (`formatRegExp` is `/%[sdjifoOc%]/`), while Jest uses its own `pretty-format` placeholder set. So:

| Title | `jest/valid-title` | `rstest/valid-title` |
| --- | :---: | :---: |
| `test.each([])('%p', fn)` | valid | **reported** — Rstest leaves `%p` in the title verbatim |
| `test.each([])('%O', fn)` | reported | **valid** — Rstest expands it |
| `test.each([])('%c', fn)` | reported | **valid** — Rstest expands it |

**`.for` titles are checked too.** Rstest has both `.each` and `.for`, and both format their titles the same way. Jest has no `.for`.

**There are no `f` / `x` prefixed aliases.** Rstest provides no `fit`, `xit`, `xtest`, `fdescribe`, or `xdescribe`, so the keyword compared against the title is always the API being registered — `test`, `it`, or `describe` — even when it was imported or aliased under another name. `import { it as xit } from '@rstest/core'; xit('it works', fn)` is reported, and `mustMatch: { test: … }` applies to `import { test as check } …; check(…)`.

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

## Original Documentation

- [jest/valid-title](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/valid-title.md)
