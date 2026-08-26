# no-array-constructor

## Rule Details

Use of the `Array` constructor to construct a new array is generally discouraged in favor of array literal notation because of the single-argument pitfall and because the `Array` global may be redefined. The exception is when the `Array` constructor is used to intentionally create sparse arrays of a specified size by giving the constructor a single numeric argument.

This rule disallows `Array` constructors.

Examples of **incorrect** code for this rule:

```javascript
Array();

Array(0, 1, 2);

new Array(0, 1, 2);

Array(...args);
```

Examples of **correct** code for this rule:

```javascript
Array(500);

new Array(someOtherArray.length);

[0, 1, 2];

const createArray = Array => new Array();
```

This rule additionally supports TypeScript type syntax.

Examples of **correct** code for this rule:

```typescript
new Array<number>(1, 2, 3);

new Array<Foo>();

Array<number>(1, 2, 3);

Array<Foo>();

Array?.foo();
```

Examples of **incorrect** code for this rule:

```typescript
new Array();

new Array(0, 1, 2);

Array?.(x, y);

Array?.(0, 1, 2);
```

## Differences from ESLint

- When the array constructor call is fixed onto a new line right after certain TypeScript-only constructs — a type alias (`type T = Foo`), an ambient or overload function declaration (`declare function foo()`), an import-equals declaration (`import Foo = Bar`), or an `as`/`satisfies` type cast — rslint's autofix inserts a leading `;` that ESLint omits (e.g. `type T = Foo\n;[0, 1]` instead of `type T = Foo\n[0, 1]`). The extra semicolon never changes the resulting code's behavior.
- In TypeScript files, rslint treats `Array` as a predefined library variable even when `parserOptions.lib` is explicitly empty. For example, with `parserOptions.lib: []` and `globals: { Array: "off" }`, rslint still reports and fixes `Array()`, while typescript-eslint does not. Native rule contexts do not currently expose parser-specific library selection.

## Original Documentation

- [ESLint: no-array-constructor](https://eslint.org/docs/latest/rules/no-array-constructor)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-array-constructor.js)
