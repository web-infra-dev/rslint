# jsx-no-undef

## Rule Details

This rule helps locate potential `ReferenceError`s resulting from misspellings or missing components. It flags any JSX tag whose leftmost identifier is not in scope.

Examples of **incorrect** code for this rule:

```jsx
<Hello name="John" />;
```

```jsx
var Hello = React.createClass({
  render: function () {
    return <Text>Hello</Text>;
  },
});
module.exports = Hello;
```

Examples of **correct** code for this rule:

```jsx
var Hello = require('./Hello');

<Hello name="John" />;
```

- Intrinsic (lowercase) tags such as `<div>`, `<span>`, `<svg:path>`, and `<x-gif>` are always allowed.
- Namespaced JSX tags such as `<a:b />` are skipped.
- For member-expression tags (`<Foo.Bar>`, `<foo.bar.Baz>`), only the leftmost identifier is checked. `<this.Foo>` is always allowed.

## Differences from ESLint

- The `allowGlobals` option is **not supported**. A JSX tag identifier that is not declared in the source file (via `var` / `let` / `const` / `function` / `class` / `enum` / `namespace` / `import` / `declare` / function parameter / catch binding / loop binding) is always reported. To silence a report for an ambient value, add an explicit declaration — e.g. `declare const Foo: any;` or `import Foo from '...';`.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-undef.md
