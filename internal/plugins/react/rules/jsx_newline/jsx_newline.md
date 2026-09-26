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

rslint interprets numeric character references in JSX text as Unicode code
points. Compared with ESLint's default parser, Espree:

- `&#65546;` represents U+1000A, not a line feed. With the default options,
  rslint reports a missing blank line for `<><A/>&#65546;&#65546;<B/></>`;
  Espree treats the references as a blank line and accepts it.
- Leading zeros do not prevent decoding. For example, two adjacent
  `&#00000000010;` references count as a blank line in rslint, while Espree
  leaves this spelling as text and reports a missing blank line.

These cases match ESLint configured with `@typescript-eslint/parser`.

## Original Documentation

- [eslint-plugin-react: jsx-newline](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-newline.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-newline.js)
