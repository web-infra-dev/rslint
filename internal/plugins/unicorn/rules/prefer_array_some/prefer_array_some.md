# prefer-array-some

## Rule Details

Prefer using [`Array#some(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/some) over:

- [`Array#find(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/find) / [`Array#findLast(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/findLast) when the result is only used as a boolean.
- [`Array#findIndex(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/findIndex) / [`Array#findLastIndex(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/findLastIndex) compared against `-1` / `0`.
- A non-zero length check on [`Array#filter(…)`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter).

`.some(…)` communicates the intent — "is there a match?" — more directly and can stop iterating at the first match.

Typed arrays carry the same methods and are checked too. Keyed collections
(`Map`, `Set`, `WeakMap`, `WeakSet`) are not: their `.find(…)` / `.filter(…)`
are unrelated APIs where the rewrite would not hold.

Examples of **incorrect** code for this rule:

```javascript
if (array.find(element => element === "🦄")) {
	// …
}

const hasUnicorn = array.findIndex(element => element === "🦄") !== -1;

const hasUnicorn = array.filter(element => element === "🦄").length > 0;

const foo = array.find(element => element === "🦄") ? bar : baz;
```

Examples of **correct** code for this rule:

```javascript
if (array.some(element => element === "🦄")) {
	// …
}

const hasUnicorn = array.some(element => element === "🦄");

// The result is used, not just as a boolean.
const unicorn = array.find(element => element === "🦄");

// The index is used.
const index = array.findIndex(element => element === "🦄");
```

## Original Documentation

- [eslint-plugin-unicorn / prefer-array-some](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/prefer-array-some.md)
