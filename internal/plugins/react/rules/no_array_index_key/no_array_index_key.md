# no-array-index-key

Warn if an array `index` is used as the `key` prop of a React element rendered by `Array.prototype.map`, `forEach`, `filter`, `find`, `findIndex`, `flatMap`, `reduce`, `reduceRight`, `some`, `every`, or `React.Children.map` / `React.Children.forEach`.

`key` should identify an element across renders. Since the array index changes whenever an item is added, removed, or reordered, using it as `key` defeats React's reconciliation and can cause subtle state leakage and rendering bugs.

## Rule Details

The following shapes are reported when they reference the iterator callback's index parameter:

- Direct identifier (`key={i}`)
- Template literal substitution (`` key={`item-${i}`} ``)
- Concatenation chain (`key={'item-' + i}`, `key={'item-' + i + '-x'}`)
- Coercion via `i.toString()` or `String(i)`

This applies to both JSX `key` attributes and the `key` property of the props object passed to `React.createElement` / `React.cloneElement` (or to bare `createElement` / `cloneElement` imported from `'react'`).

Examples of **incorrect** code for this rule:

```jsx
things.map((thing, index) => <Hello key={index} />);
```

```jsx
things.map((thing, index) => <Hello key={`item-${index}`} />);
```

```jsx
things.map((thing, index) => <Hello key={'item-' + index} />);
```

```jsx
things.map((thing, index) => React.cloneElement(thing, { key: index }));
```

```jsx
things.forEach((thing, index) => {
  otherThings.push(<Hello key={index} />);
});
```

```jsx
React.Children.map(this.props.children, (child, index) =>
  React.cloneElement(child, { key: index }),
);
```

Examples of **correct** code for this rule:

```jsx
things.map((thing) => <Hello key={thing.id} />);
```

```jsx
things.map((thing, index) => <Hello key={thing.id} />);
```

```jsx
React.Children.map(this.props.children, (child) =>
  React.cloneElement(child, { key: child.id }),
);
```

## When Not To Use It

If you have a stable list of items that will never be reordered, inserted into, or removed from, the consequences of an unstable key may be acceptable. In that case, disable the rule for the specific call site rather than globally.

## Original Documentation

- [eslint-plugin-react / no-array-index-key](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-array-index-key.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/no-array-index-key.js)
