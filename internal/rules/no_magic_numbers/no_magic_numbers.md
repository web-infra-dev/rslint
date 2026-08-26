# no-magic-numbers

## Rule Details

"Magic numbers" are numbers that occur multiple times in code without an explicit meaning. They should preferably be replaced by named constants.

The `no-magic-numbers` rule aims to make code more readable and refactoring easier by ensuring that special numbers are declared as constants to make their meaning explicit.

Examples of **incorrect** code for this rule:

```javascript
const dutyFreePrice = 100,
  finalPrice = dutyFreePrice + dutyFreePrice * 0.25;

const data = ['foo', 'bar', 'baz'];
const dataLast = data[2];

let SECONDS;
SECONDS = 60;
```

Examples of **correct** code for this rule:

```javascript
const TAX = 0.25;

const dutyFreePrice = 100,
  finalPrice = dutyFreePrice + dutyFreePrice * TAX;
```

## Options

This rule accepts an object with the following properties:

- `ignore` (default `[]`): an array of numbers to ignore. Values may be `number` or a `string` parsed as a `bigint` literal (e.g. `"100n"`).
- `ignoreArrayIndexes` (default `false`): whether numbers used as array indexes are considered okay.
- `ignoreDefaultValues` (default `false`): whether numbers used in default value assignments are considered okay.
- `ignoreClassFieldInitialValues` (default `false`): whether numbers used as initial values of class fields are considered okay.
- `enforceConst` (default `false`): whether to check for the `const` keyword in variable declarations of numbers.
- `detectObjects` (default `false`): whether to detect numbers when setting object properties.
- `ignoreEnums` (default `false`, TypeScript only): whether numbers used in enum members are considered okay.
- `ignoreNumericLiteralTypes` (default `false`, TypeScript only): whether numbers used in numeric literal types are considered okay.
- `ignoreReadonlyClassProperties` (default `false`, TypeScript only): whether numbers used in `readonly` class properties are considered okay.
- `ignoreTypeIndexes` (default `false`, TypeScript only): whether numbers used to index types are considered okay.

### `ignore`

Examples of **correct** code for this rule with `{ "ignore": [1] }`:

```json
{ "no-magic-numbers": ["error", { "ignore": [1] }] }
```

```javascript
const data = ['foo', 'bar', 'baz'];
const dataLast = data.length && data[data.length - 1];
```

Examples of **correct** code for this rule with `{ "ignore": ["1n"] }`:

```json
{ "no-magic-numbers": ["error", { "ignore": ["1n"] }] }
```

```javascript
foo(1n);
```

### `ignoreArrayIndexes`

This option allows only valid array indexes: numbers that will be coerced to one of `"0"`, `"1"`, `"2"` ... `"4294967294"`.

Examples of **correct** code for this rule with `{ "ignoreArrayIndexes": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreArrayIndexes": true }] }
```

```javascript
const item = data[2];
data[100] = a;
f(data[0]);
a = data[-0]; // same as data[0], -0 will be coerced to "0"
a = data[10n]; // same as data[10], 10n will be coerced to "10"
a = data[4294967294]; // max array index
```

Examples of **incorrect** code for this rule with `{ "ignoreArrayIndexes": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreArrayIndexes": true }] }
```

```javascript
f(2); // not used as array index
a = data[-1];
a = data[2.5];
a = data[4294967295]; // above the max array index
```

### `ignoreDefaultValues`

Examples of **correct** code for this rule with `{ "ignoreDefaultValues": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreDefaultValues": true }] }
```

```javascript
const { tax = 0.25 } = accountancy;

function mapParallel(concurrency = 3) {}
```

### `ignoreClassFieldInitialValues`

Examples of **correct** code for this rule with `{ "ignoreClassFieldInitialValues": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreClassFieldInitialValues": true }] }
```

```javascript
class C {
  foo = 2;
  bar = -3;
  #baz = 4;
  static qux = 5;
}
```

Examples of **incorrect** code for this rule with `{ "ignoreClassFieldInitialValues": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreClassFieldInitialValues": true }] }
```

```javascript
class C {
  foo = 2 + 3;
}
```

### `enforceConst`

Examples of **incorrect** code for this rule with `{ "enforceConst": true }`:

```json
{ "no-magic-numbers": ["error", { "enforceConst": true }] }
```

```javascript
let TAX = 0.25;
```

### `detectObjects`

Examples of **incorrect** code for this rule with `{ "detectObjects": true }`:

```json
{ "no-magic-numbers": ["error", { "detectObjects": true }] }
```

```javascript
const magic = {
  tax: 0.25,
};
```

### `ignoreEnums`

Examples of **correct** TypeScript code for this rule with `{ "ignoreEnums": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreEnums": true }] }
```

```typescript
enum foo {
  SECOND = 1000,
}
```

### `ignoreNumericLiteralTypes`

Examples of **correct** TypeScript code for this rule with `{ "ignoreNumericLiteralTypes": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreNumericLiteralTypes": true }] }
```

```typescript
type Foo = 1 | 2 | 3;
```

### `ignoreReadonlyClassProperties`

Examples of **correct** TypeScript code for this rule with `{ "ignoreReadonlyClassProperties": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreReadonlyClassProperties": true }] }
```

```typescript
class Foo {
  readonly A = 1;
  public static readonly B = 2;
}
```

### `ignoreTypeIndexes`

Examples of **correct** TypeScript code for this rule with `{ "ignoreTypeIndexes": true }`:

```json
{ "no-magic-numbers": ["error", { "ignoreTypeIndexes": true }] }
```

```typescript
type Foo = Bar[0];
type Baz = Parameters<Foo>[2];
```

## Original Documentation

- [ESLint: no-magic-numbers](https://eslint.org/docs/latest/rules/no-magic-numbers)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-magic-numbers.js)
