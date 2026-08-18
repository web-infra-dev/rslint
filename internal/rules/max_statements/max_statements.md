# max-statements

## Rule Details

This rule enforces a maximum number of statements allowed in function blocks.

Examples of **incorrect** code for this rule with the default `{ "max": 10 }` option:

```javascript
function foo() {
  const foo1 = 1;
  const foo2 = 2;
  const foo3 = 3;
  const foo4 = 4;
  const foo5 = 5;
  const foo6 = 6;
  const foo7 = 7;
  const foo8 = 8;
  const foo9 = 9;
  const foo10 = 10;

  const foo11 = 11; // Too many.
}
```

Examples of **correct** code for this rule with the default `{ "max": 10 }` option:

```javascript
function foo() {
  const foo1 = 1;
  const foo2 = 2;
  const foo3 = 3;
  const foo4 = 4;
  const foo5 = 5;
  const foo6 = 6;
  const foo7 = 7;
  const foo8 = 8;
  const foo9 = 9;
  return function () {
    // 10

    // The number of statements in the inner function does not count toward the
    // statement maximum.

    let bar;
    let baz;
    return 42;
  };
}
```

Note that this rule does not apply to class static blocks, and that statements in class static blocks do not count as statements in the enclosing function.

Examples of **correct** code for this rule with `{ "max": 2 }` option:

```json
{ "max-statements": ["error", 2] }
```

```javascript
function foo() {
  let one;
  let two = class {
    static {
      let three;
      let four;
      let five;
      if (six) {
        let seven;
        let eight;
        let nine;
      }
    }
  };
}
```

Examples of additional **correct** code for this rule with the `{ "max": 10 }, { "ignoreTopLevelFunctions": true }` options:

```json
{ "max-statements": ["error", 10, { "ignoreTopLevelFunctions": true }] }
```

```javascript
function foo() {
  const foo1 = 1;
  const foo2 = 2;
  const foo3 = 3;
  const foo4 = 4;
  const foo5 = 5;
  const foo6 = 6;
  const foo7 = 7;
  const foo8 = 8;
  const foo9 = 9;
  const foo10 = 10;
  const foo11 = 11;
}
```

`ignoreTopLevelFunctions` only ignores a function's statement count when it is the single top-level function in the file — for example, a module wrapped entirely in one IIFE. When more than one top-level function exists, each one is still checked individually.

## Original Documentation

- [ESLint: max-statements](https://eslint.org/docs/latest/rules/max-statements)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/max-statements.js)
