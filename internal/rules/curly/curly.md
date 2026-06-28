# curly

## Rule Details

This rule enforces consistent use of curly braces around blocks that follow
`if`, `else`, `for`, `for...in`, `for...of`, `while`, and `do...while`
statements. By default it requires braces everywhere, but it can also be
configured to forbid them where they are unnecessary.

Examples of **incorrect** code for this rule:

```javascript
if (foo) foo++;

while (bar) baz();

if (foo) {
  baz();
} else qux();
```

Examples of **correct** code for this rule:

```javascript
if (foo) {
  foo++;
}

while (bar) {
  baz();
}

if (foo) {
  baz();
} else {
  qux();
}
```

## Options

The rule accepts a primary string option and an optional secondary
`"consistent"` modifier: `["error", "multi", "consistent"]`.

### `"all"` (default)

Requires braces around every block.

### `"multi"`

Forbids braces around blocks that contain a single statement, and requires them
when the block contains two or more statements.

```json
{ "curly": ["error", "multi"] }
```

Examples of **incorrect** code for `"multi"`:

```javascript
if (foo) {
  foo++;
}

for (var i = 0; foo; i++) {
  doSomething();
}
```

Examples of **correct** code for `"multi"`:

```javascript
if (foo) foo++;

while (true) {
  doSomething();
  doSomethingElse();
}
```

### `"multi-line"`

Allows brace-less single-line statements, but requires braces once a statement
spans multiple lines.

```json
{ "curly": ["error", "multi-line"] }
```

Examples of **correct** code for `"multi-line"`:

```javascript
if (foo) foo++;
else doSomething();

while (true) {
  doSomething();
  doSomethingElse();
}
```

### `"multi-or-nest"`

Forces brace-less syntax for a single-line statement, and requires braces for a
multi-line statement or a statement that contains a nested block.

```json
{ "curly": ["error", "multi-or-nest"] }
```

Examples of **correct** code for `"multi-or-nest"`:

```javascript
if (foo) bar();

if (foo) {
  bar();
  baz();
}
```

### `"consistent"`

Used together with `"multi"`, `"multi-line"`, or `"multi-or-nest"`, this option
forces all branches of an `if`/`else if`/`else` chain to agree: either all have
braces or none do.

```json
{ "curly": ["error", "multi", "consistent"] }
```

Examples of **correct** code for `["multi", "consistent"]`:

```javascript
if (foo) {
  bar();
} else {
  baz();
}

if (foo) bar();
else baz();
```

## Original Documentation

- [ESLint curly](https://eslint.org/docs/latest/rules/curly)
