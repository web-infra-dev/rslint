# process-exit-as-throw

Lets other rules treat `process.exit()` as ending the current execution path, like `throw`.
This rule does not report diagnostics itself and has no options, fixes, or suggestions.

For example, enabling this rule together with `consistent-return` accepts:

```javascript
function foo(a) {
  if (a) {
    return new Bar();
  } else {
    process.exit(1);
  }
}
```

With `no-unreachable`, the statement after the exit is reported:

```javascript
process.exit(1);
doMoreWork(); // Unreachable code.
```

This also lets a `process.exit()` call satisfy `no-fallthrough` and `getter-return`.
Only direct, non-computed
`process.exit()` calls match; `process["exit"]()` and aliases do not. Like upstream,
the match is syntactic, so a locally declared variable named `process` also matches.

## Differences from upstream

Use `node/process-exit-as-throw` in your rslint configuration instead of
`n/process-exit-as-throw`. Enabling it affects supported built-in rules;
it does not change TypeScript errors or diagnostics from third-party ESLint plugins.

`array-callback-return`, `constructor-super`, and `no-this-before-super` do not
yet recognize this rule's exits and can still report errors. For example,
`class C extends Base { constructor() { process.exit(1); } }` can still trigger
`constructor-super`. In these function bodies, `return process.exit(1);` makes
the exit explicit to those rules as well.

There are also differences with `no-unreachable` when comparing against the
upstream rule with ESLint 7:

- rslint reports unreachable bodies and following statements when the source of a
  `for-in` or `for-of` loop exits, as in
  `for (const x of process.exit()) { work(); } after();`.
  It also reports unreachable cases in
  `switch (x) { default: work(); case process.exit(): more(); }`.
  ESLint 7 misses these unreachable statements and can still report fallthrough
  in that switch when `no-fallthrough` is enabled.
- An exit inside a class field initializer or static block only affects
  diagnostics within that member. For example, rslint accepts
  `class C { value = process.exit(); } after();` with `no-unreachable`, while
  the upstream rule with ESLint 7 reports `after()` as unreachable.

## Original Documentation

- [eslint-plugin-n documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/process-exit-as-throw.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/process-exit-as-throw.js)
