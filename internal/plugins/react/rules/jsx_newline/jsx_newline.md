# jsx-newline

## Rule Details

Require or prevent blank lines between JSX elements and expressions.

By default, this rule requires a blank line between sibling elements or
expressions separated by JSX text. It also checks children of fragments.
Block comments written as `{/* comment */}` stay with the following child.

Examples of **incorrect** code for this rule:

```jsx
<div>
  <Button />
  <List />
</div>
```

Examples of **correct** code for this rule:

```jsx
<div>
  <Button />

  <List />
</div>
```

The rule provides automatic fixes. With `prevent: true`, blank lines containing
indentation or Windows (CRLF) line endings may need to be removed manually.
Children separated only by a space may also need a manual line break. These
limitations match the upstream rule.

## Options

The rule accepts an object with two boolean options, both defaulting to `false`:

- `prevent`: disallow blank lines between adjacent elements and expressions.
- `allowMultilines`: require blank lines when either adjacent element or
  expression spans multiple lines, while preventing blank lines between
  single-line children. This option requires `prevent: true`.

```json
{
  "react/jsx-newline": ["error", { "prevent": true, "allowMultilines": true }]
}
```

Examples of **correct** code with these options:

```jsx
<div>
  <Button />
  <List />

  {/* Keep the comment with the multiline child. */}
  <Panel>
    <Content />
  </Panel>

  <Footer />
</div>
```

## When Not To Use It

Disable this rule if you do not want to enforce blank lines between JSX children.

## Differences from Upstream

When migrating from ESLint configured with `@typescript-eslint/parser`, numeric
character references above U+FFFF in JSX text can produce different diagnostics.
For example, rslint accepts `<><A/>&#65546;&#65546;<B/></>` with the default options,
matching ESLint's default parser. With `@typescript-eslint/parser`, upstream
reports a missing blank line for this example.

## Original Documentation

- [eslint-plugin-react: jsx-newline](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-newline.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-newline.js)
