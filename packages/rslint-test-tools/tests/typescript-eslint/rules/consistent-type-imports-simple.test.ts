import { noFormat, RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      emitDecoratorMetadata: false,
      experimentalDecorators: false,
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('@typescript-eslint/consistent-type-imports', {
  valid: [
    `
      import Foo from 'foo';
      const foo: Foo = new Foo();
    `,
    `
      import foo from 'foo';
      const foo: foo.Foo = foo.fn();
    `,
    `
      import { A, B } from 'foo';
      const foo: A = B();
      const bar = new A();
    `,
    `
      import Foo from 'foo';
    `,
    `
      import Foo from 'foo';
      type T<Foo> = Foo; // shadowing
    `,
    `
      import Foo from 'foo';
      function fn() {
        type Foo = {}; // shadowing
        let foo: Foo;
      }
    `,
    `
      import { A, B } from 'foo';
      const b = B;
    `,
    `
      import { A, B, C as c } from 'foo';
      const d = c;
    `,
    `
      import {} from 'foo'; // empty
    `,
    {
      code: `
let foo: import('foo');
let bar: import('foo').Bar;
      `,
      options: [{ disallowTypeAnnotations: false }],
    },
    {
      code: `
import Foo from 'foo';
let foo: Foo;
      `,
      options: [{ prefer: 'no-type-imports' }],
    },
    // type queries
    `
      import type Type from 'foo';

      type T = typeof Type;
      type T = typeof Type.foo;
    `,
    `
      import type { Type } from 'foo';

      type T = typeof Type;
      type T = typeof Type.foo;
    `,
  ],
  invalid: [
    {
      code: `
import { Foo } from 'foo';
let foo: Foo;
      `,
      errors: [
        {
          messageId: 'typeOverValue',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'fixToTypeImport',
              output: `
import type { Foo } from 'foo';
let foo: Foo;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
import Foo from 'foo';
let foo: Foo;
      `,
      errors: [
        {
          messageId: 'typeOverValue',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'fixToTypeImport',
              output: `
import type Foo from 'foo';
let foo: Foo;
      `,
            },
          ],
        },
      ],
    },
  ],
});