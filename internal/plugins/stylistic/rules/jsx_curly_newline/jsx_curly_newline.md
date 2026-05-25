# jsx-curly-newline

Enforce consistent linebreaks inside the curly braces of JSX expression containers (`{ ... }` in attribute values and children).

## Rule Details

For each JSX expression container the rule checks two positions: right after the opening `{` and right before the closing `}`. Whether a newline is required at those positions is controlled by the option, chosen by whether the contained expression spans a single line or multiple lines.

The rule only manages the linebreaks themselves; it does not adjust indentation (that is the job of an indentation rule).

## Options

This rule accepts either a string or an object.

- `"consistent"` (default) — the closing brace's linebreak must match the opening brace's: if there is a newline after `{`, there must be one before `}`, and vice versa.
- `"never"` — no linebreaks are allowed after `{` or before `}`.
- An object `{ "singleline": ..., "multiline": ... }`, where each property is one of:
  - `"consistent"` (default for each property) — match the opening brace, as above.
  - `"require"` — a newline is required after `{` and before `}`.
  - `"forbid"` — newlines after `{` and before `}` are not allowed.

  `singleline` applies when the expression is on a single line; `multiline` applies when it spans multiple lines.

Examples of **incorrect** code for this rule with the default `"consistent"` option:

```javascript
<div>
  { foo
  }
</div>
```

Examples of **correct** code for this rule with the default `"consistent"` option:

```javascript
<div>{foo}</div>;

<div>
  {
    foo
  }
</div>;

<div foo={
  bar
} />;
```

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "@stylistic/jsx-curly-newline": ["error", "never"] }
```

```javascript
<div>
  {
    foo
  }
</div>
```

Examples of **correct** code for this rule with the `"never"` option:

```json
{ "@stylistic/jsx-curly-newline": ["error", "never"] }
```

```javascript
<div>{foo}</div>;

<div>
  { foo &&
    foo.bar }
</div>;
```

Examples of **incorrect** code for this rule with the `{ "singleline": "consistent", "multiline": "require" }` option:

```json
{ "@stylistic/jsx-curly-newline": ["error", { "singleline": "consistent", "multiline": "require" }] }
```

```javascript
<div>
  { foo &&
    bar }
</div>
```

Examples of **correct** code for this rule with the `{ "singleline": "consistent", "multiline": "require" }` option:

```json
{ "@stylistic/jsx-curly-newline": ["error", { "singleline": "consistent", "multiline": "require" }] }
```

```javascript
<div>{foo}</div>;

<div>
  {
    foo &&
    bar
  }
</div>;
```

## Original Documentation

- [@stylistic/jsx-curly-newline](https://eslint.style/rules/jsx-curly-newline)
