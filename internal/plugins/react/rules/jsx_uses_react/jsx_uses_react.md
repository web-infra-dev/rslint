# jsx-uses-react

## Rule Details

Prevent React from being incorrectly marked as unused when JSX is used. This
rule is a no-op for configuration compatibility because native `no-unused-vars`
handles implicit factory references for classic JSX directly, without requiring
type information.

Select classic JSX with `settings.react.runtime: "classic"` or the project's
`jsx: "react"` compiler option. Configuration precedence is file pragmas, then
explicit Rslint settings, then project compiler options, then default factories.
File `@jsxRuntime` overrides the settings runtime; an explicit settings runtime
overrides the project's JSX mode in either direction.

Custom factories can be specified with file `@jsx` / `@jsxFrag` pragmas,
`settings.react.pragma` / `settings.react.fragment`, or project `jsxFactory` /
`jsxFragmentFactory`, in that order. A factory or fragment declaration alone
selects classic JSX only when no explicit runtime has been selected. Automatic
JSX does not use the classic factories, even if those factories remain configured.
A fragment factory is read only for shorthand fragments. File pragmas follow
TypeScript's leading block-comment syntax.

Enabling this rule alone does not select classic JSX.

## Original Documentation

- [eslint-plugin-react: jsx-uses-react](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-uses-react.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-uses-react.js)
