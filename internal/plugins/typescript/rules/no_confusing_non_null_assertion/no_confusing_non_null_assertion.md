# no-confusing-non-null-assertion

## Rule Details

Disallow non-null assertion in locations that may be confusing.

A non-null assertion (`!`) placed immediately before `=`, `==`, `===`, `in`, or `instanceof` is visually almost indistinguishable from the operators `!=`, `!==`, `!(... in ...)`, or `!(... instanceof ...)`. This rule flags those combinations and offers suggestions to either remove the assertion or wrap the left-hand side in parentheses to disambiguate.

Examples of **incorrect** code for this rule:

```typescript
a! == b;
a! === b;
a! in b;
a! instanceof b;
```

Examples of **correct** code for this rule:

```typescript
a == b;
(1 + foo.num!) == 2;
foo.bar == 'hello';
!(a in b);
```

## Original Documentation

- [typescript-eslint no-confusing-non-null-assertion](https://typescript-eslint.io/rules/no-confusing-non-null-assertion)
