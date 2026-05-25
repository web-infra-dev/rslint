# complexity

Enforce a maximum cyclomatic complexity allowed in a program.

## Rule Details

Cyclomatic complexity measures the number of linearly independent paths through a program's source code. This rule allows setting a cyclomatic complexity threshold per function. Functions whose complexity exceeds the threshold are reported.

The complexity counter is seeded at 1 (a single execution path) and increased by 1 for every branching construct: `if`, `else if`, `for`, `for…in`, `for…of`, `while`, `do…while`, `switch` cases, conditional expressions (`?:`), logical operators (`&&`, `||`, `??`), logical assignment operators (`&&=`, `||=`, `??=`), `catch` clauses, optional chaining links (`?.`), default parameter values (`function f(x = 1)`), and destructuring defaults (`const { x = 1 } = obj`).

Class field initializers and class static blocks each form their own complexity scope (separate from the enclosing function), matching ESLint's `class-field-initializer` and `class-static-block` code-path origins. A class field whose initializer is itself a function or arrow does NOT create a separate field-initializer scope — the function takes over.

Examples of **incorrect** code for this rule with the default `{ "max": 20 }`:

```javascript
function a(x) {
    if (true) {
        return x;
    } else if (false) {
        return x + 1;
    } else {
        return 4; // 3rd path exceeds limit if max is 2
    }
}
```

Examples of **correct** code for this rule with the default `{ "max": 20 }`:

```javascript
function a(x) {
    if (true) {
        return x;
    } else {
        return 4;
    }
}
```

## Options

This rule accepts either a number (the threshold) or an object.

### `max` (default: `20`)

Sets the maximum cyclomatic complexity allowed.

```json
{ "complexity": ["error", { "max": 2 }] }
```

```javascript
function a(x) {
    if (true) {
        return x;
    } else if (false) {
        return x + 1;
    } else {
        return 4;
    }
}
```

The deprecated property `maximum` is also recognized; when both `maximum` and `max` are present, `maximum` wins if its value is truthy (mirroring ESLint's `option.maximum || option.max` coercion).

### `variant` (default: `"classic"`)

Selects the complexity calculation method.

- `"classic"` — Standard McCabe cyclomatic complexity. Each `case` clause adds 1.
- `"modified"` — Each `switch` statement adds 1 regardless of how many `case` clauses it contains, and individual `case` clauses do not add to the complexity.

```json
{ "complexity": ["error", { "max": 3, "variant": "modified" }] }
```

```javascript
function a(x) {
    switch (x) {
        case 1: return 1;
        case 2: return 2;
        case 3: return 3;
        default: return 0;
    }
}
```

## Original Documentation

- [ESLint `complexity` rule](https://eslint.org/docs/latest/rules/complexity)
