# immutability

## Rule Details

Validates against mutating component props, hook arguments, state values, hook
return values, values that have already been passed to hooks, and local values
that have already been used to create JSX.

React treats props and state as immutable snapshots. Mutating them directly can
leave React with the same object or array reference, so React has no reliable
signal that the UI needs to update.

The rule reports direct assignments, update/delete expressions, and known
mutating calls such as `push`, `sort`, `splice`, and `Object.assign` when their
target is an immutable value or an alias derived from one.

Examples of **incorrect** code for this rule:

```javascript
function Component() {
  const [items, setItems] = useState([1, 2, 3]);

  const addItem = () => {
    items.push(4);
    setItems(items);
  };
}
```

```javascript
function Component({ user }) {
  user.name = "Alice";
  return <div>{user.name}</div>;
}
```

Examples of **correct** code for this rule:

```javascript
function Component() {
  const [items, setItems] = useState([1, 2, 3]);

  const addItem = () => {
    setItems([...items, 4]);
  };
}
```

```javascript
function Component({ user, setUser }) {
  setUser({ ...user, name: "Alice" });
  return <div>{user.name}</div>;
}
```

## Differences from ESLint

- Files that use Flow component or hook syntax are not analyzed because rslint
  does not parse those declarations.

## Original Documentation

- [react.dev - immutability](https://react.dev/reference/eslint-plugin-react-hooks/lints/immutability)
- [Source code](https://github.com/facebook/react/blob/main/packages/eslint-plugin-react-hooks/src/shared/ReactCompiler.ts)
