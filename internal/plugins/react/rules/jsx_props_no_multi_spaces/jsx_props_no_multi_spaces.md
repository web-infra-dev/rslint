# jsx-props-no-multi-spaces

## Rule Details

Disallow multiple spaces between inline JSX props and blank lines between multiline JSX props.

Examples of **incorrect** code for this rule:

```jsx
<Foo  bar="baz" />
<Foo bar="baz"  bam="qux" />
```

Examples of **correct** code for this rule:

```jsx
<Foo bar="baz" />
<Foo bar="baz" bam="qux" />
<Foo
  bar="baz"
  bam="qux"
/>
```

## Original Documentation

- [eslint-plugin-react: jsx-props-no-multi-spaces](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-props-no-multi-spaces.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-props-no-multi-spaces.js)
