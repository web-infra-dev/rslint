# arrow-body-style

## Rule Details

This rule enforces or disallows the use of braces around arrow function bodies.

Arrow functions have two syntactic forms for their function bodies. They may be defined with a _block_ body (denoted with curly braces) `() => { ... }` or with a single expression `() => ...`, whose value is implicitly returned.

This rule has a string option and an object option.

The string option is one of:

- `"as-needed"` (default) enforces no braces where they can be omitted.
- `"always"` enforces braces around the function body.
- `"never"` enforces no braces around the function body (forbids any use of braces).

The object option (only available with `"as-needed"`):

- `requireReturnForObjectLiteral: true` requires braces and an explicit return for object literals. Default is `false`.

### `"as-needed"`

Examples of **incorrect** code for this rule with the default `"as-needed"` option:

```javascript
let foo = () => {
  return 0;
};
let foo = () => {
  return {
    bar: {
      foo: 1,
      bar: 2,
    },
  };
};
```

Examples of **correct** code for this rule with the default `"as-needed"` option:

```javascript
let foo = () => 0;
let foo = () => ({ bar: { foo: 1, bar: 2 } });
let foo = () => {
  let retVal = 0;
  return retVal;
};
let foo = () => {
  /* do nothing */
};
let foo = () => {
  // do nothing.
};
let foo = () => ({ bar: 0 });
```

### `requireReturnForObjectLiteral`

Examples of **incorrect** code for this rule with the `{ "requireReturnForObjectLiteral": true }` option:

```json
{ "arrow-body-style": ["error", "as-needed", { "requireReturnForObjectLiteral": true }] }
```

```javascript
let foo = () => ({});
let foo = () => ({ bar: 0 });
```

Examples of **correct** code for this rule with the `{ "requireReturnForObjectLiteral": true }` option:

```json
{ "arrow-body-style": ["error", "as-needed", { "requireReturnForObjectLiteral": true }] }
```

```javascript
let foo = () => {};
let foo = () => {
  return { bar: 0 };
};
```

### `"always"`

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "arrow-body-style": ["error", "always"] }
```

```javascript
let foo = () => 0;
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "arrow-body-style": ["error", "always"] }
```

```javascript
let foo = () => {
  return 0;
};
```

### `"never"`

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "arrow-body-style": ["error", "never"] }
```

```javascript
let foo = () => {
  return 0;
};
let foo = (data, name) => {
  data[name] = true;
  return data;
};
```

Examples of **correct** code for this rule with the `"never"` option:

```json
{ "arrow-body-style": ["error", "never"] }
```

```javascript
let foo = () => 0;
let foo = () => ({ foo: 0 });
```

## Differences from ESLint

- When an arrow function with a block body is immediately followed on the next line (no semicolon in between) by a token starting with `/` — e.g. `() => { return x }` then `/re/.test(y)` — rslint reports the rule but does not auto-fix it; ESLint removes the braces.

## Original Documentation

- [ESLint arrow-body-style](https://eslint.org/docs/latest/rules/arrow-body-style)
