# preserve-manual-memoization

## Rule Details

Validates that manual `useMemo` and `useCallback` dependency arrays preserve the memoization behavior that React Compiler infers.

Examples of **incorrect** code for this rule:

```javascript
function Component({ data, filter }) {
  const filtered = useMemo(() => data.filter(filter), [data]);
  return <List items={filtered} />;
}
```

```javascript
function Component({ onUpdate, value }) {
  const handleClick = useCallback(() => {
    onUpdate(value);
  }, [onUpdate]);
  return <button onClick={handleClick} />;
}
```

Examples of **correct** code for this rule:

```javascript
function Component({ data, filter }) {
  const filtered = useMemo(() => data.filter(filter), [data, filter]);
  return <List items={filtered} />;
}
```

```javascript
function Component({ data, filter }) {
  const filtered = data.filter(filter);
  return <List items={filtered} />;
}
```

## Differences from ESLint

- rslint implements this rule as a syntax-level dependency preservation check and does not run the full React Compiler pipeline. Some compiler-only diagnostics about pruned memo blocks may not be reported.

## Original Documentation

- [react.dev - preserve-manual-memoization](https://react.dev/reference/eslint-plugin-react-hooks/lints/preserve-manual-memoization)
- [Source code](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Validation/ValidatePreservedManualMemoization.ts)
