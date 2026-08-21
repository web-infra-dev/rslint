# func-style

## Rule Details

There are two ways of defining functions in JavaScript: `function` declarations and function expressions assigned to variables. Function expressions can either be arrow functions or use the `function` keyword with an optional name.

```javascript
// function declaration
function doSomething() {
  // ...
}

// arrow function expression assigned to a variable
const doSomethingElse = () => {
  // ...
};

// function expression assigned to a variable
const doSomethingAgain = function () {
  // ...
};
```

The primary difference between `function` declarations and function expressions is that declarations are _hoisted_ to the top of the scope in which they are defined, which allows using the function before its declaration; a function expression must be defined before it is used.

This rule enforces a particular type of function style, either `function` declarations or expressions assigned to variables.

This rule does not apply to all functions. A callback function passed as an argument to another function, or a method assigned to an object, is not checked by this rule.

Examples of **incorrect** code for this rule with the default `"expression"` option:

```javascript
function foo() {
  // ...
}
```

Examples of **correct** code for this rule with the default `"expression"` option:

```javascript
const foo = function () {
  // ...
};

const foo1 = () => {};
```

Overloaded function declarations (multiple declarations of the same name with different parameter or return types) are never reported by this rule, regardless of the configured style:

```typescript
function process(value: string): string;
function process(value: number): number;
function process(value: unknown) {
  return value;
}
```

Examples of **incorrect** code for this rule with the `"declaration"` option:

```json
{ "func-style": ["error", "declaration"] }
```

```javascript
const foo = function () {
  // ...
};

const foo1 = () => {};
```

Examples of **correct** code for this rule with the `"declaration"` option:

```json
{ "func-style": ["error", "declaration"] }
```

```javascript
function foo() {
  // ...
}

// Methods (functions assigned to objects) are not checked by this rule
SomeObject.foo = function () {
  // ...
};
```

## Options

This rule has a string option:

- `"expression"` (default) requires the use of function expressions instead of function declarations
- `"declaration"` requires the use of function declarations instead of function expressions

This rule has an object option:

- `"allowArrowFunctions"`: `true` (default `false`) allows the use of arrow functions when the string option is `"declaration"`. Arrow functions are always allowed when the string option is `"expression"`, regardless of this option.
- `"allowTypeAnnotation"`: `true` (default `false`) allows a function expression or arrow function whose variable declaration has a type annotation, regardless of `allowArrowFunctions`. This option applies only when the string option is `"declaration"`.
- `"overrides"`:
  - `"namedExports"`: `"expression" | "declaration" | "ignore"` overrides the function style required for named exports. `"ignore"` accepts either style.

### allowArrowFunctions

Examples of additional **correct** code for this rule with `{ "allowArrowFunctions": true }`:

```json
{ "func-style": ["error", "declaration", { "allowArrowFunctions": true }] }
```

```javascript
const foo = () => {};
```

### allowTypeAnnotation

Examples of **incorrect** code for this rule with `{ "allowTypeAnnotation": true }`:

```json
{ "func-style": ["error", "declaration", { "allowTypeAnnotation": true }] }
```

```typescript
const foo = function (): void {};
```

Examples of **correct** code for this rule with `{ "allowTypeAnnotation": true }`:

```json
{ "func-style": ["error", "declaration", { "allowTypeAnnotation": true }] }
```

```typescript
type Fn = () => undefined;

const foo: Fn = function () {};

const bar: Fn = () => {};
```

### overrides.namedExports

Examples of **incorrect** code for this rule with `{ "overrides": { "namedExports": "expression" } }`:

```json
{
  "func-style": [
    "error",
    "declaration",
    { "overrides": { "namedExports": "expression" } }
  ]
}
```

```javascript
export function foo() {
  // ...
}
```

Examples of **correct** code for this rule with `{ "overrides": { "namedExports": "expression" } }`:

```json
{
  "func-style": [
    "error",
    "declaration",
    { "overrides": { "namedExports": "expression" } }
  ]
}
```

```javascript
export const foo = function () {
  // ...
};

export const bar = () => {};
```

Examples of **incorrect** code for this rule with `{ "overrides": { "namedExports": "declaration" } }`:

```json
{
  "func-style": [
    "error",
    "expression",
    { "overrides": { "namedExports": "declaration" } }
  ]
}
```

```javascript
export const foo = function () {
  // ...
};

export const bar = () => {};
```

Examples of **correct** code for this rule with `{ "overrides": { "namedExports": "declaration" } }`:

```json
{
  "func-style": [
    "error",
    "expression",
    { "overrides": { "namedExports": "declaration" } }
  ]
}
```

```javascript
export function foo() {
  // ...
}
```

Examples of **correct** code for this rule with `{ "overrides": { "namedExports": "ignore" } }`:

```json
{
  "func-style": [
    "error",
    "expression",
    { "overrides": { "namedExports": "ignore" } }
  ]
}
```

```javascript
export const foo = function () {
  // ...
};

export const bar = () => {};

export function baz() {
  // ...
}
```

## Original Documentation

- [ESLint: func-style](https://eslint.org/docs/latest/rules/func-style)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/func-style.js)
