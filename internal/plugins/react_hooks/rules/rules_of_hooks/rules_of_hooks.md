# rules-of-hooks

Enforce React's Rules of Hooks: hooks must only be called at the top level of
React function components or custom Hooks, never inside loops, conditions,
nested functions, class components, or async functions.

## Rule Details

This rule reports React Hook calls that violate the [Rules of Hooks](https://react.dev/reference/rules/rules-of-hooks):

- A hook (a function whose name is `use` or starts with `use[A-Z0-9]`, or a
  member access like `Namespace.useFoo`) must be called at the top level of a
  function component or custom hook — never inside `if` / `else`, ternary
  expressions, `&&` / `||` / `??` chains, `try` / `catch` blocks, loops, or
  callbacks.
- A hook must not be called after an early `return` or `throw`.
- A hook must not be called inside a class component (method, accessor, or
  class-field arrow), inside an `async` function, or at the top level of a
  module.
- The React `use(...)` hook is exempt from the loop, conditional, and
  early-return checks (it may be called conditionally), but is still rejected
  inside `try` / `catch` blocks and async functions.
- Functions created with `useEffectEvent(...)` may only be called from inside
  a React effect hook (`useEffect`, `useLayoutEffect`, `useInsertionEffect`,
  or any hook matching the `additionalEffectHooks` setting) or another
  `useEffectEvent` callback in the same component.

Examples of **incorrect** code for this rule:

```javascript
function ComponentWithConditionalHook() {
  if (cond) {
    useConditionalHook();
  }
}

function ComponentWithHookInsideLoop() {
  while (cond) {
    useHookInsideLoop();
  }
}

function ComponentWithHookAfterEarlyReturn() {
  if (a) return;
  useState();
}

class ClassComponentWithHook extends React.Component {
  render() {
    React.useState();
  }
}

async function AsyncComponent() {
  useState();
}

function notAComponent() {
  useState();
}
```

Examples of **correct** code for this rule:

```javascript
function ComponentWithHook() {
  useHook();
}

function useCustomHook() {
  useState();
}

const FancyButton = React.forwardRef((props, ref) => {
  useHook();
  return <button {...props} ref={ref} />;
});

function App() {
  // `use(...)` may be called conditionally
  if (shouldShowText) {
    const text = use(query);
  }
}
```

## Options

> **Warning**: options passed to this rule are ignored. The rule's schema
> accepts `{ "additionalHooks": "..." }` — exactly as upstream
> `eslint-plugin-react-hooks` declares it — but, also exactly as upstream, the
> implementation never reads it: upstream added the schema entry and the
> settings-based implementation in the same commit
> ([facebook/react#34497](https://github.com/facebook/react/pull/34497)) and
> only ever wired up the latter. Configure extra effect hooks through the
> `additionalEffectHooks` **setting** below instead.

## Settings

```json
{
  "settings": {
    "react-hooks": {
      "additionalEffectHooks": "(useMyEffect|useServerEffect)"
    }
  }
}
```

`additionalEffectHooks`
: A regular expression (matched against the bare hook name) that names extra
hooks for which `useEffectEvent`-bound identifiers may be referenced as
callbacks. Defaults to none.

## Original Documentation

- [react.dev — Rules of Hooks](https://react.dev/reference/rules/rules-of-hooks)
- [Source code](https://github.com/facebook/react/blob/main/packages/eslint-plugin-react-hooks/src/rules/RulesOfHooks.ts)
