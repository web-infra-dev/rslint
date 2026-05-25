# arrow-parens

Require parentheses around arrow function arguments.

## Rule Details

This rule enforces parentheses around the parameter of arrow functions that take a single argument. Multi-parameter and zero-parameter arrows are never reported — JavaScript syntactically requires parentheses around them, so the rule has no decision to make.

The default option is `"always"`, which requires parentheses around every single-parameter arrow.

## Options

This rule has a string option:

- `"always"` (default) requires parentheses around every single-parameter arrow.
- `"as-needed"` removes parentheses when the parameter is a plain identifier with no annotation, default, optional marker, rest, or comments.

This rule has an object option, only meaningful with `"as-needed"`:

- `"requireForBlockBody": true` reverses the requirement when the arrow's body is a block — `a => { ... }` becomes invalid, `(a) => { ... }` becomes correct.

### always

Examples of **incorrect** code for this rule with the default `"always"` option:

```javascript
a => {};
a => a;
a.then(foo => {});
```

Examples of **correct** code for this rule with the default `"always"` option:

```javascript
() => {};
(a) => {};
(a) => a;
a.then((foo) => {});
```

### as-needed

Examples of **incorrect** code for this rule with the `"as-needed"` option:

```json
{ "@stylistic/arrow-parens": ["error", "as-needed"] }
```

```javascript
(a) => a;
```

Examples of **correct** code for this rule with the `"as-needed"` option:

```json
{ "@stylistic/arrow-parens": ["error", "as-needed"] }
```

```javascript
() => {};
a => {};
a => a;
([a, b]) => a;
({ a, b }) => a;
(a = 10) => a;
(...a) => a[0];
(a, b) => a + b;
```

The exceptions that keep parentheses under `"as-needed"`:

- Destructured parameters (`([a]) => a`, `({a}) => a`).
- A parameter with a default value (`(a = 1) => a`).
- A rest parameter (`(...a) => a`).
- A TypeScript optional parameter (`(a?: number) => a`).
- A parameter with a type annotation (`(a: number) => a`).
- An arrow with a return type annotation (`(a): number => a`).
- An arrow with generic type parameters (`<T>(a) => a`).
- A comment between the parentheses (`(/* doc */ a) => a`).

### requireForBlockBody

Examples of **incorrect** code for this rule with `"as-needed", { "requireForBlockBody": true }`:

```json
{ "@stylistic/arrow-parens": ["error", "as-needed", { "requireForBlockBody": true }] }
```

```javascript
(a) => a;
a => { return a; };
```

Examples of **correct** code for this rule with `"as-needed", { "requireForBlockBody": true }`:

```json
{ "@stylistic/arrow-parens": ["error", "as-needed", { "requireForBlockBody": true }] }
```

```javascript
a => a;
(a) => { return a; };
a => ({});
```

## Original Documentation

- [@stylistic/arrow-parens](https://eslint.style/rules/arrow-parens)
