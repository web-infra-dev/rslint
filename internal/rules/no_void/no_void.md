# no-void

## Rule Details

Disallow use of the `void` operator.

Examples of **incorrect** code for this rule:

```javascript
void foo;
void someFunction();

const foo = void bar();
function baz() {
  return void 0;
}
```

## Options

### `allowAsStatement`

When `true`, permits `void` as a standalone statement but still prohibits its
use in expression contexts such as variable assignments or return statements.
Default: `false`.

Examples of **incorrect** code for this rule with `{ "allowAsStatement": true }`:

```json
{ "no-void": ["error", { "allowAsStatement": true }] }
```

```javascript
const foo = void bar();
function baz() {
  return void 0;
}
```

Examples of **correct** code for this rule with `{ "allowAsStatement": true }`:

```json
{ "no-void": ["error", { "allowAsStatement": true }] }
```

```javascript
void foo;
void someFunction();
```

## Original Documentation

- [ESLint: no-void](https://eslint.org/docs/latest/rules/no-void)
- [Source code](https://github.com/eslint/eslint/blob/main/lib/rules/no-void.js)
