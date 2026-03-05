# no-unnecessary-template-expression

## Rule Details

Disallow unnecessary template expressions.

Template literals that contain only a single string variable or simple literal values can be simplified. If a template literal contains a single interpolation with a string-typed expression and nothing else, the template wrapper is unnecessary.

Examples of **incorrect** code for this rule:

```typescript
const ab = `${'a'}`;
const greeting = `${name}`; // when name is typed as string
const value = `${true}`;
const num = `${100}`;
```

Examples of **correct** code for this rule:

```typescript
const ab = 'a';
const greeting = name;
const combined = `Hello, ${name}!`;
const tagged = tag`${value}`;
```

## Original Documentation

- [typescript-eslint no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)
