# jsx-uses-vars

## Rule Details

Mark JSX component bindings as used by core `no-unused-vars`. For example,
`<Button />` marks `Button`, and `<UI.Button />` marks `UI` in the JSX element's
lexical scope. Lowercase standalone tags such as `<div />` and namespaced tags
such as `<svg:path />` do not mark variables.

Core `no-unused-vars` already tracks JSX component references, following
ESLint 10. This rule additionally marks bindings explicitly, which can matter
for otherwise-discarded self-references. Expressions inside JSX, such as
`<div>{value}</div>`, also count as references without this rule.

## Original Documentation

- [eslint-plugin-react: jsx-uses-vars](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-uses-vars.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-uses-vars.js)
