# prefer-single-call

Combine adjacent calls to `Array#push()`, `Array#unshift()`,
`Element#classList.add()`, `Element#classList.remove()`, or `importScripts()`.

## Rule details

These methods accept multiple arguments. Calls must be adjacent expression
statements on the same receiver. For `unshift()`, the second call's arguments
come first so the resulting element order stays the same.

Examples of **incorrect** code:

```javascript
const values = [];
values.push(1);
values.push(2, 3);

values.unshift(1);
values.unshift(2, 3);

element.classList.add('active');
element.classList.add('visible');

importScripts('first.js');
importScripts('second.js');
```

Examples of **correct** code:

```javascript
const values = [];
values.push(1, 2, 3);
values.unshift(2, 3, 1);
element.classList.add('active', 'visible');
importScripts('first.js', 'second.js');
```

Known non-array receivers are ignored. Unknown array receivers receive a
suggestion; known arrays are automatically fixed only when every argument has
a static value without side effects. Arguments such as `values.length` and
spread arguments require a suggestion because merging changes when they run
relative to the first mutation. For class lists and `importScripts()`, a side
effect in the second call's arguments also requires a suggestion.

Comments in the text that would be removed suppress both fixes and suggestions.
The rule preserves optional-chain behavior and does not replace loops with
spread calls.

## Options

### `ignore`

An array of exact function or method paths to ignore. The default is `[]`.

```json
{
  "unicorn/prefer-single-call": [
    "error",
    { "ignore": ["readable.push", "element.classList.add", "importScripts"] }
  ]
}
```

Both `push` and `unshift` on `stream`, `this`, `this.stream`, `process.stdin`,
`process.stdout`, and `process.stderr` are always ignored.

## Differences from upstream

In TypeScript files, rslint uses inferred types to recognize arrays and ignore
custom objects with a `push` or `unshift` method. This matches upstream with
type information enabled. For example, rslint automatically fixes these calls,
while upstream without type information offers a suggestion:

```typescript
const values = [0].map(value => value);
values.push(1);
values.push(2);
```

When an argument contains an array spread, rslint offers a suggestion even if
the spread has a known value. Upstream may automatically fix the same calls:

```javascript
const values = [];
values.push([...[1]]);
values.push(2);
```

Both tools report these calls and can merge them into `values.push([...[1]], 2)`.

## Original documentation

- [eslint-plugin-unicorn: prefer-single-call](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/prefer-single-call.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/prefer-single-call.js)
