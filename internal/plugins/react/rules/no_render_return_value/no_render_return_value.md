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
- Object property value (`{ k: ReactDOM.render(...) }`)
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

The callee object pattern depends on `settings.react.version`:

| Version range | Matched object(s) |
| ------------- | ----------------- |
| `>= 15.0.0` (default) | `ReactDOM` |
| `^0.14.0` | `React` or `ReactDOM` |
| `^0.13.0` | `React` |

Any other version (e.g. `0.0.1`) falls back to `ReactDOM`-only, matching
upstream's default branch.

## Original Documentation

- [eslint-plugin-react / no-render-return-value](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-render-return-value.md)
