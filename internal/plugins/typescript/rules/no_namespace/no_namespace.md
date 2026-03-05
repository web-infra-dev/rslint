# no-namespace

## Rule Details

Disallows the use of TypeScript `namespace` declarations. TypeScript historically allowed organizing code with custom `namespace` blocks, but ES2015 modules (using `import`/`export`) are the modern standard for code organization. Namespaces are generally considered outdated and should be replaced with modules. By default, this rule allows namespaces in `.d.ts` definition files, since they are commonly used there.

Examples of **incorrect** code for this rule:

```typescript
namespace MyNamespace {
  export const value = 1;
}

namespace Nested {
  export namespace Inner {
    export type Foo = string;
  }
}
```

Examples of **correct** code for this rule:

```typescript
// Use ES modules instead
export const value = 1;

// Declare namespaces are allowed with allowDeclarations option
declare namespace ExternalLib {
  function doWork(): void;
}

// Namespaces in .d.ts files are allowed by default
// (in file.d.ts)
declare namespace MyLib {
  interface Options {}
}
```

## Original Documentation

- [typescript-eslint no-namespace](https://typescript-eslint.io/rules/no-namespace)
