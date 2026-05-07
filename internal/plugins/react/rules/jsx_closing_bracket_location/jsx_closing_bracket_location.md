# jsx-closing-bracket-location

## Rule Details

Enforces the closing bracket location for JSX multiline elements. By default the closing bracket must be aligned with the opening tag.

Examples of **incorrect** code for this rule:

```jsx
<Hello
  lastName="Smith"
  firstName="John" />;

<Hello
  lastName="Smith"
  firstName="John"
  />;
```

Examples of **correct** code for this rule:

```jsx
<Hello firstName="John" lastName="Smith" />;

<Hello
  firstName="John"
  lastName="Smith"
/>;
```

## Rule Options

There are two ways to configure this rule.

The first form is a string shortcut corresponding to the `location` values specified below. If omitted, it defaults to `"tag-aligned"`.

```json
{ "react/jsx-closing-bracket-location": ["error", "tag-aligned"] }
```

The second form allows you to distinguish between non-empty and self-closing tags. Both properties are optional, and both default to `"tag-aligned"`. You can also disable the rule for one particular type of tag by setting the value to `false`.

```json
{
  "react/jsx-closing-bracket-location": [
    "error",
    { "nonEmpty": "tag-aligned", "selfClosing": "tag-aligned" }
  ]
}
```

### `location`

Enforced location for the closing bracket.

- `tag-aligned`: must be aligned with the opening tag.
- `line-aligned`: must be aligned with the line containing the opening tag.
- `after-props`: must be placed right after the last prop.
- `props-aligned`: must be aligned with the last prop.

Defaults to `tag-aligned`.

For backward compatibility, you may pass an object `{ "location": <location> }` that is equivalent to the first string shortcut form.

Examples of **incorrect** code for this rule with `"tag-aligned"` (default) or `"line-aligned"`:

```jsx
<Hello
  firstName="John"
  lastName="Smith"
  />;

<Say
  firstName="John"
  lastName="Smith">
  Hello
</Say>;
```

Examples of **incorrect** code for this rule with `"after-props"`:

```json
{ "react/jsx-closing-bracket-location": ["error", "after-props"] }
```

```jsx
<Hello
  firstName="John"
  lastName="Smith" />;
```

Examples of **incorrect** code for this rule with `"props-aligned"`:

```json
{ "react/jsx-closing-bracket-location": ["error", "props-aligned"] }
```

```jsx
<Hello
  firstName="John"
  lastName="Smith"
  />;
```

Examples of **correct** code for this rule with `"tag-aligned"` (default) or `"line-aligned"`:

```jsx
<Hello
  firstName="John"
  lastName="Smith"
/>;

<Say
  firstName="John"
  lastName="Smith"
>
  Hello
</Say>;
```

Examples of **correct** code for this rule with `{ "selfClosing": "after-props" }`:

```json
{
  "react/jsx-closing-bracket-location": [
    "error",
    { "selfClosing": "after-props" }
  ]
}
```

```jsx
<Hello
  firstName="John"
  lastName="Smith" />;

<Say
  firstName="John"
  lastName="Smith"
>
  Hello
</Say>;
```

Examples of **correct** code for this rule with `{ "selfClosing": "props-aligned", "nonEmpty": "after-props" }`:

```json
{
  "react/jsx-closing-bracket-location": [
    "error",
    { "selfClosing": "props-aligned", "nonEmpty": "after-props" }
  ]
}
```

```jsx
<Hello
  firstName="John"
  lastName="Smith"
  />;

<Say
  firstName="John"
  lastName="Smith">
  Hello
</Say>;
```

## When Not To Use It

If you are not using JSX, you can disable this rule.

## Original Documentation

- [react/jsx-closing-bracket-location](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-closing-bracket-location.md)
