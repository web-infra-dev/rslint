# prefer-while-loop-condition

## Rule Details

Prefer putting a loop's continuation condition in its `while` statement. This
rule reports infinite loops whose first statement is an `if` without an `else`
that only performs an unlabeled `break`.

Examples of **incorrect** code:

```javascript
while (true) {
  if (!hasMore()) {
    break;
  }
  processNext();
}

for (;;) {
  if (done) break;
  processNext();
}
```

Examples of **correct** code:

```javascript
while (hasMore()) {
  processNext();
}

while (!done) {
  processNext();
}
```

The rule checks `while (true)`, `for (;;)`, `for (; true;)`, and
`do ... while (true)`. It ignores labeled loops, additional breaks targeting
the same loop, and guards that reference lexical bindings declared inside the
loop body. Breaks inside a nested loop, switch, or function do not prevent a
report.

## Automatic Fixes

The fix moves the negated guard condition into the loop header and removes the
guard. Infinite `for` and `do ... while` loops become `while` loops.

Comments around the guard or loop syntax prevent automatic fixes. Comments
inside the retained body are preserved and do not prevent fixing.

Conditions containing direct `eval` are reported without an automatic fix.

## Options

This rule has no options.

## Differences from upstream

Unlike eslint-plugin-unicorn v79.0.0, rslint withholds automatic fixes when the
guard condition contains direct `eval`. For example, this code is reported but
left unchanged by automatic fixing:

```javascript
'use strict';

while (true) {
  if (eval('done()')) break;
  function done() {
    return true;
  }
}
```

Moving the `eval` call into the loop header would make `done` inaccessible and
cause a `ReferenceError`.

## Original Documentation

- [eslint-plugin-unicorn: prefer-while-loop-condition](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-while-loop-condition.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-while-loop-condition.js)
