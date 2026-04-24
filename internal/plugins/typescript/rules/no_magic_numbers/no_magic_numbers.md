# no-magic-numbers

## Rule Details

Disallow magic numbers. A "magic number" is a numeric literal that is used in the code without explanation or assignment to a named constant. Magic numbers make code less readable and harder to maintain.

Examples of **incorrect** code for this rule:

```javascript
var total = 500;
if (foo === 10) {}
var data = ['foo', 'bar', 'baz'];
var dataLast = data[2];
```

Examples of **correct** code for this rule:

```javascript
var TAX = 0.25;
var total = 500;
if (foo === TAX) {}
const data = ['foo', 'bar', 'baz'];
const LAST = 2;
var dataLast = data[LAST];
```

Examples of **correct** code for this rule with `{ "ignoreEnums": true }`:

```json
{ "@typescript-eslint/no-magic-numbers": ["error", { "ignoreEnums": true }] }
```

```javascript
enum foo {
  SECOND = 1000,
  NEG = -1,
}
```

Examples of **correct** code for this rule with `{ "ignoreNumericLiteralTypes": true }`:

```json
{ "@typescript-eslint/no-magic-numbers": ["error", { "ignoreNumericLiteralTypes": true }] }
```

```javascript
type Foo = 1;
type Foo = 1 | 2 | 3;
```

Examples of **correct** code for this rule with `{ "ignoreReadonlyClassProperties": true }`:

```json
{ "@typescript-eslint/no-magic-numbers": ["error", { "ignoreReadonlyClassProperties": true }] }
```

```javascript
class Foo {
  readonly A = 1;
  readonly B = 2;
  public static readonly C = 1;
}
```

Examples of **correct** code for this rule with `{ "ignoreTypeIndexes": true }`:

```json
{ "@typescript-eslint/no-magic-numbers": ["error", { "ignoreTypeIndexes": true }] }
```

```javascript
type Foo = Bar[0];
type Foo = Bar[1 | -2];
type Foo = Parameters<Bar>[2];
```

## Original Documentation

[ESLint - no-magic-numbers](https://eslint.org/docs/latest/rules/no-magic-numbers)
[typescript-eslint - no-magic-numbers](https://typescript-eslint.io/rules/no-magic-numbers)
