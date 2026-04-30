# no-loop-func

Disallow function declarations that contain unsafe references inside loop statements.

## Rule Details

This is the TypeScript-aware version of the ESLint [`no-loop-func`](https://eslint.org/docs/latest/rules/no-loop-func) rule. It reports any function created inside a loop body that closes over a variable that may be modified across iterations, which typically indicates a mistake — every closure ends up reading the final value of the variable instead of the value at the time the closure was created.

Type-only references (used as TypeScript type annotations) are not flagged, because they have no runtime impact.

Examples of **incorrect** code for this rule:

```javascript
for (var i = 0; i < 10; i++) {
  function foo() {
    console.log(i);
  }
}
```

```typescript
for (var i = 0; i < 10; i++) {
  const handler = (event: Event) => {
    console.log(i);
  };
}
```

Examples of **correct** code for this rule:

```javascript
for (let i = 0; i < 10; i++) {
  function foo() {
    console.log(i);
  }
}
```

```typescript
let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}
```

```typescript
type MyType = 1;
let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}
```

## When Not To Use It

If you do not want to be notified about functions defined inside loops, you can safely disable this rule.

## Original Documentation

- [typescript-eslint/no-loop-func](https://typescript-eslint.io/rules/no-loop-func/)
- [ESLint no-loop-func](https://eslint.org/docs/latest/rules/no-loop-func)
