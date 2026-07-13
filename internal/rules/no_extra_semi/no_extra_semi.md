# no-extra-semi

Disallow unnecessary semicolons.

## Rule Details

This rule flags semicolons that can be safely removed without changing the semantics of the JavaScript code.

Examples of **incorrect** code for this rule:

```javascript
var x = 5;;

function foo(){};

for(;;);;

while(0);;

do;while(0);;

class A {
  ;
  field;;
  a() {};
}
```

Examples of **correct** code for this rule:

```javascript
var x = 5;

function foo(){}

for(;;);

while(0);

do;while(0);

class A {
  field;
  a() {}
}
```

## Original Documentation

[ESLint no-extra-semi documentation](https://eslint.org/docs/latest/rules/no-extra-semi)
