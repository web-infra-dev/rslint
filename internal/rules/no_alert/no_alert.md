# no-alert

## Rule Details

Disallow the use of `alert`, `confirm`, and `prompt`.

JavaScript's `alert`, `confirm`, and `prompt` functions are widely considered to be obtrusive as UI elements and should be replaced by a more appropriate custom UI implementation. Furthermore, `alert` is often used while debugging code, which should be removed before deployment to production.

Examples of **incorrect** code for this rule:

```javascript
alert('here!');
confirm('Are you sure?');
prompt("What's your name?", 'John Doe');

window.alert('here!');
globalThis.confirm('Are you sure?');
this.prompt("What's your name?");
```

Examples of **correct** code for this rule:

```javascript
customAlert('Something happened!');
customConfirm('Are you sure?');
customPrompt('Who are you?');

function foo() {
  var alert = myCustomLib.customAlert;
  alert();
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-alert
