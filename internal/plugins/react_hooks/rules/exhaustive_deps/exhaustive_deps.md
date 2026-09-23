# exhaustive-deps

## Rule Details

Verifies the list of dependencies for Hooks like `useEffect`, `useCallback`,
`useMemo`, `useImperativeHandle` and `useLayoutEffect`.
Reports missing, unnecessary or duplicate entries in the dependency array
of these Hooks, and offers a suggested fix. `useInsertionEffect` is checked
only when configured through `additionalHooks` or the shared settings below.

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

- **`additionalHooks`** (string regex, default empty): Check the named
  custom hooks with the callback at index 0 and dependencies at index 1.
  Names matching `Effect($|[^a-z])` use effect-specific checks. Only matches
  bare-identifier callees — `Foo.useBar` and `React.useBar` are NOT
  matched even if the identifier suffix is in the regex (mirrors
  upstream's `node === calleeNode` gate). Falls back to
  `settings['react-hooks'].additionalEffectHooks` when omitted or empty.
  ```json
  { "react-hooks/exhaustive-deps": ["error", { "additionalHooks": "(useMyEffect|useAsync)" }] }
  ```
- **`enableDangerousAutofixThisMayCauseInfiniteLoops`** (boolean, default
  `false`): Promote all edits from the first suggestion into a top-level
  autofix while keeping the suggestion array. Off by default because
  applying it without code review can introduce render loops.
- **`requireExplicitEffectDeps`** (boolean, default `false`): Require
  effect-style Hooks to receive a dependencies argument. An explicit
  `undefined` satisfies this requirement; omitting the argument does not.
- **`experimental_autoDependenciesHooks`** (string array, default `[]`):
  Skip dependency-array and missing-array state-update checks for the named
  Hooks when their deps argument is `null`, `undefined` or absent.
  Async callbacks, stale assignments and cleanup ref accesses are still checked.
  The Hook must already be recognized as a built-in or through `additionalHooks`
  or shared settings. `useMemo` and `useCallback` still warn without an array.
  ```json
  {
    "react-hooks/exhaustive-deps": [
      "error",
      {
        "additionalHooks": "useAutoEffect",
        "experimental_autoDependenciesHooks": ["useAutoEffect"]
      }
    ]
  }
  ```

## Differences from ESLint

- Components written in Flow syntax (`component MyComp() { ... }` or
  `hook useFoo() { ... }`) are not analyzed; the rule produces no
  diagnostics on them.
- The order of diagnostics in the returned list may differ from ESLint.
  Diagnostics retain their source locations and associated fixes.
- For a state name beginning with a non-BMP character, the functional-update
  hint preserves the complete first character. For example,
  `const [𐀀, setX] = useState(0); useEffect(() => setX(𐀀 + 1), []);`
  suggests `setX(𐀀 => ...)`. ESLint truncates that character to one UTF-16
  code unit, producing an unpaired surrogate. Dependency analysis and fixes
  are unaffected.

## Original Documentation

- [react.dev — Rules of Hooks](https://react.dev/reference/rules/rules-of-hooks)
- [Source code](https://github.com/facebook/react/blob/eslint-plugin-react-hooks@7.1.1/packages/eslint-plugin-react-hooks/src/rules/ExhaustiveDeps.ts)
