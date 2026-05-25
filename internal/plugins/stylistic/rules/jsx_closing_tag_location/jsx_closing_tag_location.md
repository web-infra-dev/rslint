# jsx-closing-tag-location

Enforce closing tag location for multiline JSX.

## Rule Details

This rule checks the location of the closing tag (`</Foo>`, or `</>` for fragments) of a multiline JSX element. A single-line element — whose opening and closing tags share a line — is always allowed. When the element spans multiple lines, the closing tag must be on its own line and aligned according to the chosen option; otherwise the rule reports it and the autofix re-indents the closing tag (or moves it onto its own line).

## Options

This rule has a string option:

- `"tag-aligned"` (default) — closing tag aligned with the column of the opening `<`.
- `"line-aligned"` — closing tag aligned with the indentation of the line containing the opening tag (useful when the opening tag does not start its line, e.g. `const App = <Bar>`).

### tag-aligned

Examples of **incorrect** code for this rule with the default `"tag-aligned"` option:

```javascript
<Say
  firstName="John"
  lastName="Smith">
  Hello
    </Say>;
```

Examples of **correct** code for this rule with the default `"tag-aligned"` option:

```javascript
<Say
  firstName="John"
  lastName="Smith">
  Hello
</Say>;
```

### line-aligned

Examples of **incorrect** code for this rule with the `"line-aligned"` option:

```json
{ "@stylistic/jsx-closing-tag-location": ["error", "line-aligned"] }
```

```javascript
const App = <Bar>
  Foo
                </Bar>;
```

Examples of **correct** code for this rule with the `"line-aligned"` option:

```json
{ "@stylistic/jsx-closing-tag-location": ["error", "line-aligned"] }
```

```javascript
const App = <Bar>
  Foo
</Bar>;
```

## Original Documentation

- [@stylistic/jsx-closing-tag-location](https://eslint.style/rules/jsx-closing-tag-location)
