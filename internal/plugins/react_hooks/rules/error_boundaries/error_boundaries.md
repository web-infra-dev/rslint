# error-boundaries

Validates usage of Error Boundaries instead of `try` / `catch` for errors in
child components.

## Rule Details

`try` / `catch` blocks cannot catch errors that happen later during React's
rendering process. JSX only describes UI to render later, so render errors from
that child tree should be handled by an Error Boundary.

This rule follows React Compiler's lint target selection: it checks top-level
React components, custom Hooks, nested custom Hooks, and `memo` / `forwardRef`
callbacks once the file contains React code. JSX in the `catch` or `finally`
block of the same `try` statement is not reported.

Examples of **incorrect** code for this rule:

```javascript
function Parent() {
  try {
    return <ChildComponent />;
  } catch (error) {
    return <div>Error occurred</div>;
  }
}
```

```javascript
function Parent() {
  try {
    let child;
    try {
      child = getChild();
    } catch {
      return <ChildComponent value={child} />;
    }
  } catch {
    return null;
  }
}
```

Examples of **correct** code for this rule:

```javascript
function Parent() {
  return (
    <ErrorBoundary>
      <ChildComponent />
    </ErrorBoundary>
  );
}
```

```javascript
function Parent() {
  try {
    doSomething();
  } catch (error) {
    return <div>Error occurred</div>;
  }
}
```

## Differences from ESLint

- Files that use Flow component or hook syntax are not analyzed because rslint
  does not parse those declarations.

## Original Documentation

- [react.dev - error-boundaries](https://react.dev/reference/eslint-plugin-react-hooks/lints/error-boundaries)
- [Source code](https://github.com/facebook/react/blob/main/compiler/packages/babel-plugin-react-compiler/src/Validation/ValidateNoJSXInTryStatement.ts)
