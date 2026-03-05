# no-implied-eval

## Rule Details

Disallows the use of `eval()`-like methods. Functions such as `setTimeout`, `setInterval`, `setImmediate`, and `execScript` can accept a string argument that is evaluated as code, similar to `eval()`. This is dangerous because it can execute arbitrary code and makes the application vulnerable to injection attacks. This rule also disallows using the `Function` constructor to dynamically create functions from strings.

Examples of **incorrect** code for this rule:

```typescript
setTimeout("alert('hello')", 100);

setInterval('doWork()', 1000);

const fn = new Function('a', 'b', 'return a + b');

window.setTimeout('doSomething()', 100);
```

Examples of **correct** code for this rule:

```typescript
setTimeout(() => alert('hello'), 100);

setInterval(doWork, 1000);

const fn = (a: number, b: number) => a + b;

window.setTimeout(() => doSomething(), 100);
```

## Original Documentation

- [typescript-eslint no-implied-eval](https://typescript-eslint.io/rules/no-implied-eval)
