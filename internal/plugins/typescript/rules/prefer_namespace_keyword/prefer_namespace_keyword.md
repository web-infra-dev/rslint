# prefer-namespace-keyword

## Rule Details

Require using `namespace` keyword over `module` keyword to declare custom TypeScript modules.

TypeScript historically allowed `module` keyword as a way to group related code. The newer `namespace` keyword was added in TypeScript 1.5 to distinguish between built-in modules and user-defined modules. While the two keywords are functionally identical, using `namespace` is recommended as `module` may cause confusion with ECMAScript modules.

Examples of **incorrect** code for this rule:

```typescript
module Foo {}

declare module Foo {}
```

Examples of **correct** code for this rule:

```typescript
namespace Foo {}

declare namespace Foo {}

declare module 'foo' {}
```

## Original Documentation

- [typescript-eslint prefer-namespace-keyword](https://typescript-eslint.io/rules/prefer-namespace-keyword)
