# forbid-dom-props

Forbid certain props from being used on DOM Nodes (e.g. `<div />`). This rule
only applies to DOM Nodes and does not affect Components (e.g. `<Foo />`). The
list of forbidden props is configured via the `forbid` option.

## Rule Details

This rule checks all JSX elements and verifies that no forbidden props are used
on DOM Nodes. This rule is off by default.

Examples of **incorrect** code for this rule:

```json
{ "react/forbid-dom-props": ["error", { "forbid": ["id"] }] }
```

```jsx
<div id='Joe' />
```

Examples of **correct** code for this rule:

```json
{ "react/forbid-dom-props": ["error", { "forbid": ["id"] }] }
```

```jsx
<Hello id='foo' />
```

## Rule Options

```json
{ "react/forbid-dom-props": ["error", { "forbid": ["id", "style"] }] }
```

### `forbid`

An array specifying the names of props that are forbidden. The default value of
this option is `[]`. Each array element can either be a string with the
property name, or an object specifying the property name, an optional custom
message, and DOM nodes the prop is `disallowedFor`:

```json
{
  "react/forbid-dom-props": [
    "error",
    {
      "forbid": [
        {
          "propName": "someProp",
          "disallowedFor": ["DOMNode", "AnotherDOMNode"],
          "message": "Avoid using someProp on DOMNode and AnotherDOMNode"
        }
      ]
    }
  ]
}
```

You can also forbid a prop only when it has a particular value, by using the
`disallowedValues` option:

```json
{
  "react/forbid-dom-props": [
    "error",
    {
      "forbid": [
        {
          "propName": "someProp",
          "disallowedValues": ["someValue", "anotherValue"],
          "message": "Avoid using someProp with values someValue and anotherValue"
        }
      ]
    }
  ]
}
```

`disallowedValues` and `disallowedFor` can be combined to restrict a prop to a
particular value on a particular DOM Node:

```json
{
  "react/forbid-dom-props": [
    "error",
    {
      "forbid": [
        {
          "propName": "someProp",
          "disallowedFor": ["DOMNode"],
          "disallowedValues": ["someValue"],
          "message": "Avoid using someProp with value someValue on DOMNode"
        }
      ]
    }
  ]
}
```

## Original Documentation

- [eslint-plugin-react/forbid-dom-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/forbid-dom-props.md)
