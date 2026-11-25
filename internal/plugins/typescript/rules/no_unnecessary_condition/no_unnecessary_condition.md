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

With `noUncheckedIndexedAccess` enabled, rslint preserves optional chaining on a
possibly missing indexed value, including a key variable with a literal type.
typescript-eslint 8.74.0 can suggest removing the last `?.` in these examples,
which introduces a TypeScript error.

```typescript
declare const items: { id: string }[] | undefined;
declare const records: Record<number, { id: string }> | undefined;
declare const key: 0;

items?.[(key)]?.id;
records?.[(key)]?.id;
```

The same protection applies when an index signature explicitly includes
`undefined` or `null`, or a property keyed by a unique symbol is optional. Index
matching follows TypeScript's rules for numeric keys, numeric strings, symbols,
and template literal keys.

rslint also reports constant conditions in array predicate callbacks and
truthiness assertions, unreachable logical fallbacks, and computed reads narrowed
to `null` or `undefined`. typescript-eslint 8.74.0 can miss these cases:

```typescript
declare const flags: boolean[];
declare const label: string | undefined;

flags.filter(flag => flag || true);
const display = label || 'Untitled' || 'Fallback';

function read(value: { x?: number }) {
  if (value.x === undefined) return value['x'] ?? 42;
  return value.x;
}
```

Async and generator callbacks return objects, regardless of the value in their
`return` statement. For example, `flags.filter(async () => false)` keeps every
element. rslint reports the callback as always truthy instead of reporting its
body as always falsy.

Logical assignments check the value being read, including control-flow narrowing.
For accessors, this is the getter's type; for a union of keys, it is the union of
the selected read types. Setter types and write intersections must not determine
whether the condition is necessary. For example, this `??=` remains necessary:

```typescript
interface Box {
  get value(): number | undefined;
  set value(reset: null);
}

function reset(box: Box) {
  box['value'] ??= null;
}
```

## Original Documentation

- [typescript-eslint: no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.74.0/packages/eslint-plugin/src/rules/no-unnecessary-condition.ts)
