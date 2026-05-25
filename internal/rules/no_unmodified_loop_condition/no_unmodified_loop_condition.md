# no-unmodified-loop-condition

## Rule Details

Disallows variables in loop conditions that are not modified inside the loop body. If a variable used in a loop's test condition is never assigned to, incremented, or decremented within the loop, it is likely a bug that leads to an infinite loop or incorrect termination.

Conditions that contain function calls, member access expressions, `new` expressions, or tagged template expressions are skipped, since those may have side effects that modify the condition indirectly.

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

## Original Documentation

- [ESLint no-unmodified-loop-condition](https://eslint.org/docs/latest/rules/no-unmodified-loop-condition)
