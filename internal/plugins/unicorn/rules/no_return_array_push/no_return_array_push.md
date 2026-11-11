# no-return-array-push

## Rule details

Disallow using the return value of `Array#push()` or `Array#unshift()` when the
call has at least one argument. Both methods return the updated array length.
Read `.length` after the mutation when you need that value.

The rule allows standalone calls and calls whose results are discarded with
`void`. It also allows member access on the result, as in
`router.push(to).catch(handleError)`, and skips receivers such as `Map` and
`Set` when it can identify their type. Common router and stream `push()` calls
are allowed unless the rule identifies the receiver as an array.

A direct `return` inside a block can offer an editor suggestion that moves the
call before a bare `return`.

Examples of **incorrect** code for this rule

```javascript
function add(item) {
  return items.push(item);
}

const length = items.unshift(item);
source.forEach(item => target.push(item));
```

Examples of **correct** code for this rule

```javascript
function add(item) {
  items.push(item);
  return;
}

items.unshift(item);
const length = items.length;
target.push(...source);
void items.push(item);
```

## Options

This rule has no options.

## Differences from upstream

Rslint can identify a non-array receiver from an explicit JSDoc cast in
JavaScript. In this example, Rslint allows the custom `push()` call. Unicorn
v79.0.0 can report it when run without type information.

```javascript
class Queue {
  push(value) {
    return value;
  }
}

const queue = /** @type {Queue} */ (readQueue());
const value = queue.push(1);
```

When ESLint is configured with TypeScript type information for JavaScript,
Unicorn can infer that `stream` is an array in the example below. Rslint
allows this call. Add `/** @returns {number[]} */` to `read` for Rslint to
identify the array and report the call.

```javascript
function read() {
  return [];
}

const stream = read();
const length = stream.push(1);
```

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-return-array-push.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-return-array-push.js)
