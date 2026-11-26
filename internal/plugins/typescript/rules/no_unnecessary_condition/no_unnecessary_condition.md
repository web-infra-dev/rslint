# no-unnecessary-condition

## Rule Details

Disallow conditionals where the type is always truthy or always falsy.

Any expression being used as a condition must be able to evaluate as truthy or falsy in order to be considered necessary. Conversely, any expression that always evaluates to truthy or always evaluates to falsy is considered unnecessary and will be flagged by this rule.

Examples of **incorrect** code for this rule:

```typescript
function head<T>(items: T[]) {
  // items is always truthy (arrays are objects)
  if (items) {
    return items[0].toUpperCase();
  }
}

function foo(arg: 'bar' | 'baz') {
  // arg is always truthy (non-empty string literals)
  if (arg) {
  }
}
```

Examples of **correct** code for this rule:

```typescript
function head<T>(items: T[]) {
  // items.length can be zero
  if (items.length) {
    return items[0].toUpperCase();
  }
}

function foo(arg: string) {
  // string can be empty
  if (arg) {
  }
}
```

## Differences from upstream

- An `async` or generator callback passed to an array predicate method such as `filter` is judged by the Promise or generator object it returns, because that object is the value the method consumes. `flags.filter(async () => false)` reports `alwaysTruthyFunc` instead of upstream's `alwaysFalsy` on `false`, and `flags.filter(async flag => flag)` reports `alwaysTruthyFunc` where upstream reports nothing.
- When an accessor's getter can return values its setter does not accept, a read of that property uses the getter's declared type instead of the type TypeScript narrows to after an assignment. After `box.value = 1`, `box.value ?? 0` is not reported when `get value(): number | undefined` is paired with `set value(v: number)`, whereas upstream reports `neverNullish`. Other narrowing of such a property is not kept either, so `if (box.value !== undefined) box.value ?? 0;` is not reported. A value already copied into a variable, such as `const saved = box.value`, keeps TypeScript's narrowed type.

## Original Documentation

- [typescript-eslint: no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.74.0/packages/eslint-plugin/src/rules/no-unnecessary-condition.ts)
