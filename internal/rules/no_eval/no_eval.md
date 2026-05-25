# no-eval

## Rule Details

Disallow the use of `eval()`. JavaScript's `eval()` function is potentially dangerous and is often misused. Using `eval()` on untrusted code can open a program up to several different injection attacks. The use of `eval()` in most contexts can be substituted for a better, alternative approach to a problem.

Examples of **incorrect** code for this rule:

```javascript
eval('var a = 0');

var foo = eval;

this.eval('var a = 0');

window.eval('var a = 0');

global.eval('var a = 0');

globalThis.eval('var a = 0');
```

Examples of **correct** code for this rule:

```javascript
var obj = { eval: function () {} };
obj.eval('var a = 0');

class A {
  eval() {}
}
new A().eval('var a = 0');
```

### Options

This rule has an option to allow indirect calls to `eval`. Indirect calls to `eval` are less dangerous than direct calls because they cannot dynamically change the scope.

```json
{
  "no-eval": ["error", { "allowIndirect": true }]
}
```

With `{ "allowIndirect": true }`, the following patterns are **correct**:

```javascript
(0, eval)('var a = 0');

var EVAL = eval;
EVAL('var a = 0');

window.eval('var a = 0');
```

## Original Documentation

- [ESLint no-eval](https://eslint.org/docs/latest/rules/no-eval)
