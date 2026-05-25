# eol-last

Require or disallow newline at the end of files.

## Rule Details

This rule enforces at least one newline (or the absence thereof) at the end of non-empty files.

Trailing newlines in non-empty files are a common UNIX idiom. Benefits include the ability to concatenate or append to files and to output files to the terminal without interfering with shell prompts.

## Options

This rule has a string option:

- `"always"` (default) requires the file to end with at least one newline (`\n` or `\r\n`).
- `"never"` disallows any trailing newline at the end of the file.

### always

Examples of **incorrect** code for this rule with the default `"always"` option:

```javascript
function doSomething() {
  var foo = 2;
}
```

Examples of **correct** code for this rule with the default `"always"` option:

```javascript
function doSomething() {
  var foo = 2;
}

```

### never

Examples of **incorrect** code for this rule with the `"never"` option:

```json
{ "@stylistic/eol-last": ["error", "never"] }
```

```javascript
function doSomething() {
  var foo = 2;
}

```

Examples of **correct** code for this rule with the `"never"` option:

```json
{ "@stylistic/eol-last": ["error", "never"] }
```

```javascript
function doSomething() {
  var foo = 2;
}
```

## Original Documentation

- [@stylistic/eol-last](https://eslint.style/rules/eol-last)
