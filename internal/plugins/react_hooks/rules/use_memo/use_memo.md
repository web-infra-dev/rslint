# react-hooks/use-memo

## Rule Details

Validates usage of the `useMemo()` hook against common mistakes.

`useMemo()` should receive an inline function that computes the cached value and an optional dependency array that can be checked statically. The callback should be synchronous, take no parameters, and should not reassign variables declared outside of the callback.

Examples of **incorrect** code for this rule:

```javascript
function Component({ value }) {
  const doubled = useMemo(async () => value * 2, [value]);
  return <div>{doubled}</div>;
}
```

```javascript
function Component({ value, deps }) {
  const doubled = useMemo(() => value * 2, deps);
  return <div>{doubled}</div>;
}
```

```javascript
function Component() {
  let value;
  const memoized = useMemo(() => {
    value = [];
    return value;
  }, []);
  return <div>{memoized}</div>;
}
```

Examples of **correct** code for this rule:

```javascript
function Component({ value }) {
  const doubled = useMemo(() => value * 2, [value]);
  return <div>{doubled}</div>;
}
```

This rule follows `eslint-plugin-react-hooks@7.x`: checking that a `useMemo()` callback returns a value belongs to the separate upstream `react-hooks/void-use-memo` rule. Like upstream, this rule tracks direct `useMemo()` and `React.useMemo()` call shapes.

## Original Documentation

- [react.dev - use-memo](https://react.dev/reference/eslint-plugin-react-hooks/lints/use-memo)
- [Source code - ValidateUseMemo.ts](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Validation/ValidateUseMemo.ts)
- [Source code - DropManualMemoization.ts](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Inference/DropManualMemoization.ts)
