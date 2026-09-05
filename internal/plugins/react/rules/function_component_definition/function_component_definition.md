# react/function-component-definition

## Rule Details

This rule enforces a consistent function type for React function components. By
default it prefers function declarations for named components and function
expressions for unnamed ones.

Examples of **incorrect** code for this rule:

```jsx
const Component = function (props) {
  return <div>{props.content}</div>;
};

const Component = (props) => {
  return <div>{props.content}</div>;
};

function getComponent() {
  return (props) => {
    return <div>{props.content}</div>;
  };
}
```

Examples of **correct** code for this rule:

```jsx
function Component(props) {
  return <div>{props.content}</div>;
}

function getComponent() {
  return function (props) {
    return <div>{props.content}</div>;
  };
}
```

## Rule Options

```json
{
  "react/function-component-definition": [
    "error",
    { "namedComponents": "function-declaration", "unnamedComponents": "function-expression" }
  ]
}
```

- `namedComponents` — `"function-declaration"` (default), `"function-expression"`,
  `"arrow-function"`, or an array of any of those.
- `unnamedComponents` — `"function-expression"` (default), `"arrow-function"`, or an
  array of either.

When an array is given, any listed form is accepted and the first entry is the
one the autofix rewrites to.

Examples of **incorrect** code with `{ "namedComponents": "arrow-function" }`:

```json
{ "react/function-component-definition": ["error", { "namedComponents": "arrow-function" }] }
```

```jsx
function Component(props) {
  return <div />;
}

const Component = function (props) {
  return <div />;
};
```

Examples of **correct** code with `{ "namedComponents": "arrow-function" }`:

```json
{ "react/function-component-definition": ["error", { "namedComponents": "arrow-function" }] }
```

```jsx
const Component = (props) => {
  return <div />;
};
```

Examples of **incorrect** code with `{ "unnamedComponents": "arrow-function" }`:

```json
{ "react/function-component-definition": ["error", { "unnamedComponents": "arrow-function" }] }
```

```jsx
function getComponent() {
  return function (props) {
    return <div />;
  };
}
```

Examples of **correct** code with `{ "unnamedComponents": "arrow-function" }`:

```json
{ "react/function-component-definition": ["error", { "unnamedComponents": "arrow-function" }] }
```

```jsx
function getComponent() {
  return (props) => {
    return <div />;
  };
}
```

## Unfixable patterns

Some reports intentionally come without an autofix, because no equivalent
rewrite exists:

- A default-exported function declaration, since `export default var Component = …`
  is not valid syntax.
- A named function expression, since the rewrite would drop its own binding.
- A variable with a type annotation being rewritten to a function declaration,
  since a function declaration has nowhere to put the annotation.
- A function with exactly one unconstrained type parameter being rewritten to an
  arrow function, since `<T>(props) => …` is ambiguous with JSX. Two type
  parameters, or one constrained type parameter, are fixable.

The autofix rewrites the function's head, so an `async` component loses its
`async` keyword: `async function Component(props) { … }` becomes
`const Component = (props) => { … }`. Review such a fix before accepting it.

## Differences from ESLint

- A function that returns no JSX is not treated as a component here, even when a
  later `Hello.propTypes = {}` / `Hello.defaultProps = {}` assignment marks it as
  one. ESLint reports `var Hello = function(props) { return 1; }` in that file;
  rslint does not.
- `namedComponents: []` / `unnamedComponents: []` allows no function form at all.
  The schema accepts an empty array, but ESLint then has no message to report and
  aborts the whole lint run with "Missing `message` property in report() call".
  rslint reports nothing for that category instead of failing the run.

## Original Documentation

- [eslint-plugin-react: function-component-definition](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/function-component-definition.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/function-component-definition.js)
