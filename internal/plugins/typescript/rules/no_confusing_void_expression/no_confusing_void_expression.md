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

- [typescript-eslint: no-confusing-void-expression](https://typescript-eslint.io/rules/no-confusing-void-expression)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-confusing-void-expression.ts)

## Differences from upstream

Rslint only removes grouping parentheses when they are unnecessary in statement
position and contain no comments. For example,
`const f = () => ((console.log()));` is fixed to
`const f = () => { console.log(); };`. Upstream v8.67.0 removes only the innermost
pair and produces `const f = () => ({ console.log(); });`, which is invalid syntax.

Autofixes preserve necessary parentheses around function, class and object
expressions, preserve comments, and protect statement boundaries when moving a
return argument. This also covers expressions starting with `+`, `-`, a regular
expression, a template literal or a TypeScript type assertion. Upstream v8.67.0
can produce invalid syntax or join the expression to the preceding statement.

Rslint also withholds autofixes for non-final returns whose complete argument
is not void-like. For example, `return flag ? console.log() : 42;` must retain
the value `42` when `flag` is false. Upstream v8.67.0 can replace this with an
expression followed by a bare `return`, discarding that value.
