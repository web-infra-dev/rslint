# no-warning-comments

## Rule Details

This rule reports comments that include any of the predefined terms specified in its configuration.

```javascript
// TODO: do something
// FIXME: this is not a good idea
```

## Options

This rule has an options object literal:

- `"terms"`: optional array of terms to match. Defaults to `["todo", "fixme", "xxx"]`. Terms are matched case-insensitively and as whole words: `fix` would match `FIX` but not `fixing`. Terms can consist of multiple words: `really bad idea`.
- `"location"`: optional string that configures where in your comments to check for matches. Defaults to `"start"`. The start is from the first non-decorative character, ignoring whitespace, new lines and characters specified in `decoration`. The other value is match `anywhere` in comments.
- `"decoration"`: optional array of characters that are ignored at the start of a comment, when location is `"start"`. Defaults to `[]`. Any sequence of whitespace or the characters from this property are ignored. This option is ignored when location is `"anywhere"`.

Examples of **incorrect** code for this rule with the default `{ "terms": ["todo", "fixme", "xxx"], "location": "start" }` options:

```javascript
/*
FIXME
*/
function callback(err, results) {
  if (err) {
    console.error(err);
    return;
  }
  // TODO
}
```

Examples of **correct** code for this rule with the default `{ "terms": ["todo", "fixme", "xxx"], "location": "start" }` options:

```javascript
function callback(err, results) {
  if (err) {
    console.error(err);
    return;
  }
  // NOT READY FOR PRIME TIME
  // but too bad, it is not a predefined warning term
}
```

### terms and location

Examples of **incorrect** code for the `{ "terms": ["todo", "fixme", "any other term"], "location": "anywhere" }` options:

```json
{
  "no-warning-comments": [
    "error",
    { "terms": ["todo", "fixme", "any other term"], "location": "anywhere" }
  ]
}
```

```javascript
// TODO: this
// todo: this too
// Even this: TODO
/*
 * The same goes for this TODO comment
 * Or a fixme
 * as well as any other term
 */
```

Examples of **correct** code for the `{ "terms": ["todo", "fixme", "any other term"], "location": "anywhere" }` options:

```javascript
// This is to do
// even not any other    term
// any other terminal
/*
 * The same goes for block comments
 * with any other interesting term
 * or fix me this
 */
```

### Decoration Characters

Examples of **incorrect** code for the `{ "decoration": ["*"] }` options:

```json
{ "no-warning-comments": ["error", { "decoration": ["*"] }] }
```

```javascript
//***** todo decorative asterisks are ignored *****//
/**
 * TODO new lines and asterisks are also ignored in block comments.
 */
```

Examples of **incorrect** code for the `{ "decoration": ["/", "*"] }` options:

```json
{ "no-warning-comments": ["error", { "decoration": ["/", "*"] }] }
```

```javascript
////// TODO decorative slashes and whitespace are ignored //////
//***** todo decorative asterisks are also ignored *****//
/**
 * TODO new lines are also ignored in block comments.
 */
```

Examples of **correct** code for the `{ "decoration": ["/", "*"] }` options:

```javascript
//!TODO preceded by non-decoration character
/**
 *!TODO preceded by non-decoration character in a block comment
 */
```

## Original Documentation

- [ESLint: no-warning-comments](https://eslint.org/docs/latest/rules/no-warning-comments)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-warning-comments.js)
