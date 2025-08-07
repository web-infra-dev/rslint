import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';

import { getFixturesRootDir } from '../RuleTester';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-namespace', {
  valid: [
    // Regular module declaration (not namespace)
    `
      declare module "foo" {
        export const bar: string;
      }
    `,

    // Global module augmentation
    `
      declare global {
        interface Window {
          foo: string;
        }
      }
    `,

    // Ambient module declaration
    `
      declare module "bar" {
        export const baz: number;
      }
    `,

    // Declare namespace (allowed when allowDeclarations is true)
    {
      code: `
        declare namespace Test {
          export const value = 1;
        }
      `,
      options: [{ allowDeclarations: true }],
    },

    // Regular TypeScript code without namespaces
    `
      const value = 1;
      function test() {
        return value;
      }
      class Test {
        constructor() {}
      }
    `,

    // Module with exports (not namespace)
    `
      export const value = 1;
      export function test() {
        return value;
      }
    `,

    // Test array format options
    {
      code: `
        declare namespace Test {
          export const value = 1;
        }
      `,
      options: [{ allowDeclarations: true }],
    },

    // Test empty options object
    {
      code: `
        const value = 1;
      `,
      options: [{}],
    },

    // Test nil options
    `
      const value = 1;
    `,
  ],

  invalid: [
    // Basic namespace usage
    {
      code: `
        namespace Test {
          export const value = 1;
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Nested namespace
    {
      code: `
        namespace Outer {
          namespace Inner {
            export const value = 1;
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Namespace with interface
    {
      code: `
        namespace Test {
          export interface Config {
            value: string;
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Namespace with class
    {
      code: `
        namespace Test {
          export class MyClass {
            constructor() {}
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Namespace with function
    {
      code: `
        namespace Test {
          export function myFunction() {
            return "test";
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Declare namespace (not allowed by default)
    {
      code: `
        declare namespace Test {
          export const value = 1;
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Multiple namespaces
    {
      code: `
        namespace A {
          export const a = 1;
        }

        namespace B {
          export const b = 2;
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Namespace with complex content
    {
      code: `
        namespace Utils {
          export interface Options {
            debug?: boolean;
            timeout?: number;
          }

          export class Helper {
            static process(options: Options): void {
              // implementation
            }
          }

          export function validate(input: string): boolean {
            return input.length > 0;
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test allowDeclarations explicitly set to false
    {
      code: `
        declare namespace Test {
          export const value = 1;
        }
      `,
      options: [{ allowDeclarations: false }],
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test array options format but allowDeclarations false
    {
      code: `
        declare namespace Test {
          export const value = 1;
        }
      `,
      options: [{ allowDeclarations: false }],
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test mix of namespace and module declaration
    {
      code: `
        namespace Test {
          export const value = 1;
        }

        declare module "external" {
          export const externalValue = 2;
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test mix of namespace and global declaration
    {
      code: `
        namespace Test {
          export const value = 1;
        }

        declare global {
          interface GlobalInterface {
            prop: string;
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test deeply nested namespaces
    {
      code: `
        namespace Level1 {
          namespace Level2 {
            namespace Level3 {
              export const value = 1;
            }
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
        {
          messageId: 'noNamespace',
        },
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test namespace with type aliases
    {
      code: `
        namespace Types {
          export type StringOrNumber = string | number;
          export type Callback<T> = (value: T) => void;
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test namespace with enums
    {
      code: `
        namespace Constants {
          export enum Status {
            Active = "active",
            Inactive = "inactive"
          }
        }
      `,
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },

    // Test regular namespace should be reported even when allowDefinitionFiles is true
    {
      code: `
        namespace Test {
          export const value = 1;
        }
      `,
      options: [{ allowDefinitionFiles: true }],
      errors: [
        {
          messageId: 'noNamespace',
        },
      ],
    },
  ],
});
