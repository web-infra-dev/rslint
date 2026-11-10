// Every case from typescript-eslint v8.67.0, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/tests/rules/no-namespace.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
});

ruleTester.run('no-namespace', {
  valid: [
    {
      code: 'declare global {}',
      filename: 'file.ts',
    },
    {
      code: "declare module 'foo' {}",
      filename: 'file.ts',
    },
    {
      code: 'declare module foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: 'declare namespace foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: `
declare global {
  namespace foo {}
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: `
declare module foo {
  namespace bar {}
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: `
declare global {
  namespace foo {
    namespace bar {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: `
declare namespace foo {
  namespace bar {
    namespace baz {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: `
export declare namespace foo {
  export namespace bar {
    namespace baz {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
    },
    {
      code: 'namespace foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: true,
        },
      ],
    },
    {
      code: 'module foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: true,
        },
      ],
    },
  ],
  invalid: [
    {
      code: 'module foo {}',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'namespace foo {}',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'module foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'namespace foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'module foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'namespace foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'declare module foo {}',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'declare namespace foo {}',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'declare module foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'declare namespace foo {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'namespace foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'module foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'declare module foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'declare namespace foo {}',
      filename: 'test.d.ts',
      options: [
        {
          allowDefinitionFiles: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'namespace Foo.Bar {}',
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: false,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
namespace Foo.Bar {
  namespace Baz.Bas {
    interface X {}
  }
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
namespace A {
  namespace B {
    declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
namespace A {
  namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
namespace A {
  declare namespace B {
    namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
namespace A {
  export declare namespace B {
    namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
namespace A {
  export declare namespace B {
    declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
namespace A {
  export declare namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
namespace A {
  declare namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
namespace A {
  export namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 1,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 10,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
export namespace A {
  namespace B {
    declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
export namespace A {
  namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
export namespace A {
  declare namespace B {
    namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
export namespace A {
  export declare namespace B {
    namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
export namespace A {
  export declare namespace B {
    declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
export namespace A {
  export declare namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
export namespace A {
  declare namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
      ],
    },
    {
      code: `
export namespace A {
  export namespace B {
    export declare namespace C {}
  }
}
      `,
      filename: 'file.ts',
      options: [
        {
          allowDeclarations: true,
        },
      ],
      errors: [
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 2,
        },
        {
          messageId: 'moduleSyntaxIsPreferred',
          message: 'ES2015 module syntax is preferred over namespaces.',
          line: 3,
          column: 10,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
  ],
});
