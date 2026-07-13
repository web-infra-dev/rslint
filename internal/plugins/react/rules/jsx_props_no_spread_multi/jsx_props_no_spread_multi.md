# jsx-props-no-spread-multi

## Rule Details

Disallow spreading the same JSX props identifier multiple times in one element.

Examples of **incorrect** code for this rule:

```jsx
<App {...props} myAttr="1" {...props} />
```

Examples of **correct** code for this rule:

```jsx
<App myAttr="1" {...props} />
<App {...props} myAttr="1" />
```

## Options

This rule has no options.

## Original Documentation

- [react/jsx-props-no-spread-multi](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-props-no-spread-multi.md)
