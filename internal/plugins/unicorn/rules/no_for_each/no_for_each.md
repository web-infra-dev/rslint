# no-for-each

Prefer `for…of` over the `forEach` method.

Using a `for…of` statement can improve readability, allow early exits with
`break` or `return`, and allow iterations to be skipped with `continue`.
TypeScript control-flow narrowing also remains in the surrounding function
instead of crossing a callback boundary.

A receiver known not to be an array, such as a `Map` or `Set`, is ignored. A
typed array is reported because `for…of` iterates it in the same way as
`TypedArray#forEach()`, although the rule only offers an edit when the receiver
is definitely an `Array`, `ReadonlyArray`, or tuple.

## Examples

Incorrect:

```js
array.forEach(element => {
	bar(element);
});
```

Correct:

```js
for (const element of array) {
	bar(element);
}
```

Optional arrays are guarded before iteration:

```js
if (array) {
	for (const element of array) {
		bar(element);
	}
}
```

The callback index becomes the first value from `entries()`:

```js
for (const [index, element] of array.entries()) {
	bar(element, index);
}
```

## Further reading

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-for-each.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-for-each.js)
