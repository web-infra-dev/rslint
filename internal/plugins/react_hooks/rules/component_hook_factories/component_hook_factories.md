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

## Options

```json
{
  "react-hooks/component-hook-factories": [
    "error",
    {
      "environment": {
        "hookPattern": "^signal[A-Z]"
      }
    }
  ]
}
```

- `environment.hookPattern`: A JavaScript regular expression pattern for
  custom Hook names. Function names that match this pattern are treated as
  Hooks when the rule classifies nested factories. When omitted or invalid,
  the rule falls back to React's default Hook naming convention.

Examples of **incorrect** code for this rule with `{ "environment": { "hookPattern": "^signal[A-Z]" } }`:

```javascript
function createSignalHook(source) {
  return function signalValue() {
    return signalRead(source);
  };
}
```

## Original Documentation

- [react.dev — component-hook-factories](https://react.dev/reference/eslint-plugin-react-hooks/lints/component-hook-factories)
