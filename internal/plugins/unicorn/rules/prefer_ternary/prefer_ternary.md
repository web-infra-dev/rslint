# prefer-ternary

## Rule Details

This rule flags simple `if` statements that return or assign a value and
suggests replacing them with a ternary expression. A separate branch
collapses `let x = a; if (test) { x = b; }` into a single
`const x = test ? b : a;` declaration (or `let x = …` when the variable
gets a later write).

The motivation is that `if-else` statements typically produce more lines
than a ternary, leading to a larger, harder-to-maintain codebase. They can
also force developers to use `let` or `var` for variables that only get
reassigned, which unnecessarily introduces mutability and blocks
`prefer-const` from flagging the variable.

Examples of **incorrect** code for this rule:

```javascript
function unicorn() {
    if (test) {
        return a;
    } else {
        return b;
    }
}

let foo;
if (test) {
    foo = 1;
} else {
    foo = 2;
}
```

Examples of **correct** code for this rule:

```javascript
function unicorn() {
    return test ? a : b;
}

const foo = test ? 1 : 2;
```

## Options

Type: `string`

Default: `'always'`

- `'always'` (default) — Always report supported `IfStatement` returns and
  assignments where a ternary can be used.
- `'only-single-line'` — Only report when the condition and the merged
  expressions fit on a single line.

Example with `'only-single-line'`:

```javascript
// Not flagged — array literal spans multiple lines.
if (test) {
    foo = [
        'multiple line array'
    ];
} else {
    foo = bar;
}
```

## Differences from ESLint

The upstream `unicorn/prefer-ternary` rule is reported under
`eslint-plugin-unicorn`'s `recommended` and `unopinionated` configs and
auto-fixes via `--fix`. The current port reproduces that contract:

- The simple `if/else` form is rewritten to a ternary and the result is
  reported with a single diagnostic whose `fix` rewrites the source.
- The `let x = a; if (test) x = b;` form is reported with a diagnostic
  whose `suggest` rewrites the source to `const x = test ? b : a;`
  (or `let` when the variable receives a later write).
- Two upstream test cases are not ported: the
  `(foo)['b' + 'ar'] = a` / `foo.bar = b` shape (the LHS is not the same
  reference after `isSameReference` walks through the parens and
  computed-key access) and the nested `if` outer branches whose
  alternate is another `if` (upstream reports both, but the rslint port
  only reports the inner one whose consequent and alternate are
  mergeable). These are surfaced as no-report in the upstream suite
  here; they remain as `// Reported upstream, not flagged here` notes in
  the test file for traceability.

## Original Documentation

- [eslint-plugin-unicorn: prefer-ternary](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/prefer-ternary.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/prefer-ternary.js)
