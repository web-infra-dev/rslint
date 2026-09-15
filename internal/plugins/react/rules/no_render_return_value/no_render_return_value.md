# no-render-return-value

Disallow usage of the return value of `ReactDOM.render`.

`ReactDOM.render()` currently returns a reference to the root `ReactComponent`
instance. However, using this return value is legacy and should be avoided
because future versions of React may render components asynchronously in some
cases. If you need a reference to the root `ReactComponent` instance, the
preferred solution is to attach a
[callback ref](https://reactjs.org/docs/refs-and-the-dom.html#callback-refs)
to the root element.

## Rule Details

The rule flags `ReactDOM.render` calls whose return value is consumed — i.e.
when the call sits in one of these positions:

- Variable initializer (`var x = ReactDOM.render(...)`)
- Object property value or computed key (`{ k: ReactDOM.render(...) }`,
  `{ [ReactDOM.render(...)]: value }`)
- `return` argument (`return ReactDOM.render(...)`)
- Arrow function expression body (`(a, b) => ReactDOM.render(a, b)`)
- Right-hand side of an assignment (`x = ReactDOM.render(...)`)

Examples of **incorrect** code for this rule:

```javascript
const inst = ReactDOM.render(<App />, document.body);
doSomethingWithInst(inst);
```

Examples of **correct** code for this rule:

```javascript
ReactDOM.render(<App ref={(inst) => doSomethingWithInst(inst)} />, document.body);

ReactDOM.render(<App />, document.body, () => {
  // the render has finished
});
```

## React Version

The callee object pattern depends on `settings.react.version`. When `version`
is omitted, `settings.react.defaultVersion` is used; when both are omitted, the
rule assumes the latest React version.

| Version range | Matched object(s) |
| ------------- | ----------------- |
| `>= 15.0.0` (default) | `ReactDOM` |
| `^0.14.0` | `React` or `ReactDOM` |
| `^0.13.0` | `React` |

Any other version (e.g. `0.0.1`) falls back to `ReactDOM`-only, matching
upstream's default branch.

## Differences from upstream

- Rslint reports optional-chain forms such as
  `var instance = ReactDOM?.render(<App />, root)`. The value still depends on
  the call result when the call runs (and on `undefined` otherwise); upstream
  skips it with modern ESTree parsers because they insert a `ChainExpression`
  between the call and the consuming parent.
- Rslint reports `ReactDOM.render(...)` used as a default value inside a
  destructuring assignment, such as
  `[instance = ReactDOM.render(<App />, root)] = values`. The default consumes
  the call result when selected; upstream skips it because ESTree classifies
  the inner `=` as an `AssignmentPattern` rather than an `AssignmentExpression`.

## Original Documentation

- [eslint-plugin-react: no-render-return-value](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-render-return-value.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-render-return-value.js)
