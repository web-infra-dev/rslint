# no-loop-func

Disallow function declarations that contain unsafe references inside loop statements.

## Rule Details

Writing functions within loops tends to result in errors due to the way the function creates a closure around the loop. For example:

```javascript
for (var i = 10; i; i--) {
    (function() { return i; })();
}
```

Generally speaking, it is safer to keep the closure code outside of the loop, or to use `let` / `const` for loop variables so each iteration produces a fresh binding.

This rule disallows any function within a loop that contains unsafe references (e.g. to modified variables).

Examples of **incorrect** code for this rule:

```javascript
for (var i = 0; i < 10; i++) {
    funcs[i] = function() { return i; };
}

for (var i = 0; i < 10; i++) {
    funcs[i] = () => i;
}

for (var i = 0; i < 10; i++) {
    funcs[i] = function() { return i; };
    funcs[i]();
}

var foo = 100;
for (var i = 0; i < 10; i++) {
    funcs[i] = function() { return foo; };
    foo += 1;
}

var foo = 100;
check(function() { return foo; });
foo = 200;
```

Examples of **correct** code for this rule:

```javascript
var a = function() {};

for (var i = 0; i < 10; i++) {
    funcs[i] = a;
}

for (let i = 0; i < 10; i++) {
    funcs[i] = function() { return i; };
}

const foo = 100;
for (var i = 0; i < 10; i++) {
    funcs[i] = function() { return foo; };
}

// IIFEs are fine because they execute immediately.
for (var i = 0; i < 10; i++) {
    funcs[i] = (function() { return i; })();
}
```

## Original Documentation

- [ESLint rule](https://eslint.org/docs/latest/rules/no-loop-func)
- [Source code](https://github.com/eslint/eslint/blob/main/lib/rules/no-loop-func.js)
