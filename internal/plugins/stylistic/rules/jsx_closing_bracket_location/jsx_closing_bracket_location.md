# jsx-closing-bracket-location

Enforce closing bracket location in JSX.

## Rule Details

This rule checks the location of the closing bracket (`>` for opening tags, `/>` for self-closing tags) of a multiline JSX element. When the opening tag's last attribute (or, in the zero-attribute form, the tag name) is followed by a trailing comment inside the opening element, the configured `after-props` / `after-tag` location is upgraded to `line-aligned` so the fix does not move the comment onto the bracket's line.

## Options

This rule has a string option:

- `"tag-aligned"` (default) — closing bracket aligned with the column of the opening `<`.
- `"line-aligned"` — closing bracket aligned with the indentation of the line containing the opening tag.
- `"props-aligned"` — closing bracket aligned with the column of the last attribute.
- `"after-props"` — closing bracket placed immediately after the last attribute (no leading whitespace).

This rule has an object option:

- `"location"` — string value (any of the four above). Sets the same location for both self-closing and non-self-closing forms. When present, `nonEmpty` and `selfClosing` are ignored.
- `"nonEmpty"` — string value or `false`. Location for non-self-closing tags (`<Foo>…</Foo>`). `false` disables the rule for this form.
- `"selfClosing"` — string value or `false`. Location for self-closing tags (`<Foo />`). `false` disables the rule for this form.

### tag-aligned

Examples of **incorrect** code for this rule with the default `"tag-aligned"` option:

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
    />;

var x = <Hello
  firstName="John"
  lastName="Smith"
  />;
```

Examples of **correct** code for this rule with the default `"tag-aligned"` option:

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
/>;

var x = <Hello firstName="John" lastName="Smith" />;
```

### line-aligned

Examples of **incorrect** code for this rule with the `"line-aligned"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "line-aligned"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
    />;
```

Examples of **correct** code for this rule with the `"line-aligned"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "line-aligned"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
/>;
```

### props-aligned

Examples of **incorrect** code for this rule with the `"props-aligned"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "props-aligned"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
/>;
```

Examples of **correct** code for this rule with the `"props-aligned"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "props-aligned"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
  />;
```

### after-props

Examples of **incorrect** code for this rule with the `"after-props"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "after-props"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"
/>;
```

Examples of **correct** code for this rule with the `"after-props"` option:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", "after-props"] }
```

```javascript
var x = <Hello
  firstName="John"
  lastName="Smith"/>;
```

### nonEmpty

Examples of **correct** code for this rule with the `{ "nonEmpty": "after-props" }` option (the closing bracket of non-self-closing tags must sit after the last prop; self-closing tags keep the default `"tag-aligned"`):

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", { "nonEmpty": "after-props" }] }
```

```javascript
<Provider
  store>
  <App
    foo
  />
</Provider>
```

Setting `"nonEmpty": false` disables the rule for non-self-closing tags entirely:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", { "nonEmpty": false }] }
```

### selfClosing

Examples of **correct** code for this rule with the `{ "selfClosing": "after-props" }` option (the closing bracket of self-closing tags must sit after the last prop; non-self-closing tags keep the default `"tag-aligned"`):

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", { "selfClosing": "after-props" }] }
```

```javascript
<Provider store>
  <App
    foo />
</Provider>
```

Setting `"selfClosing": false` disables the rule for self-closing tags entirely:

```json
{ "@stylistic/jsx-closing-bracket-location": ["error", { "selfClosing": false }] }
```

## Original Documentation

- [@stylistic/jsx-closing-bracket-location](https://eslint.style/rules/jsx-closing-bracket-location)
