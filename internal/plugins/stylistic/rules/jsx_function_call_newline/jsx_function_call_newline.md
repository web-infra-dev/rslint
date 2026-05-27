# jsx-function-call-newline

Enforce line breaks before and after JSX elements when they are used as arguments to a function.

## Rule Details

This rule checks whether a line break is needed before and after every JSX element that serves as an argument to a function call or `new` expression. A trailing comma after the argument counts as a separator, so no line break is required between the element and that comma.

## Options

This rule has a string option:

- `"multiline"` (default) — a line break before and after a JSX argument is required only when the element itself spans multiple lines.
- `"always"` — a line break before and after every JSX argument is required.

### multiline

Examples of **incorrect** code for this rule with the default `"multiline"` option:

```jsx
fn(<div
 />, <span>
 foo</span>
)

fn (
  <div />, <span>
    bar
  </span>
)
```

Examples of **correct** code for this rule with the default `"multiline"` option:

```jsx
fn(<div />)

fn(<span>foo</span>)

fn(
  <div
 />,
 <span>
 foo</span>
)

fn (
  <div />,
  <span>
    bar
  </span>
)
```

### always

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-function-call-newline": ["error", "always"] }
```

```jsx
fn(<div />)

fn(<span>foo</span>)

fn(<span>
bar
</span>)

fn(<div />, <div
  style={{ color: 'red' }}
  />, <p>baz</p>)
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-function-call-newline": ["error", "always"] }
```

```jsx
fn(
  <div />
)

fn(
  <span>foo</span>
)

fn(
  <span>
    bar
  </span>
)

fn(
  <div />,
  <div
    style={{ color: 'red' }}
  />,
  <p>baz</p>
)
```

## Original Documentation

- [@stylistic/jsx-function-call-newline](https://eslint.style/rules/jsx-function-call-newline)
