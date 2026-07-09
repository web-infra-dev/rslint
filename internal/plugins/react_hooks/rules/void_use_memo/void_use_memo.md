# react-hooks/void-use-memo

## Rule Details

Validates that `useMemo()` callbacks return a value and that the memoized value is used.

`useMemo()` is for computing and caching values. If the callback has no return, or if the result of `useMemo()` is discarded, the hook is likely being used for side effects instead.

Examples of **incorrect** code for this rule:

```javascript
function Component({ value }) {
  const memoized = useMemo(() => {
    console.log(value);
  }, [value]);
  return <div>{memoized}</div>;
}
```

```javascript
function Component() {
  useMemo(() => {
    return [];
  }, []);
  return <div />;
}
```

Examples of **correct** code for this rule:

```javascript
function Component({ value }) {
  const memoized = useMemo(() => {
    return value * 2;
  }, [value]);
  return <div>{memoized}</div>;
}
```

```javascript
function Component({ value }) {
  useEffect(() => {
    console.log(value);
  }, [value]);
  return <div>{value}</div>;
}
```

Like upstream, this rule tracks direct `useMemo()` and `React.useMemo()` call shapes. It checks that the callback has an explicit or implicit return; it does not inspect the returned value.

## Options

This rule has no rule-specific options. Like upstream, it accepts an option object for forward compatibility and ignores it.

## Original Documentation

- [react.dev - void-use-memo](https://react.dev/reference/eslint-plugin-react-hooks/lints/void-use-memo)
- [Source code - ValidateUseMemo.ts](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Validation/ValidateUseMemo.ts)
