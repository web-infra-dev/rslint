# max-nested-calls

Limit the depth of nested calls and constructor calls passed as arguments to other calls.

Deep call nesting is harder to read and can usually be clarified by extracting an intermediate result.

```js
// Incorrect
foo(bar(baz(qux())));

// Correct
const value = baz(qux());
foo(bar(value));
```

Fluent receiver chains such as `query().filter().map().toArray()` do not increase the nesting depth. Function, class, JSX, and static-block boundaries reset the depth.

## Options

### `max`

Type: `integer`  
Default: `3`

Sets the maximum allowed nested call depth.

```js
{
  'unicorn/max-nested-calls': ['error', { max: 4 }]
}
```

This port follows eslint-plugin-unicorn v75.0.0.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/max-nested-calls.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/max-nested-calls.js)
