# no-unnecessary-template-expression

## Rule Details

Disallow unnecessary expressions inside template literals and template literal
types.

Literal interpolations can be written directly into the surrounding template.
When a template contains only one string-typed expression, the template wrapper
can be removed entirely. The rule provides an autofix for both forms while
preserving template escapes.

Examples of **incorrect** code for this rule:

```typescript
const ab = `${'a'}`;
const greeting = `${name}`; // when name is typed as string
const value = `${true}`;
const num = `${100}`;
type EventName = `on${'Click'}`;
```

Examples of **correct** code for this rule:

```typescript
const ab = 'a';
const greeting = name;
const combined = `Hello, ${name}!`;
const tagged = tag`${value}`;
type EventName = 'onClick';
```

## Original Documentation

- [typescript-eslint: no-unnecessary-template-expression](https://typescript-eslint.io/rules/no-unnecessary-template-expression)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.29.1/packages/eslint-plugin/src/rules/no-unnecessary-template-expression.ts)
