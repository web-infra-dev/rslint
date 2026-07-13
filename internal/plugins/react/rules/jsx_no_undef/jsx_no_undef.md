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

## Options

### `allowGlobals`

When `false` (default), a JSX tag identifier in a module (a file with `import`/`export`) is only recognized if it's declared in the source file (via `var` / `let` / `const` / `function` / `class` / `enum` / `namespace` / `import` / `declare` / function parameter / catch binding / loop binding) — names declared only via config `languageOptions.globals` or `/* global */` comments are still reported. Set `allowGlobals: true` to also allow those. Script files (no `import`/`export`) always consult declared globals, regardless of this option — matching ESLint.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-undef.md
