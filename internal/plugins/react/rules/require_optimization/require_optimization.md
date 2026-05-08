# react/require-optimization

Enforce React components to declare a `shouldComponentUpdate` method (or use an equivalent optimization mechanism such as extending `React.PureComponent` or applying a PureRender decorator/mixin).

## Rule Details

This rule prevents you from creating React components without declaring a `shouldComponentUpdate` method.

Examples of **incorrect** code for this rule:

```jsx
class YourComponent extends React.Component {

}
```

```jsx
createReactClass({
});
```

Examples of **correct** code for this rule:

```jsx
class YourComponent extends React.Component {
  shouldComponentUpdate () {
    return false;
  }
}
```

```jsx
class YourComponent extends React.PureComponent {
}
```

```jsx
createReactClass({
  shouldComponentUpdate: function () {
    return false;
  }
});
```

```jsx
createReactClass({
  mixins: [PureRenderMixin]
});
```

```jsx
@reactMixin.decorate(PureRenderMixin)
class YourComponent extends Component {
}
```

## Rule Options

```json
{ "react/require-optimization": ["error", { "allowDecorators": [] }] }
```

- `allowDecorators`: optional array of decorator names. When a class is decorated with one of these decorators, the rule treats it as already optimized.

### `allowDecorators`

```json
{ "react/require-optimization": ["error", { "allowDecorators": ["pureRender"] }] }
```

Examples of **correct** code with the option above:

```jsx
@pureRender
class Hello extends React.Component {}
```

The decorator name must appear as a bare identifier in the class's decorator list. Call-form decorators (`@pureRender()`) and member-access forms (`@some.pureRender`) do not match this option — only the matching `@reactMixin.decorate(PureRenderMixin)` shape, which is recognized independently of `allowDecorators`, accepts a call form.

## Original Documentation

- [eslint-plugin-react: require-optimization](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/require-optimization.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/require-optimization.js)
