# no-direct-mutation-state

## Rule Details

Disallow direct mutation of `this.state`. State should only be updated through
`setState()` (or the `useState` setter), so React can correctly schedule a
re-render and reconcile the resulting UI. Writing to `this.state` directly is
allowed only inside the component's constructor, where the initial state is
being seeded.

The rule targets both ES6 class components extending `Component` /
`PureComponent` (or their pragma-qualified forms, e.g. `React.Component`) and
ES5 components created with `createReactClass(...)` (or
`<pragma>.<createClass>(...)`).

Examples of **incorrect** code for this rule:

```javascript
var Hello = createReactClass({
  render: function () {
    this.state.foo = "bar";
    return <div>Hello {this.props.name}</div>;
  },
});
```

```javascript
class Hello extends React.Component {
  componentDidMount() {
    this.state.foo = "bar";
  }
}
```

```javascript
class Hello extends React.Component {
  constructor(props) {
    super(props);
    doSomethingAsync(() => {
      this.state = "bad";
    });
  }
}
```

Examples of **correct** code for this rule:

```javascript
class Hello extends React.Component {
  constructor() {
    super();
    this.state = { foo: "bar" };
  }
  update() {
    this.setState({ foo: "baz" });
  }
}
```

```javascript
class Hello {
  getFoo() {
    this.state.foo = "bar";
    return this.state.foo;
  }
}
```

## Original Documentation

- [eslint-plugin-react `no-direct-mutation-state`](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-direct-mutation-state.md)
