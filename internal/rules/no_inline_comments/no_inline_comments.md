# no-inline-comments

## Rule Details

Some style guides disallow a comment on the same line as code. Comments alone on a line, or on a line consisting entirely of whitespace, are allowed.

This rule disallows comments on the same line as code.

Examples of **incorrect** code for this rule:

```javascript
var a = 1; // declaring a to 1

function getRandomNumber() {
  return 4; // chosen by fair dice roll.
  // guaranteed to be random.
}

/* A block comment before code */ var b = 2;

var c = 3; /* A block comment after code */
```

Examples of **correct** code for this rule:

```javascript
// This is a comment above a line of code
var foo = 5;

var bar = 5;
//This is a comment below a line of code
```

### JSX exception

Comments that are the only content of a JSX expression container are not considered inline, since there is no code on either side of them.

```jsx
var a = (
  <div>
    {/* comment */}
    <h1>Some heading</h1>
  </div>
);
```

## Options

This rule accepts a single options object with the following property:

- `ignorePattern` (string) - a pattern of comments to ignore

### ignorePattern

Examples of **correct** code for this rule with the `{ "ignorePattern": "webpackChunkName:\\s.+" }` option:

```json
{ "no-inline-comments": ["error", { "ignorePattern": "webpackChunkName:\\s.+" }] }
```

```javascript
import(/* webpackChunkName: "my-chunk-name" */ './locale/en');
```

## Original Documentation

- [ESLint: no-inline-comments](https://eslint.org/docs/latest/rules/no-inline-comments)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-inline-comments.js)
