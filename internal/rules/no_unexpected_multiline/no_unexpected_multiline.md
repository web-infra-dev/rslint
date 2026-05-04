# no-unexpected-multiline

## Rule Details

JavaScript inserts semicolons automatically (ASI), but only at certain positions. When a line break appears between an expression and tokens like `(`, `[`, `` ` ``, or `/`, the parser treats the next line as a continuation rather than a new statement, which often surprises the author.

This rule reports four cases where a newline produces an unintended continuation:

- A function call where `(` opens on the next line.
- A computed property access where `[` opens on the next line.
- A tagged template where the `` ` `` opens on the next line.
- A division by a value that visually resembles a regular-expression literal (e.g. `foo / bar /gym`), where what looks like a regex is actually parsed as two divisions.

Examples of **incorrect** code for this rule:

```javascript
var a = b
(x || y).doSomething()

var a = b
[a, b, c].forEach(doSomething)

let x = function() {}
`hello`

foo
/ bar /gym
```

Examples of **correct** code for this rule:

```javascript
var a = b;
(x || y).doSomething()

var a = b;
[a, b, c].forEach(doSomething)

let x = function() {};
`hello`

foo / bar / gym
```

## Original Documentation

- [ESLint no-unexpected-multiline](https://eslint.org/docs/latest/rules/no-unexpected-multiline)
