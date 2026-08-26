# no-unmodified-loop-condition

## Rule Details

Disallows variables in loop conditions that are not modified by the loop. A modification may occur while evaluating the condition, in the loop body, or in a `for` loop's increment expression. If a variable used in a loop's test condition is never assigned to, incremented, or decremented in any of those places, it is likely a bug that leads to an infinite loop or incorrect termination.

References nested inside function calls, member access expressions, `new` expressions, or `yield` expressions are skipped, since those may have side effects that modify the condition indirectly. Binary expression groups, and by default ternary expression groups, are also skipped when they contain one of those dynamic expressions or a tagged template.

Examples of **incorrect** code for this rule:

```javascript
var foo = 0;
while (foo) {
  // foo is never modified
  doSomething();
}

var bar = 0;
do {
  doSomething();
} while (bar);

for (var i = 0; i < 10; ) {
  // i is never modified, no incrementor
  doSomething();
}
```

Examples of **correct** code for this rule:

```javascript
var foo = 0;
while (foo) {
  foo++;
}

var bar = 0;
do {
  bar = getNextValue();
} while (bar);

var remaining = 10;
while (remaining--) {
  processNext();
}

for (var i = 0; i < 10; i++) {
  doSomething();
}

// Function calls in condition are allowed (side effects possible)
while (hasNext()) {
  process();
}

// Member access in condition is allowed
while (obj.ready) {
  process();
}
```

## Options

This rule accepts an object option with `checkConditionalExpressions`, which defaults to `false`.

When `checkConditionalExpressions` is `true`, references in the test and branches of a ternary expression are checked independently instead of treating the whole ternary as one group. For example, this reports `done` because only `chunk` is modified:

```javascript
/* rslint no-unmodified-loop-condition: ["error", { "checkConditionalExpressions": true }] */

let chunk = getInitialChunk();
let done = false;

while (chunk ? !done : false) {
  chunk = nextOrNull();
}
```

## Original Documentation

- [ESLint: no-unmodified-loop-condition](https://eslint.org/docs/latest/rules/no-unmodified-loop-condition)
- [Source code](https://github.com/eslint/eslint/blob/9ef407a3b051e74f50dc7fb8914e2bd89b3e5e53/lib/rules/no-unmodified-loop-condition.js)
