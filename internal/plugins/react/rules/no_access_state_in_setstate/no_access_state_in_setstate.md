# no-access-state-in-setstate

## Rule Details

Disallow reading `this.state` inside the first argument of `this.setState(...)`.

Because React schedules and batches state updates asynchronously, the value
of `this.state` at the moment `setState` is called is not necessarily the
value the updater will see when it runs. Writing an update that derives from
the previous state via direct `this.state` access can therefore yield stale
results. The correct form passes a callback — `this.setState(prev => ...)` —
which always receives the latest committed state.

Examples of **incorrect** code for this rule:

```javascript
class Hello extends React.Component {
  onClick() {
    this.setState({ value: this.state.value + 1 });
  }
}
```

```javascript
class Hello extends React.Component {
  onClick() {
    var nextValue = this.state.value + 1;
    this.setState({ value: nextValue });
  }
}
```

```javascript
class Hello extends React.Component {
  onClick() {
    var { state } = this;
    this.setState({ value: state.value + 1 });
  }
}
```

```javascript
class Hello extends React.Component {
  nextState() {
    return this.state.value + 1;
  }
  onClick() {
    this.setState({ value: nextState() });
  }
}
```

Examples of **correct** code for this rule:

```javascript
class Hello extends React.Component {
  onClick() {
    this.setState(state => ({ value: state.value + 1 }));
  }
}
```

```javascript
class Hello extends React.Component {
  onClick() {
    this.setState({}, () => console.log(this.state));
  }
}
```

## Original Documentation

- ESLint rule: https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-access-state-in-setstate.md
