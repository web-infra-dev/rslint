# no-empty

## Rule Details

Disallow empty block statements. Empty block statements, while not technically errors, usually occur due to refactoring that wasn't completed. They can cause confusion when reading code.

Examples of **incorrect** code for this rule:

```javascript
if (foo) {
}

while (foo) {}

switch (foo) {
}

try {
  doSomething();
} catch (e) {}
```

Examples of **correct** code for this rule:

```javascript
if (foo) {
  // empty
}

while (foo) {
  /* todo */
}

try {
  doSomething();
} catch (e) {
  // expected
}

function foo() {}
```

## Options

- `allowEmptyCatch`: If `true`, allows empty `catch` clauses (i.e., which do not contain a comment). Default: `false`.

## Original Documentation

https://eslint.org/docs/latest/rules/no-empty
