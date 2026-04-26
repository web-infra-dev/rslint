# react-hooks/exhaustive-deps

## Rule Details

Verifies the list of dependencies for Hooks like `useEffect`, `useCallback`,
`useMemo`, `useImperativeHandle`, `useLayoutEffect` and `useInsertionEffect`.
Reports missing, unnecessary or duplicate entries in the second argument
of these Hooks, and offers a suggested fix.

Examples of **incorrect** code for this rule:

```javascript
function MyComponent({ id }) {
  // 'id' is captured by the effect but missing from deps
  useEffect(() => {
    console.log(id);
  }, []);
}
```

```javascript
function MyComponent({ a, b }) {
  // 'b' is in deps but never used inside the callback
  useCallback(() => a, [a, b]);
}
```

```javascript
function MyComponent({ list }) {
  // Spread elements can't be statically checked
  useEffect(() => {}, [...list]);
}
```

```javascript
function MyComponent() {
  const ref = useRef(null);
  useEffect(() => {
    return () => {
      // ref.current may have changed by cleanup time
      console.log(ref.current);
    };
  }, []);
}
```

Examples of **correct** code for this rule:

```javascript
function MyComponent({ id }) {
  useEffect(() => {
    console.log(id);
  }, [id]);
}
```

```javascript
function MyComponent({ list }) {
  useMemo(() => list.length, [list]);
}
```

```javascript
function MyComponent() {
  // useState setter is stable; no need to list it
  const [count, setCount] = useState(0);
  useEffect(() => {
    setCount(c => c + 1);
  }, []);
}
```

```javascript
function MyComponent({ theme }) {
  // useEffectEvent return values are stable
  const onClick = useEffectEvent(() => {
    console.log(theme);
  });
  useEffect(() => {
    onClick();
  }, []);
}
```

## Options

The rule accepts a single options object:

```json
{
  "react-hooks/exhaustive-deps": [
    "error",
    {
      "additionalHooks": "(useMyEffect|useAsync)",
      "enableDangerousAutofixThisMayCauseInfiniteLoops": false,
      "requireExplicitEffectDeps": false
    }
  ]
}
```

- **`additionalHooks`** (string regex, default empty): Treat the named
  custom hooks as effect-style Hooks (callback at index 0). Only matches
  bare-identifier callees — `Foo.useBar` and `React.useBar` are NOT
  matched even if the identifier suffix is in the regex (mirrors
  upstream's `node === calleeNode` gate). Falls back to
  `settings['react-hooks'].additionalHooks` when omitted or empty.
  ```json
  { "react-hooks/exhaustive-deps": ["error", { "additionalHooks": "(useMyEffect|useAsync)" }] }
  ```
- **`enableDangerousAutofixThisMayCauseInfiniteLoops`** (boolean, default
  `false`): Promote the first suggestion's first fix into a top-level
  autofix while keeping the suggestion array. Off by default because
  applying it without code review can introduce render loops.
- **`requireExplicitEffectDeps`** (boolean, default `false`): Require
  effect-style Hooks to be passed an explicit deps array (or
  `undefined`). Useful in codebases that disable the deps-array fallback
  to make every effect's reactive surface explicit.
- **`experimental_autoDependenciesHooks`** (string array, default `[]`):
  Skip dependency analysis entirely for the named custom hooks when their
  deps argument is `null` or absent (the hook is expected to infer deps
  itself). Used by tooling that does its own dep auto-injection.
  ```json
  {
    "react-hooks/exhaustive-deps": [
      "error",
      { "experimental_autoDependenciesHooks": ["useAutoEffect"] }
    ]
  }
  ```

## Differences from ESLint

- Components written in Flow syntax (`component MyComp() { ... }` or
  `hook useFoo() { ... }`) are not analyzed; the rule produces no
  diagnostics on them.
- When multiple diagnostics are emitted for the same source line, the
  order in the diagnostics list may differ from ESLint. The set of
  diagnostics, their messages, and their fix outputs are identical;
  only the position of each entry within the array can vary. IDEs and
  CLI reporters that sort by source position display them identically.

## Original Documentation

- [react.dev — Rules of Hooks](https://react.dev/reference/rules/rules-of-hooks)
- [Source code](https://github.com/facebook/react/blob/main/packages/eslint-plugin-react-hooks/src/rules/ExhaustiveDeps.ts)
