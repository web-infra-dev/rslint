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

## Original Documentation

https://typescript-eslint.io/rules/no-unnecessary-condition
