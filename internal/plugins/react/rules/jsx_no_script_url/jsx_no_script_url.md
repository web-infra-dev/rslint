# react/jsx-no-script-url

## Rule Details

Disallow usage of `javascript:` URLs in JSX attributes. In React 16.9, any URLs starting with `javascript:` log a warning. In a future major release, React will throw an error if it encounters a `javascript:` URL.

By default, this rule checks the `href` attribute of the `<a>` element. Additional components and attributes can be configured via options.

Examples of **incorrect** code for this rule:

```jsx
<a href="javascript:"></a>
<a href="javascript:void(0)"></a>
<a href="j
a
v	ascript:"></a>
```

Examples of **correct** code for this rule:

```jsx
<a href="https://reactjs.org"></a>
<a href="mailto:foo@bar.com"></a>
<a href="#"></a>
<a href={"javascript:"}></a>
<Foo href="javascript:"></Foo>
```

## Options

This rule accepts an optional array of custom component configurations and an optional object option.

### Array option (default `[]`)

```json
{ "react/jsx-no-script-url": ["error", [{ "name": "Link", "props": ["to"] }, { "name": "Foo", "props": ["href", "to"] }]] }
```

Allows you to indicate a specific list of properties used by a custom component to be checked.

Examples of **incorrect** code with the above options:

```jsx
<Link to="javascript:void(0)"></Link>
<Foo href="javascript:void(0)"></Foo>
<Foo to="javascript:void(0)"></Foo>
```

### Object option

#### `includeFromSettings` (default `false`)

Indicates if the `linkComponents` config in shared settings should also be taken into account. If enabled, components and properties defined in settings will be added to the list provided in the first option (if provided).

```json
{ "react/jsx-no-script-url": ["error", [{ "name": "Link", "props": ["to"] }], { "includeFromSettings": true }] }
```

If only settings should be used, the array option can be omitted:

```json
{ "react/jsx-no-script-url": ["error", { "includeFromSettings": true }] }
```

## Original Documentation

- [eslint-plugin-react/jsx-no-script-url](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-script-url.md)
