# prefer-number-properties

## Rule Details

Prefer `Number` static properties and methods over global number-related
properties and functions.

Examples of **incorrect** code for this rule:

```javascript
const value = parseInt("10", 2);
const number = parseFloat("10.5");
const finite = isFinite(value);
const missing = Number.isNaN(NaN);
```

Examples of **correct** code for this rule:

```javascript
const value = Number.parseInt("10", 2);
const number = Number.parseFloat("10.5");
const finite = Number.isFinite(value);
const missing = Number.isNaN(Number.NaN);
```

## Options

### checkInfinity

Type: `boolean`\
Default: `false`

When `true`, the rule also checks global `Infinity` and `-Infinity`.

```json
{ "unicorn/prefer-number-properties": ["error", { "checkInfinity": true }] }
```

Examples of **incorrect** code for this option:

```javascript
const upperBound = Infinity;
const lowerBound = -Infinity;
```

Examples of **correct** code for this option:

```javascript
const upperBound = Number.POSITIVE_INFINITY;
const lowerBound = Number.NEGATIVE_INFINITY;
```

### checkNaN

Type: `boolean`\
Default: `true`

When `false`, the rule does not check global `NaN`.

```json
{ "unicorn/prefer-number-properties": ["error", { "checkNaN": false }] }
```

Examples of **correct** code for this option:

```javascript
const missing = NaN;
```

## Original Documentation

- [eslint-plugin-unicorn: prefer-number-properties](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-number-properties.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/rules/prefer-number-properties.js)
