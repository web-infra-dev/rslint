# jsx-props-no-spreading

Disallow JSX prop spreading.

## Rule Details

Enforces that JSX props are passed explicitly. This improves readability and
helps avoid passing unintended props to components and HTML elements.

Examples of **incorrect** code for this rule:

```jsx
<App {...props} />
<MyCustomComponent {...props} some_other_prop={some_other_prop} />
<img {...props} />
```

Examples of **correct** code for this rule:

```jsx
const {src, alt} = props;
<MyCustomComponent src={src} alt={alt} />
<img src={src} alt={alt} />
```

## Rule Options

The default configuration enforces prop spreading on HTML tags, custom
components, and non-explicit object spreads.

### `html`

`"ignore"` allows prop spreading on lowercase HTML tags. The default is
`"enforce"`.

Examples of **correct** code for this rule with `{ "html": "ignore" }`:

```json
{ "react/jsx-props-no-spreading": ["error", { "html": "ignore" }] }
```

```jsx
<img {...props} />
```

Examples of **incorrect** code for this rule with `{ "html": "ignore" }`:

```jsx
<App {...props} />
```

### `custom`

`"ignore"` allows prop spreading on custom components. The default is
`"enforce"`.

Examples of **correct** code for this rule with `{ "custom": "ignore" }`:

```json
{ "react/jsx-props-no-spreading": ["error", { "custom": "ignore" }] }
```

```jsx
<MyComponent {...props} />
```

Examples of **incorrect** code for this rule with `{ "custom": "ignore" }`:

```jsx
<img {...props} />
```

### `explicitSpread`

`"ignore"` allows a spread of an object containing only explicitly listed
properties. The default is `"enforce"`.

Examples of **correct** code for this rule with `{ "explicitSpread": "ignore" }`:

```json
{ "react/jsx-props-no-spreading": ["error", { "explicitSpread": "ignore" }] }
```

```jsx
<App {...{ prop1, prop2, prop3 }} />
```

An object containing another spread, a conditional expression, or any other
non-object expression is still reported.

### `exceptions`

`exceptions` is an array of tag names whose HTML or custom-component setting is
inverted. Exceptions use the complete JSX tag name, including member
components such as `components.Group`.

```json
{ "react/jsx-props-no-spreading": ["error", { "exceptions": ["Image", "img"] }] }
```

```jsx
<Image {...props} />
<img {...props} />
```

Examples of **incorrect** code for this rule with `{ "exceptions": ["Image", "img"] }`:

```jsx
<MyComponent {...props} />
```

The `html` and `custom` options cannot both be `"ignore"` when `exceptions` is
empty, matching the upstream schema.

## When Not To Use It

If your project intentionally uses prop spreading, or if spreading is the
preferred interface for higher-order components, leave this rule disabled.

## Original Documentation

- [eslint-plugin-react: jsx-props-no-spreading](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-props-no-spreading.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-props-no-spreading.js)
