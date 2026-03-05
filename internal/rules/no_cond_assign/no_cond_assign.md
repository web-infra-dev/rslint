# no-cond-assign

## Rule Details

Disallows assignment operators in conditional expressions (`if`, `while`, `do-while`, `for`, and ternary). Assignments in conditional statements are frequently a typo where the developer meant to use a comparison operator (`===`) instead of an assignment operator (`=`).

In the default `"except-parens"` mode, assignments are allowed if they are wrapped in extra parentheses, which signals the assignment is intentional. In `"always"` mode, all assignments in conditionals are flagged.

Examples of **incorrect** code for this rule:

```javascript
if ((x = 0)) {
}

while ((x = next())) {}

var result = x ? (y = 1) : z;
```

Examples of **correct** code for this rule:

```javascript
if (x === 0) {
}

while ((x = next())) {} // extra parens signal intent

if (x === 0 || (y = getValue())) {
}

for (; (a = b); ) {}
```

## Original Documentation

- [ESLint no-cond-assign](https://eslint.org/docs/latest/rules/no-cond-assign)
