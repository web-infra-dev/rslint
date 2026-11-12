# prefer-at

## Rule details

Prefer `.at()` to read an element or character from an array, typed array, or
string. Negative indices count from the end.

The rule checks length-based indices such as `items[items.length - 1]`,
one-element reads from `.slice()`, and calls to last-element helpers such as
`lodash.last()`. It also checks `charAt()` and one-character `substring()` calls.

Examples of **incorrect** code for this rule

```javascript
const last = items[items.length - 1];
const secondLast = items.slice(-2, -1)[0];
const lastItem = lodash.last(items);
const character = text.charAt(text.length - 1);
const first = text.substring(0, 1);
```

Examples of **correct** code for this rule

```javascript
const last = items.at(-1);
const secondLast = items.at(-2);
const lastItem = items.at(-1);
const character = text.at(-1);
const first = text.at(0);
```

The rule ignores indexed writes, updates, method calls, and tagged templates
because these operations need the original property access. It also ignores
indexed reads on `arguments` and recognized DOM collections, such as
`element.children` and the result of `document.querySelectorAll()`.

## Fixes and suggestions

The rule can autofix length-based reads, one-element `.slice()` patterns, and
last-element helper calls. It offers editor suggestions for string methods so
you can check how the replacement handles out-of-range indices. For example,
`''.charAt(0)` returns an empty string and `''.at(0)` returns `undefined`.

## Options

### checkAllIndexAccess

Defaults to `false`. Set it to `true` to check non-negative integer indices
that can be determined from the code, such as `items[0]`, and additional
`charAt()` calls. The rule still allows `charAt(-1)` because it returns an
empty string while `at(-1)` reads from the end.

### getLastElementFunctions

Defaults to `[]`. Add names or dotted paths such as `getLast` or
`utils.lastElement` for helpers that return the last element of their argument.
The rule always recognizes `_.last`, `lodash.last`, and `underscore.last`.

```javascript
export default [
  {
    plugins: ['unicorn'],
    rules: {
      'unicorn/prefer-at': [
        'error',
        {
          checkAllIndexAccess: true,
          getLastElementFunctions: ['getLast', 'utils.lastElement'],
        },
      ],
    },
  },
];
```

## Differences from upstream

Rslint withholds autofixes on obvious non-array values such as numbers and
plain objects. For example, an object literal can provide numeric keys and a
`length` while lacking an `.at()` method.
The rule reports the pattern and leaves the edit for you.

```javascript
const items = { 0: 'first', length: 1 };
const last = items[items.length - 1];
```

Rslint also ignores `arguments` in slice patterns and through TypeScript
assertions. A rest parameter gives you an array with an `.at()` method.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-at.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-at.js)
