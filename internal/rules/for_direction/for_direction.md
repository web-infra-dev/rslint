# for-direction

## Rule Details

Enforces that the update clause in a `for` loop moves the counter variable in the correct direction relative to the loop's stop condition. A `for` loop with a counter that moves in the wrong direction will run infinitely.

Examples of **incorrect** code for this rule:

```javascript
for (var i = 0; i < 10; i--) {}

for (var i = 10; i >= 0; i++) {}

for (var i = 0; i < 10; i -= 1) {}
```

Examples of **correct** code for this rule:

```javascript
for (var i = 0; i < 10; i++) {}

for (var i = 10; i >= 0; i--) {}

for (var i = 0; i < 10; i += 1) {}
```

## Original Documentation

- [ESLint for-direction](https://eslint.org/docs/latest/rules/for-direction)
