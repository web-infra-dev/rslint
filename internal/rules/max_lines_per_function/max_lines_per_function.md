# max-lines-per-function

Enforce a maximum number of lines of code in a function.

## Rule Details

Long functions are harder to follow than short ones. This rule reports any
function whose body — including its declaration line and the closing brace —
exceeds the configured line limit.

By default, the rule checks for a maximum of 50 lines per function. Comments
and blank lines are counted, and IIFEs are skipped.

Examples of **incorrect** code for this rule with the default `{ "max": 50 }`
option (function body shortened for brevity):

```javascript
function longFunction() {
  // imagine 50 more lines here
}
```

Examples of **correct** code for this rule with the default option:

```javascript
function shortFunction() {
  doThing();
  return result;
}
```

## Options

This rule accepts a number (the maximum allowed) or an object with the
following properties:

- `max` (default `50`): maximum number of lines a function may contain.
- `skipBlankLines` (default `false`): ignore lines made up purely of
  whitespace.
- `skipComments` (default `false`): ignore lines that contain only comments
  (a line with both code and a comment still counts).
- `IIFEs` (default `false`): when `true`, IIFEs are checked like other
  functions; when `false`, they are skipped.

### `max`

```json
{ "max-lines-per-function": ["error", { "max": 2 }] }
```

Examples of **incorrect** code for this rule with `{ "max": 2 }`:

```javascript
function name() {
  var x = 5;
  var y = 2;
}
```

Examples of **correct** code for this rule with `{ "max": 3 }`:

```javascript
function name() {
  var x = 5;
}
```

### `skipBlankLines`

```json
{ "max-lines-per-function": ["error", { "max": 3, "skipBlankLines": true }] }
```

Examples of **correct** code for this rule with the above configuration:

```javascript
function name() {
  var x = 5;

  var y = 2;
}
```

### `skipComments`

```json
{ "max-lines-per-function": ["error", { "max": 3, "skipComments": true }] }
```

Examples of **correct** code for this rule with the above configuration:

```javascript
function name() {
  // a comment
  var x = 5;
  var y = 2;
}
```

### `IIFEs`

```json
{ "max-lines-per-function": ["error", { "max": 2, "IIFEs": true }] }
```

Examples of **incorrect** code for this rule with the above configuration:

```javascript
(function () {
  var x = 0;
  var y = 0;
})();
```

## Original Documentation

- [https://eslint.org/docs/latest/rules/max-lines-per-function](https://eslint.org/docs/latest/rules/max-lines-per-function)
