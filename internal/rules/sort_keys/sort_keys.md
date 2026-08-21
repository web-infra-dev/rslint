# sort-keys

## Rule Details

This rule checks all property definitions of object literal expressions and verifies that all keys are sorted alphabetically.

Examples of **incorrect** code for this rule:

```javascript
var obj1 = { a: 1, c: 3, b: 2 };
var obj2 = { a: 1, "c": 3, b: 2 };

// Case-sensitive by default.
var obj3 = { a: 1, b: 2, C: 3 };

// Non-natural order by default.
var obj4 = { 1: a, 2: c, 10: b };

// This rule checks computed properties which have a simple name as well.
var obj5 = { a: 1, ["c"]: 3, b: 2 };
```

Examples of **correct** code for this rule:

```javascript
var obj1 = { a: 1, b: 2, c: 3 };
var obj2 = { a: 1, "b": 2, c: 3 };

// Case-sensitive by default.
var obj3 = { C: 3, a: 1, b: 2 };

// Non-natural order by default.
var obj4 = { 1: a, 10: b, 2: c };

// This rule ignores computed properties which have a non-simple name.
var obj5 = { a: 1, [c + d]: 3, b: 2 };

// This rule does not report unsorted properties that are separated by a spread property.
var obj6 = { b: 1, ...c, a: 2 };
```

Examples of **incorrect** code for this rule with `{ "caseSensitive": false }`:

```json
{ "sort-keys": ["error", "asc", { "caseSensitive": false }] }
```

```javascript
var obj = { a: 1, C: 3, b: 2 };
```

Examples of **incorrect** code for this rule with `{ "natural": true }`:

```json
{ "sort-keys": ["error", "asc", { "natural": true }] }
```

```javascript
var obj = { 1: a, 10: c, 2: b };
```

Examples of **incorrect** code for this rule with `{ "minKeys": 4 }`:

```json
{ "sort-keys": ["error", "asc", { "minKeys": 4 }] }
```

```javascript
var obj = { a: 1, c: 2, b: 3, d: 4 };
```

Examples of **correct** code for this rule with `{ "allowLineSeparatedGroups": true }`:

```json
{ "sort-keys": ["error", "asc", { "allowLineSeparatedGroups": true }] }
```

```javascript
var obj = {
    e: 1,
    f: 2,
    g: 3,

    a: 4,
    b: 5,
    c: 6,
};
```

Examples of **correct** code for this rule with `{ "ignoreComputedKeys": true }`:

```json
{ "sort-keys": ["error", "asc", { "ignoreComputedKeys": true }] }
```

```javascript
var obj = { a: 1, [c]: 2, b: 3 };
```

## Options

The 1st option is `"asc"` or `"desc"`.

- `"asc"` (default) enforces properties to be in ascending order.
- `"desc"` enforces properties to be in descending order.

The 2nd option is an object with the following properties.

- `caseSensitive` — if `true`, enforces properties to be in case-sensitive order. Default is `true`.
- `natural` — if `true`, enforces properties to be in natural order: strings that mix letters and numbers are compared the way a human would, so `10` sorts after `2` instead of before it. Default is `false`.
- `minKeys` — the minimum number of keys an object must have for its sort order to be checked. Default is `2`.
- `allowLineSeparatedGroups` — if `true`, a blank line after a property resets sorting: the properties that follow only need to be sorted relative to each other, not to the properties before the blank line. Default is `false`.
- `ignoreComputedKeys` — if `true`, computed keys are ignored entirely and reset the sorting of the keys that follow them. Default is `false`.

## Original Documentation

- [ESLint: sort-keys](https://eslint.org/docs/latest/rules/sort-keys)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/sort-keys.js)
