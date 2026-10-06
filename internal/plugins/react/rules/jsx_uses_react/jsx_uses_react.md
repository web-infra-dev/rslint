# jsx-uses-react

## Rule Details

Mark the JSX pragma binding as used by core `no-unused-vars` whenever JSX
appears. The default pragma is `React`; an `@jsx` comment takes precedence over
`settings.react.pragma`. For shorthand fragments (`<></>`), this rule also marks
`settings.react.fragment`, which defaults to `Fragment`.

Marks apply to the binding visible at each JSX expression. They work with both
classic and automatic JSX runtimes and do not depend on TypeScript's `jsx`
compiler option. Enable `react/jsx-uses-vars` as well to mark component tags.

## Original Documentation

- [eslint-plugin-react: jsx-uses-react](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-uses-react.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-uses-react.js)
