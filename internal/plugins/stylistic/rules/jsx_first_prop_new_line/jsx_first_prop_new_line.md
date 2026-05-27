# jsx-first-prop-new-line

Enforce the position of the first prop in a JSX element.

## Rule Details

This rule checks whether the first prop of a JSX element sits on the same line as the element name or on its own line, according to the selected option. It is useful for enforcing consistent formatting of JSX props.

Examples of **incorrect** code for this rule with the default `"multiline-multiprop"` option:

```jsx
<Hello foo="bar"
  baz="quux"
/>
```

Examples of **correct** code for this rule with the default `"multiline-multiprop"` option:

```jsx
<Hello foo="bar" baz="quux" />
<Hello
  foo="bar"
  baz="quux"
/>
```

## Options

This rule has a string option:

- `"multiline-multiprop"` (default) — the first prop must be on a new line when the JSX tag spans multiple lines and has more than one prop.
- `"always"` — the first prop must always be on a new line.
- `"never"` — the first prop must never be on a new line.
- `"multiline"` — the first prop must be on a new line when the JSX tag spans multiple lines.
- `"multiprop"` — the first prop must be on a new line when the JSX tag has more than one prop.

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-first-prop-new-line": ["error", "always"] }
```

```jsx
<Hello personal="stuff" />
```

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "@stylistic/jsx-first-prop-new-line": ["error", "never"] }
```

```jsx
<Hello
  personal="stuff"
/>
```

## Differences from ESLint

These only affect autofix in rare shapes; the reported diagnostics (positions, messages, and when they fire) match ESLint exactly.

- When a comment sits between the tag name and the first prop and the fix moves the prop to a new line, rslint keeps the comment; ESLint removes it.
- When a comment sits between the tag name and the first prop and the fix would move the prop back onto the tag's line, rslint reports the issue without an autofix (the prop cannot be moved across the comment); ESLint removes the comment and collapses the line.
- For a generic element whose first prop is moved back onto the tag's line, rslint keeps the type arguments (`<Foo<T>`); ESLint drops them.

## Original Documentation

- [@stylistic/jsx-first-prop-new-line](https://eslint.style/rules/jsx-first-prop-new-line)
