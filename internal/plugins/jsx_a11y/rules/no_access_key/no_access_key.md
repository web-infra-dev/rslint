# no-access-key

## Rule Details

This rule disallows the `accessKey` prop on JSX elements. The attribute
name is matched case-insensitively, so `accesskey`, `accessKey`, and
`acCesSKeY` are all checked. Access-key shortcuts assigned through this
prop frequently conflict with the keyboard commands used by screen
readers and keyboard-only users, causing accessibility regressions.

The rule reports an element whenever it carries an `accessKey` attribute
whose value resolves to a truthy JavaScript value. An attribute whose
value is `undefined`, `null`, `false`, `0`, or the empty string is left
alone.

This rule takes no arguments.

Examples of **incorrect** code for this rule:

```jsx
<div accessKey="h" />
<div accesskey="h" />
<div accessKey="h" {...props} />
<div acCesSKeY="y" />
<div accessKey={"y"} />
<div accessKey={`${y}`} />
<div accessKey={accessKey} />
```

Examples of **correct** code for this rule:

```jsx
<div />
<div {...props} />
<div accessKey={undefined} />
```

## Resources

- [WebAIM — Keyboard Accessibility: Accesskey](https://webaim.org/techniques/keyboard/accesskey#spec)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-access-key](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-access-key.md)
