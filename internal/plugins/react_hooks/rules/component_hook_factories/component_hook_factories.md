# component-hook-factories

## Rule Details

Validates against higher order functions defining nested components or hooks.
Components and hooks should be defined at the module level.

Examples of **incorrect** code for this rule:

```javascript
function createComponent(defaultValue) {
  return function Component() {
    return <span>{defaultValue}</span>;
  };
}

function Parent() {
  function Child() {
    return <div />;
  }

  return <Child />;
}

function createCustomHook(endpoint) {
  return function useData() {
    return useMemo(() => endpoint, [endpoint]);
  };
}
```

Examples of **correct** code for this rule:

```javascript
function Component({ defaultValue }) {
  return <span>{defaultValue}</span>;
}

function useData(endpoint) {
  return useMemo(() => endpoint, [endpoint]);
}

function Button({ color, children }) {
  return <button style={{ backgroundColor: color }}>{children}</button>;
}
```

## Original Documentation

- [react.dev — component-hook-factories](https://react.dev/reference/eslint-plugin-react-hooks/lints/component-hook-factories)
