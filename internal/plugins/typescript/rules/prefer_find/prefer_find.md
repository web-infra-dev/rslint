# prefer-find

Enforce the use of `Array.prototype.find()` over `Array.prototype.filter()` followed by `[0]` when looking for a single result.

## Rule Details

When searching for the first item in an array matching a condition, it may be tempting to use code like `arr.filter(x => x > 0)[0]`. However, it is simpler to use [`Array.prototype.find()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/find) instead, `arr.find(x => x > 0)`, which also returns the first entry matching a condition. Because `.find()` only executes the callback until it finds a match, it is also more efficient.

The rule triggers on these patterns when the receiver's type is an array or tuple:

- `arr.filter(p)[0]`
- `arr.filter(p).at(0)`
- `arr.filter(p)['0']` and `` arr.filter(p)[`0`] ``
- `arr['filter'](p)[0]` and `` arr[`filter`](p)[0] ``
- Sequence and ternary wrappers: `(a, b, arr.filter(p))[0]`, `(cond ? arr1.filter(p) : arr2.filter(p))[0]`
- Optional-chain receivers: `arr?.filter(p)[0]`

It does not trigger when:

- The receiver type is not an array or tuple (e.g. a custom `Filter` interface, `null`, `undefined`).
- The subscript or `.at(...)` argument does not statically resolve to zero.
- The `[0]` access is itself optional (`?.[0]`), the `.at(0)` callee is optional (`?.at(0)`), or the `.filter(...)` call is optional (`.filter?.(...)`) — rewriting these would change short-circuiting semantics.

The rewrite is offered as a **suggestion** that you must explicitly apply — it is not auto-applied. `.find()` stops at the first match, but `.filter()` always visits every element, so if your `.filter()` callback has side effects, applying the suggestion will change behavior.

Examples of **incorrect** code for this rule:

```typescript
declare const arr: string[];
arr.filter(item => item === 'aha')[0];
```

```typescript
declare const arr: string[];
arr.filter(item => item === 'aha').at(0);
```

```typescript
declare const arr: string[];
arr.filter(item => item === 'aha')['0'];
```

```typescript
declare const arr: string[];
const zero = 0;
arr.filter(item => item === 'aha').at(zero);
```

```typescript
declare const arr: { a: 1 }[] & { b: 2 }[];
arr.filter(f)[0];
```

Examples of **correct** code for this rule:

```typescript
[1, 2, 3].find(x => x > 1);
```

```typescript
declare const arr: string[];
arr.filter(item => item === 'aha')[1];
```

```typescript
declare const arr: string[];
arr.filter(item => item === 'aha').at(1);
```

```typescript
[].filter(() => true)?.[0];
```

```typescript
[].filter(() => true)?.at?.(0);
```

```typescript
[].filter?.(() => true)[0];
```

```typescript
interface Filter<T> {
  filter(predicate: (item: T) => boolean): Filter<T>;
}
declare const f: Filter<string>;
f.filter(x => x.length > 0)[0];
```

## When Not To Use It

If you intentionally use patterns like `.filter(callback)[0]` to execute side effects in `callback` on all array elements, you will want to avoid this rule.

## Original Documentation

- [typescript-eslint prefer-find](https://typescript-eslint.io/rules/prefer-find)
