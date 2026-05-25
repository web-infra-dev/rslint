# no-is-mounted

Disallow usage of `isMounted`.

`isMounted` is an anti-pattern, is not available when using ES6 classes, and is
[officially deprecated](https://facebook.github.io/react/blog/2015/12/16/ismounted-antipattern.html)
by the React team.

## Rule Details

This rule flags calls to `this.isMounted()` inside properties of an object
literal (e.g. a `createReactClass` spec) or inside methods / getters / setters /
constructors of a class. Calls outside any such method are ignored.

Examples of **incorrect** code for this rule:

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    if (!this.isMounted()) {
      return;
    }
  },
  render: function () {
    return <div>Hello</div>;
  },
});
```

```javascript
class Hello extends React.Component {
  someMethod() {
    if (!this.isMounted()) {
      return;
    }
  }
  render() {
    return <div onClick={this.someMethod.bind(this)}>Hello</div>;
  }
}
```

Examples of **correct** code for this rule:

```javascript
var Hello = createReactClass({
  render: function () {
    return <div>Hello</div>;
  },
});
```

```javascript
var Hello = createReactClass({
  componentDidUpdate: function () {
    someNonMemberFunction(arg);
    this.someFunc = this.isMounted;
  },
  render: function () {
    return <div>Hello</div>;
  },
});
```

```javascript
class Hello extends React.Component {
  notIsMounted() {}
  render() {
    this.notIsMounted();
    return <div>Hello</div>;
  }
}
```

## Original Documentation

- [eslint-plugin-react / no-is-mounted](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-is-mounted.md)
