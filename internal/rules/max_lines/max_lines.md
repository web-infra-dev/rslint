# max-lines

Enforce a maximum number of lines per file.

## Rule Details

Large files tend to do a lot of things and can make it hard to follow what's
going on. This rule caps the number of lines in a file.

With the `{ "max": 3 }` option:

```json
{
  "rules": {
    "max-lines": ["error", 3]
  }
}
```

Examples of **incorrect** code for this rule:

```javascript
let a,
    b,
    c,
    d;
```

Examples of **correct** code for this rule:

```javascript
let a,
    b,
    c;
```

## Options

This rule accepts a number (the maximum allowed) or an object with the
following properties:

- `max` (default `300`): the maximum number of lines allowed in a file.
- `skipBlankLines` (default `false`): ignore lines made up purely of
  whitespace.
- `skipComments` (default `false`): ignore lines containing just comments
  (a comment that shares a line with code does not exclude that line).

### `skipBlankLines`

```json
{
  "rules": {
    "max-lines": ["error", { "max": 2, "skipBlankLines": true }]
  }
}
```

Examples of **correct** code with the above configuration:

```javascript
var a = 1;


var b = 2;
```

### `skipComments`

```json
{
  "rules": {
    "max-lines": ["error", { "max": 2, "skipComments": true }]
  }
}
```

Examples of **correct** code with the above configuration:

```javascript
// a header comment
var a = 1;
var b = 2;
```

## Original Documentation

- <https://eslint.org/docs/latest/rules/max-lines>
