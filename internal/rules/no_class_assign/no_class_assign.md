# no-class-assign

## Rule Details

Disallows reassigning variables that were declared as class declarations. Reassigning a class declaration is almost always a mistake, as it overwrites the class with a different value. This rule checks for assignments, increment/decrement operations, and destructuring assignments that target a class name.

Examples of **incorrect** code for this rule:

```javascript
class A {}
A = 0;

class B {}
B += 1;

class C {}
({ C } = obj);
```

Examples of **correct** code for this rule:

```javascript
class A {}
var a = new A();

let B = class {};
B = 0; // B is a variable, not a class declaration

class C {}
function fn(C) {
  C = 0; // C is a parameter, not the class
}
```

## Original Documentation

- [ESLint no-class-assign](https://eslint.org/docs/latest/rules/no-class-assign)
