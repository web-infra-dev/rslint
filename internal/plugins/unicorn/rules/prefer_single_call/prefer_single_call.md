# prefer-single-call

Combine adjacent calls to `Array#push()`, `Array#unshift()`,
`Element#classList.add()`, `Element#classList.remove()`, or `importScripts()`.

This rule is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Rule Details

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
direct spreads such as `values.push(...items)` require a suggestion because
merging changes when arguments run relative to the first mutation.
For class lists and `importScripts()`, a side effect in the second call's
arguments also requires a suggestion.

Comments in the text that would be removed suppress both fixes and suggestions.
The rule preserves optional-chain behavior and does not replace loops with
spread calls.

TypeScript's inferred receiver types also identify arrays. With static,
side-effect-free arguments, these calls are automatically fixed:

```typescript
const values = [0].map(value => value);
values.push(1);
values.push(2);
```

Static spreads inside array and object arguments are supported too. For example,
`values.push([...[1]]); values.push(2);` is automatically merged into
`values.push([...[1]], 2);`.

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

rslint offers a suggestion instead of an automatic fix when a spread source is
modified through a local alias or an operation such as `Object.defineProperty()`.
This includes aliases created by destructuring and references copied into
containers by `Object.assign()`. Visible changes to `Array.prototype` or
`String.prototype` also make affected array or string spreads require a suggestion.
For example, this custom iterator changes `values` while the second argument is
evaluated:

```javascript
const values = [];
const source = [1];
const alias = source;
alias[Symbol.iterator] = function* () {
  values.push(0);
  yield 1;
};
values.push(1);
values.push([...source]);
```

Automatically merging these calls would move the iterator's `push(0)` before
`push(1)` and change the result. Review the suggestion before applying it.

Arguments that apply decorators or call custom template tags also require a
suggestion, because those operations can run user code while evaluating the
arguments.

## Original Documentation

- [eslint-plugin-unicorn: prefer-single-call](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/prefer-single-call.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/prefer-single-call.js)
