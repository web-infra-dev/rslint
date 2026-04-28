# no-redundant-should-component-update

## Rule Details

Disallow usage of `shouldComponentUpdate` when extending `React.PureComponent`.

`React.PureComponent` already implements `shouldComponentUpdate` with a shallow
prop and state comparison. Defining your own `shouldComponentUpdate` defeats
the purpose of extending `PureComponent` — at that point you should extend
`React.Component` instead.

Examples of **incorrect** code for this rule:

```jsx
class Foo extends React.PureComponent {
  shouldComponentUpdate() {
    // do check
  }

  render() {
    return <div>Radical!</div>
  }
}

function Bar() {
  return class Baz extends React.PureComponent {
    shouldComponentUpdate() {
      // do check
    }

    render() {
      return <div>Groovy!</div>
    }
  }
}
```

Examples of **correct** code for this rule:

```jsx
class Foo extends React.Component {
  shouldComponentUpdate() {
    // do check
  }

  render() {
    return <div>Radical!</div>
  }
}

function Bar() {
  return class Baz extends React.Component {
    shouldComponentUpdate() {
      // do check
    }

    render() {
      return <div>Groovy!</div>
    }
  }
}

class Qux extends React.PureComponent {
  render() {
    return <div>Tubular!</div>
  }
}
```

## Original Documentation

- [eslint-plugin-react: no-redundant-should-component-update](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-redundant-should-component-update.md)
