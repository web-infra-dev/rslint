# no-implied-eval

## Rule Details

Disallow the use of `eval()`-like methods.

Passing a string as the first argument to `setTimeout`, `setInterval`, or `execScript` causes the string to be evaluated as JavaScript — a form of implied `eval()`. This rule flags such calls, whether made directly or through a global object reference (`window`, `global`, `globalThis`, or `self`).

Examples of **incorrect** code for this rule:

```javascript
setTimeout('alert(\'Hi!\');', 100);
setInterval('alert(\'Hi!\');', 100);
execScript('alert(\'Hi!\')');

window.setTimeout('count = 5', 10);
window.setInterval('foo = bar', 10);

globalThis.setTimeout(`code ${foo}`);
self.setInterval('foo' + bar);
```

Examples of **correct** code for this rule:

```javascript
setTimeout(function () {
  alert('Hi!');
}, 100);

setInterval(function () {
  alert('Hi!');
}, 100);

execScript(function () {
  alert('Hi!');
});

const handler = () => alert('Hi!');
setTimeout(handler, 100);
window.setTimeout(handler, 100);
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-implied-eval
