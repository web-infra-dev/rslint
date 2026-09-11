# iframe-missing-sandbox

Require a `sandbox` attribute on `iframe` elements.

The `sandbox` attribute adds restrictions to the document inside an iframe. This rule also validates statically known sandbox tokens and rejects the unsafe combination of `allow-scripts` and `allow-same-origin`, which permits the embedded document to remove its sandbox.

Examples of **incorrect** code for this rule:

```jsx
<iframe />
<iframe sandbox="not-a-sandbox-token" />
<iframe sandbox="allow-scripts allow-same-origin" />
React.createElement('iframe')
```

Examples of **correct** code for this rule:

```jsx
<iframe sandbox="allow-popups" />
React.createElement('iframe', { sandbox: 'allow-forms' })
```

Dynamic sandbox values are permitted because they cannot be evaluated statically:

```jsx
<iframe sandbox={sandboxValue} />
React.createElement('iframe', { sandbox: sandboxValue })
```

## Rule Options

This rule has no options.

## Original Documentation

- [eslint-plugin-react: iframe-missing-sandbox](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/iframe-missing-sandbox.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/iframe-missing-sandbox.js)
