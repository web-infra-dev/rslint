# react/jsx-no-bind

Disallow `.bind()` or arrow functions in JSX props.

## Rule Details

A new function (or an arrow function) is created on every render when `.bind()` or an inline function expression is used as a JSX prop. Passing a new function reference can trigger unnecessary re-renders in memoized consumers or cause effects to re-run.

Examples of **incorrect** code for this rule:

```jsx
<Foo onClick={this._handleClick.bind(this)} />
```

```jsx
<Foo onClick={() => console.log('Hello!')} />
```

```jsx
function onClick() { console.log('Hello!'); }
<Foo onClick={onClick} />
```

Examples of **correct** code for this rule:

```jsx
<Foo onClick={this._handleClick} />
```

## Rule Options

```json
{
  "react/jsx-no-bind": ["error", {
    "allowArrowFunctions": false,
    "allowBind": false,
    "allowFunctions": false,
    "ignoreRefs": false,
    "ignoreDOMComponents": false
  }]
}
```

- `allowArrowFunctions`: allow arrow function expressions as JSX prop values.
- `allowBind`: allow `.bind()` calls as JSX prop values.
- `allowFunctions`: allow function expressions/declarations as JSX prop values.
- `ignoreRefs`: skip checks on `ref` props.
- `ignoreDOMComponents`: skip checks on DOM components (lowercase tags like `<div>`).

## Differences from ESLint

- The ES bind operator (`::`, e.g. `<div foo={::this.onChange} />`) is not supported. Empirically verified: TypeScript's parser rejects the `::` token as a syntax error, so such code never reaches the rule. Consequently the `bindExpression` messageId is not produced.

## Known Limitations (matches ESLint)

- Identifiers used inside conditional expressions are not resolved against tracked declarations, so `<Foo onClick={cond ? tracked : other} />` does not flag `tracked`.
- Only `const` declarations are tracked; `let` / `var` bindings are ignored even when initialized to an arrow / bind / function.
- Forward references (JSX used before the declaration in source order) are not reported.
- A non-violating inner `const` does not shadow a violating outer one; JSX inside the inner scope still reports the outer violation.
- TypeScript-only wrappers (`as T`, `<T>expr`, `expr!`, `expr satisfies T`) are opaque to the rule, matching ESLint's behavior under the `@typescript-eslint/parser`: for example `<Foo onClick={(() => 1) as Handler} />` is **not** flagged. Plain parentheses around an expression *are* transparent because ESTree does not represent them as nodes.

## Original Documentation

- [eslint-plugin-react/jsx-no-bind](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-bind.md)
