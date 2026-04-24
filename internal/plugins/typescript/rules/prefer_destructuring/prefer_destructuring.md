# prefer-destructuring

## Rule Details

Require destructuring from arrays and/or objects.

This rule extends the base ESLint `prefer-destructuring` rule with TypeScript-specific type awareness.

Examples of **incorrect** code for this rule:

```javascript
var foo = object.foo;
var bar = array[0];
```

Examples of **correct** code for this rule:

```javascript
var { foo } = object;
var [bar] = array;
```

### Type Annotation Handling

By default, the rule does not report on variable declarations with type annotations, because the auto-fix would remove them:

```typescript
// This is correct by default (has type annotation)
const x: string = obj.x;
```

### Type-Aware Array Detection

The rule uses the TypeScript type checker to determine whether numeric index access (`x[0]`) should be treated as array destructuring or object destructuring:

- If the object type is iterable (has `[Symbol.iterator]`) or `any`, it is treated as array access
- If the object type is a plain object with numeric keys (e.g., `{ 0: unknown }`), it is treated as object access

```json
{ "@typescript-eslint/prefer-destructuring": ["error", { "object": true }, { "enforceForRenamedProperties": true }] }
```

```typescript
// Correct: x is not iterable, so numeric index is treated as object access
let x: { 0: unknown };
let y = x[0];
```

```typescript
// Incorrect: x is iterable (array), so numeric index triggers array destructuring
let x: number[];
let y = x[0]; // Use array destructuring
```

### `enforceForDeclarationWithTypeAnnotation`

```json
{ "@typescript-eslint/prefer-destructuring": ["error", { "object": true }, { "enforceForDeclarationWithTypeAnnotation": true }] }
```

Examples of **incorrect** code with `{ "enforceForDeclarationWithTypeAnnotation": true }`:

```typescript
const x: string = obj.x;
```

Examples of **correct** code with `{ "enforceForDeclarationWithTypeAnnotation": true }`:

```typescript
const { x }: { x: string } = obj;
```

## Original Documentation

- [ESLint core rule](https://eslint.org/docs/latest/rules/prefer-destructuring)
- [TypeScript-ESLint rule](https://typescript-eslint.io/rules/prefer-destructuring)
