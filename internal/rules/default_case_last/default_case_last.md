# default-case-last

## Rule Details

Enforce `default` clauses in `switch` statements to be last. A `switch` statement can optionally have a `default` clause. If present, it's usually the last clause, but it doesn't need to be. It is also allowed to put the `default` clause before all `case` clauses, or anywhere between. The behavior is mostly the same as if it was the last clause.

The `default` clause is still executed only if there is no match in the `case` clauses (including those defined after the `default`), but there is also the ability to "fall through" from the `default` clause to the following clause in the list. However, such flow is not common and can be confusing.

This rule enforces `default` clauses in `switch` statements to be last, after all `case` clauses.

Examples of **incorrect** code for this rule:

```javascript
switch (foo) {
  default:
    bar();
    break;
  case 1:
    baz();
    break;
}

switch (foo) {
  case 1:
    break;
  default:
    break;
  case 2:
    break;
}
```

Examples of **correct** code for this rule:

```javascript
switch (foo) {
  case 1:
    bar();
    break;
  default:
    baz();
    break;
}

switch (foo) {
  case 1:
    break;
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/default-case-last
