# prefer-set-has

## Rule Details

Prefer `Set#has()` over `Array#includes()` when checking for existence or
non-existence.

`Set#has()` is faster than `Array#includes()`. When a `const` array binding is
used only for existence checks (and, optionally, a few other supported
references), this rule recommends declaring it as a `Set` instead.

Examples of **incorrect** code for this rule:

```javascript
const array = [1, 2, 3];
const hasValue = value => array.includes(value);
```

Examples of **correct** code for this rule:

```javascript
const set = new Set([1, 2, 3]);
const hasValue = value => set.has(value);
```

An array with supported extra references can also be converted when it has more
than one `includes()` lookup. The array must be a plain literal with only
unique, statically known primitive or `null` values, and no holes, spreads, or
`-0`. Supported extra references are `for…of`, array spread, call or
constructor argument spread, `.length`, and `.forEach()` with a one-parameter
arrow function.

```javascript
// Incorrect
const array = [1, 2, 3];
for (const item of array) {
	console.log(item);
}

const length = array.length;
const hasValue = value => array.includes(value);
```

```javascript
// Correct
const set = new Set([1, 2, 3]);
for (const item of set) {
	console.log(item);
}

const length = set.size;
const hasValue = value => set.has(value);
```

An array that is only checked once is left alone:

```javascript
// Correct
const array = [1, 2, 3];
const hasOne = array.includes(1);
```

## Options

Type: `object`

### `minimumItems`

Type: `integer`\
Default: `0`

The minimum known array size before `Set#has()` is enforced. When this option
is greater than `0`, the rule only reports arrays with a statically known size.

```json
{ "unicorn/prefer-set-has": ["error", { "minimumItems": 5 }] }
```

Examples of **incorrect** code for this rule with `{ "minimumItems": 5 }`:

```javascript
const array = [1, 2, 3, 4, 5];
const hasValue = value => array.includes(value);
```

Examples of **correct** code for this rule with `{ "minimumItems": 5 }`:

```javascript
const array = [1, 2, 3, 4];
const hasValue = value => array.includes(value);
```

## Original Documentation

- [`unicorn/prefer-set-has`](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/prefer-set-has.md)
