# globals

Prevents assignments to variables that are declared outside a React component
or Hook while React is rendering.

## Rule Details

Validates against reassigning variables declared outside of a component or Hook
during render.

The rule checks assignments in the component or Hook body and in helper
functions that are called during render, such as direct helper calls,
`useMemo` callbacks, JSX children, and render-prop helpers that return JSX.
Event handlers, effect callbacks, and `useCallback` callbacks are not render
execution for this rule.

Property writes such as `window.location.href = value` are not reported by this
rule; those mutations are handled by React's immutability lint.

Examples of **incorrect** code for this rule:

```javascript
let moduleLocal;

function Component() {
  moduleLocal = true;
  return <div />;
}
```

```javascript
function Component() {
  const update = () => {
    someGlobal = true;
  };

  update();
  return <div />;
}
```

Examples of **correct** code for this rule:

```javascript
function Component() {
  let local;
  local = true;
  return <div />;
}
```

```javascript
function Component() {
  const onClick = () => {
    someGlobal = true;
  };

  return <button onClick={onClick} />;
}
```

## Differences from ESLint

- Files that use Flow component or hook syntax are not analyzed because rslint
  does not parse those declarations.

## Original Documentation

- [react.dev - globals](https://react.dev/reference/eslint-plugin-react-hooks/lints/globals)
- [Source code](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Inference/InferMutationAliasingEffects.ts)
