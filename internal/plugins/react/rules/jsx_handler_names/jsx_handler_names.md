# jsx-handler-names

## Rule Details

Enforce event handler naming conventions in JSX. This rule ensures that any component or prop methods used to handle events are correctly prefixed.

By default, event handler functions must be named `handleX` (camelCase, beginning with `handle`), and props that receive event handlers must be named `onX` (camelCase, beginning with `on`).

Examples of **incorrect** code for this rule:

```jsx
<MyComponent handleChange={this.handleChange} />
```

```jsx
<MyComponent onChange={this.componentChanged} />
```

Examples of **correct** code for this rule:

```jsx
<MyComponent onChange={this.handleChange} />
```

```jsx
<MyComponent onChange={this.props.onFoo} />
```

## Rule Options

```json
{
  "react/jsx-handler-names": [
    "error",
    {
      "eventHandlerPrefix": "handle",
      "eventHandlerPropPrefix": "on",
      "checkLocalVariables": false,
      "checkInlineFunction": false,
      "ignoreComponentNames": []
    }
  ]
}
```

- `eventHandlerPrefix`: Prefix for component methods used as event handlers. Defaults to `handle`. Set to `false` to disable handler-name checks.
- `eventHandlerPropPrefix`: Prefix for props that are used as event handlers. Defaults to `on`. Set to `false` to disable prop-key checks.
- `checkLocalVariables`: Determines whether event handlers stored as local variables are checked. Defaults to `false`.
- `checkInlineFunction`: Determines whether event handlers set as inline functions are checked. Defaults to `false`.
- `ignoreComponentNames`: Array of glob strings, when matched with component name, ignores the rule on that component. Defaults to `[]`. Supports namespaced component names (e.g., `A.TestComponent`, `A.MyLib*`).

Examples of **incorrect** code for this rule with `{ "checkLocalVariables": true }`:

```json
{ "react/jsx-handler-names": ["error", { "checkLocalVariables": true }] }
```

```jsx
<MyComponent onChange={takeCareOfChange} />
```

Examples of **correct** code for this rule with `{ "checkLocalVariables": true }`:

```json
{ "react/jsx-handler-names": ["error", { "checkLocalVariables": true }] }
```

```jsx
<MyComponent onChange={handleChange} />
```

Examples of **incorrect** code for this rule with `{ "checkInlineFunction": true }`:

```json
{ "react/jsx-handler-names": ["error", { "checkInlineFunction": true }] }
```

```jsx
<MyComponent onChange={() => this.takeCareOfChange()} />
```

Examples of **correct** code for this rule with `{ "checkInlineFunction": true }`:

```json
{ "react/jsx-handler-names": ["error", { "checkInlineFunction": true }] }
```

```jsx
<MyComponent onChange={() => this.handleChange()} />
```

Examples of **correct** code for this rule with `{ "ignoreComponentNames": ["MyLib*"] }`:

```json
{
  "react/jsx-handler-names": [
    "error",
    { "checkLocalVariables": true, "ignoreComponentNames": ["MyLib*"] }
  ]
}
```

```jsx
<MyLibInput customPropNameBar={handleSomething} />
```

## When Not To Use It

If you are not using JSX, or if you don't want to enforce specific naming conventions for event handlers.

## Original Documentation

- [react/jsx-handler-names](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-handler-names.md)
