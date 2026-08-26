# no-cond-assign

## Rule Details

Disallows assignment operators in test expressions (`if`, `while`, `do-while`, `for`, and ternary). Assignments in tests are frequently a typo where the developer meant to use a comparison operator (`===`) instead of an assignment operator (`=`).

In the default `"except-parens"` mode, a top-level assignment is allowed when extra parentheses signal that it is intentional. A ternary test needs two explicit pairs because, unlike a statement test, it has no grammar parentheses. In `"always"` mode, assignments descended from a test expression are flagged, stopping at nested function boundaries. An assignment in a ternary branch is not part of that ternary's test, so it is allowed when the ternary is outside another conditional test. If the whole ternary is itself an outer test, `"always"` mode reports the assignment against that outer conditional. Assignments in statement bodies are allowed in both modes.

Examples of **incorrect** code for this rule:

```javascript
if (x = 0) {
}

while (x = next()) {}

var result = (x = next()) ? y : z;
```

Examples of **correct** code for this rule:

```javascript
if (x === 0) {
}

while ((x = next())) {} // extra parens signal intent

if (x === 0 || (y = getValue())) {
}

for (; (a = b); ) {}

var result = ((x = next())) ? y : z;

var branchResult = x ? (y = 1) : z; // assignment is in a branch, not the test
```

## Original Documentation

- [ESLint: no-cond-assign](https://eslint.org/docs/latest/rules/no-cond-assign)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-cond-assign.js)
