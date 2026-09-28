# jsx-no-constructed-context-values

Disallow newly constructed context values that can cause unnecessary rerenders.

## Rule details

React compares context values by identity. Passing a new object, array, function,
or JSX value on every render can cause consumers to rerender even when the
underlying data has not changed.

This rule checks `value` expressions on `<Context.Provider>` inside React
components. It also recognizes React 19 `<Context>` providers when the context
is declared with `createContext()` or `React.createContext()`.

Examples of **incorrect** code for this rule:

```jsx
function Component() {
  return <SomeContext.Provider value={{ foo: 'bar' }} />;
}
```

```jsx
const MyContext = React.createContext();

function AnotherComponent() {
  function callback() {}
  return <MyContext value={callback} />;
}
```

Examples of **correct** code for this rule:

```jsx
function Component() {
  const value = useMemo(() => ({ foo: 'bar' }), []);
  return <SomeContext.Provider value={value} />;
}
```

```jsx
const MyContext = React.createContext();

function AnotherComponent() {
  const callback = useCallback(() => {}, []);
  return <MyContext value={callback} />;
}
```

Diagnostics point to the original construction, including when the value is
passed through a local alias. Function constructions suggest `useCallback`;
other constructions suggest `useMemo`. The rule does not provide automatic
fixes or editor suggestions because choosing dependencies requires context.

As in upstream, values declared outside the provider's current scope are not
reported. The rule does not check spread attributes or later assignments to a
variable. React 19 `<Context>` detection requires a local context declaration;
imported contexts can still be checked when used as `<Context.Provider>`.

## Options

This rule has no options. Extra options are accepted but have no effect.

## Differences from upstream

For unfinished code containing circular references such as `let a = b; let b = a;`,
rslint continues linting instead of failing. It does not report the circular
reference itself, but still reports any newly constructed value it can identify.
eslint-plugin-react v7.37.5 can fail when one of these variables is passed to a
context provider.

## Original documentation

- [eslint-plugin-react: jsx-no-constructed-context-values](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-no-constructed-context-values.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-no-constructed-context-values.js)
