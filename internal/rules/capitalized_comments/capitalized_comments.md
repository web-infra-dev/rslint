# capitalized-comments

## Rule Details

This rule enforces a consistent style of comments across your codebase, specifically by either requiring or disallowing a capitalized letter as the first word character in a comment. This rule will not issue warnings when non-cased letters are used.

By default, this rule requires a non-lowercase letter at the beginning of comments.

Examples of **incorrect** code for this rule:

```javascript
// lowercase comment
```

Examples of **correct** code for this rule:

```javascript
// Capitalized comment

// 1. Non-letter at beginning of comment

// 丈 Non-Latin character at beginning of comment

/* istanbul ignore next */
/* jscs:enable */
/* jshint asi:true */
/* global foo */
/* globals foo */
/* exported myVar */
// https://github.com
```

## Options

This rule has two options: a string value `"always"` or `"never"` which determines whether capitalization of the first word of a comment should be required or forbidden, and optionally an object containing more configuration parameters for the rule.

Here are the supported object options:

- `ignorePattern`: A string representing a regular expression pattern of words that should be ignored by this rule. If the first word of a comment matches the pattern, this rule will not report that comment.
  - Note that the following words are always ignored by this rule: `["jscs", "jshint", "eslint", "istanbul", "global", "globals", "exported"]`.
- `ignoreInlineComments`: If this is `true`, the rule will not report on comments in the middle of code. By default, this is `false`.
- `ignoreConsecutiveComments`: If this is `true`, the rule will not report on a comment which violates the rule, as long as the comment immediately follows another comment. By default, this is `false`.

Here is an example configuration:

```json
{
  "capitalized-comments": [
    "error",
    "always",
    {
      "ignorePattern": "pragma|ignored",
      "ignoreInlineComments": true
    }
  ]
}
```

### `"always"`

Using the `"always"` option means that this rule will report any comments which start with a lowercase letter. This is the default configuration for this rule.

Configuration comments and comments which start with URLs are never reported.

Examples of **incorrect** code for this rule:

```javascript
/* eslint capitalized-comments: ["error", "always"] */

// lowercase comment
```

Examples of **correct** code for this rule:

```javascript
/* eslint capitalized-comments: ["error", "always"] */

// Capitalized comment
```

### `"never"`

Using the `"never"` option means that this rule will report any comments which start with an uppercase letter.

Examples of **incorrect** code with the `"never"` option:

```javascript
/* eslint capitalized-comments: ["error", "never"] */

// Capitalized comment
```

Examples of **correct** code with the `"never"` option:

```javascript
/* eslint capitalized-comments: ["error", "never"] */

// lowercase comment
```

### `ignorePattern`

The `ignorePattern` option takes a string value, which is used as a regular expression applied to the first word of a comment.

Examples of **correct** code with the `"ignorePattern"` option set to `"pragma"`:

```json
{ "capitalized-comments": ["error", "always", { "ignorePattern": "pragma" }] }
```

```javascript
function foo() {
  /* pragma wrap(true) */
}
```

### `ignoreInlineComments`

Setting the `ignoreInlineComments` option to `true` means that comments in the middle of code (with a token on the same line as the beginning of the comment, and another token on the same line as the end of the comment) will not be reported by this rule.

Examples of **correct** code with the `"ignoreInlineComments"` option set to `true`:

```json
{ "capitalized-comments": ["error", "always", { "ignoreInlineComments": true }] }
```

```javascript
function foo(/* ignored */ a) {}
```

The exemption applies to a block comment with a token on the same line as its start and a token on the same line as its end, which is what "in the middle of code" means here:

```javascript
foo(/* ignored */ bar);
foo(/* ignored
  still ignored */ bar);
```

### `ignoreConsecutiveComments`

If the `ignoreConsecutiveComments` option is set to `true`, then comments which otherwise violate the rule will not be reported as long as they immediately follow another comment. This can be applied more than once.

Examples of **correct** code with `ignoreConsecutiveComments` set to `true`:

```json
{ "capitalized-comments": ["error", "always", { "ignoreConsecutiveComments": true }] }
```

```javascript
foo();
// This comment is valid since it has the correct capitalization.
// this comment is ignored since it follows another comment,
// and this one as well because it follows yet another comment.
```

Examples of **incorrect** code with `ignoreConsecutiveComments` set to `true`:

```javascript
/* eslint capitalized-comments: ["error", "always", { "ignoreConsecutiveComments": true }] */

foo();
// this comment is invalid, but only on this line.
// this comment does NOT get reported, since it is a consecutive comment.
```

### Use Different Options for Line and Block Comments

If you wish to have a different configuration for line comments and block comments, you can do so by using two different object configurations (note that the capitalization option will be enforced consistently for line and block comments):

```json
{
  "capitalized-comments": [
    "error",
    "always",
    {
      "line": {
        "ignorePattern": "pragma|ignored"
      },
      "block": {
        "ignoreInlineComments": true,
        "ignorePattern": "ignored"
      }
    }
  ]
}
```

Examples of **incorrect** code with different line and block comment configuration:

```javascript
/* eslint capitalized-comments: ["error", "always", { "block": { "ignorePattern": "blockignore" } }] */

// capitalized line comment, this is incorrect, blockignore does not help here
/* lowercased block comment, this is incorrect too */
```

Examples of **correct** code with different line and block comment configuration:

```javascript
/* eslint capitalized-comments: ["error", "always", { "block": { "ignorePattern": "blockignore" } }] */

// Uppercase line comment, this is correct
/* blockignore lowercase block comment, this is correct due to ignorePattern */
```

## Differences from ESLint

- An `ignorePattern` that is not a valid regular expression never matches any comment, instead of throwing an error when the rule is configured.

## Original Documentation

- [ESLint: capitalized-comments](https://eslint.org/docs/latest/rules/capitalized-comments)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/capitalized-comments.js)
