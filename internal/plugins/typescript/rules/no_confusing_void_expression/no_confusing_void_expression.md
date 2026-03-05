# no-confusing-void-expression

## Rule Details

Disallows void type expressions from being used in misleading locations such as being assigned to a variable, returned from a function, or used inside other expressions. The `void` type in TypeScript indicates a function returns nothing, and using void-returning expressions in value positions can lead to confusing code.

Examples of **incorrect** code for this rule:

```typescript
// Assigning a void expression to a variable
const result = console.log('hello');

// Returning a void expression from a function
function foo() {
  return console.log('hello');
}

// Using a void expression in an arrow function shorthand
const fn = () => console.log('hello');
```

Examples of **correct** code for this rule:

```typescript
// Void expression as a standalone statement
console.log('hello');

// Arrow function with braces
const fn = () => {
  console.log('hello');
};

// Explicit void operator (with ignoreVoidOperator option)
const fn = () => void console.log('hello');
```

## Original Documentation

- [typescript-eslint no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)
