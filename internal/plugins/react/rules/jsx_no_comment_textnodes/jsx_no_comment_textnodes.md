# jsx-no-comment-textnodes

Disallow comments from being inserted as text nodes.

This rule prevents comment strings (e.g. beginning with `//` or `/*`) from being
accidentally injected as a text node in JSX statements. Comments in JSX must be
wrapped in an expression container (`{/* ... */}`) to be treated as actual
comments instead of literal text.

## Rule Details

Examples of **incorrect** code for this rule:

```jsx
var Hello = createReactClass({
  render: function () {
    return <div>// empty div</div>;
  },
});

var Hello = createReactClass({
  render: function () {
    return (
      <div>
        /* empty div */
      </div>
    );
  },
});
```

Examples of **correct** code for this rule:

```jsx
var Hello = createReactClass({
  displayName: 'Hello',
  render: function () {
    return <div>{/* empty div */}</div>;
  },
});

var Hello = createReactClass({
  displayName: 'Hello',
  render: function () {
    return <div /* empty div */></div>;
  },
});

var Hello = createReactClass({
  displayName: 'Hello',
  render: function () {
    return <div className={'foo' /* temp class */}></div>;
  },
});
```

## Legitimate uses

It is possible to intentionally output comment-start characters (`//` or `/*`)
inside a JSX text node — wrap the content in an expression container so the
text is parsed as a string literal instead of a raw text node:

```jsx
var Hello = createReactClass({
  render: function () {
    return <div>{'/* This will be output as a text node */'}</div>;
  },
});
```

## Original Documentation

- [eslint-plugin-react / jsx-no-comment-textnodes](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-comment-textnodes.md)
