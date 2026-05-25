# jsx-equals-spacing

Enforce or disallow spaces around the `=` sign in JSX attributes.

## Rule Details

This rule normalizes the whitespace around the `=` sign in JSX attributes. Spread attributes (`{...props}`) and valueless attributes (`<App foo />`) are never checked.

## Options

This rule has a string option:

- `"never"` (default) — disallow spaces on either side of `=`.
- `"always"` — require one space on each side of `=`.

### never

Examples of **incorrect** code for this rule with the default `"never"` option:

```javascript
<Hello name = {firstName} />;
<Hello name ={firstName} />;
<Hello name= {firstName} />;
```

Examples of **correct** code for this rule with the default `"never"` option:

```javascript
<Hello name={firstName} />;
<Hello name />;
<Hello {...props} />;
```

### always

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-equals-spacing": ["error", "always"] }
```

```javascript
<Hello name={firstName} />;
<Hello name ={firstName} />;
<Hello name= {firstName} />;
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/jsx-equals-spacing": ["error", "always"] }
```

```javascript
<Hello name = {firstName} />;
<Hello name />;
<Hello {...props} />;
```

## Original Documentation

- [@stylistic/jsx-equals-spacing](https://eslint.style/rules/jsx-equals-spacing)
