import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';
import { noFormat } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2018,
    tsconfigRootDir: rootPath,
    project: './tsconfig.json',
  },
});

ruleTester.run('no-empty-interface', null, {
  valid: [
    `
interface Foo {
  name: string;
}
    `,
    `
interface Foo {
  name: string;
}

interface Bar {
  age: number;
}

// valid because extending multiple interfaces can be used instead of a union type
interface Baz extends Foo, Bar {}
    `,
    {
      code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
      `,
      options: [{ allowSingleExtends: true }],
    },
    {
      code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
      `,
      options: [{ allowSingleExtends: true }],
    },
  ],
  invalid: [
    {
      code: 'interface Foo {}',
      errors: [
        {
          messageId: 'noEmpty',
          line: 1,
          column: 11,
        },
      ],
      output: null,
    },
    {
      code: noFormat`interface Foo extends {}`,
      errors: [
        {
          messageId: 'noEmpty',
          line: 1,
          column: 11,
        },
      ],
      output: null,
    },
    {
      code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Baz {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 6,
          column: 11,
        },
      ],
      options: [{ allowSingleExtends: false }],
      output: `
interface Foo {
  props: string;
}

type Bar = Foo

class Baz {}
      `,
    },
    {
      code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 6,
          column: 11,
        },
      ],
      options: [{ allowSingleExtends: false }],
      output: null,
    },
    {
      code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

const bar = class Bar {};
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 6,
          column: 11,
        },
      ],
      options: [{ allowSingleExtends: false }],
      output: `
interface Foo {
  props: string;
}

type Bar = Foo

const bar = class Bar {};
      `,
    },
    {
      code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 6,
          column: 11,
        },
      ],
      options: [{ allowSingleExtends: false }],
      output: `
interface Foo {
  name: string;
}

type Bar = Foo
      `,
    },
    {
      code: 'interface Foo extends Array<number> {}',
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 1,
          column: 11,
        },
      ],
      output: `type Foo = Array<number>`,
    },
    {
      code: 'interface Foo extends Array<number | {}> {}',
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 1,
          column: 11,
        },
      ],
      output: `type Foo = Array<number | {}>`,
    },
    {
      code: `
interface Bar {
  bar: string;
}
interface Foo extends Array<Bar> {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 5,
          column: 11,
        },
      ],
      output: `
interface Bar {
  bar: string;
}
type Foo = Array<Bar>
      `,
    },
    {
      code: `
type R = Record<string, unknown>;
interface Foo extends R {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 3,
          column: 11,
        },
      ],
      output: `
type R = Record<string, unknown>;
type Foo = R
      `,
    },
    {
      code: `
interface Foo<T> extends Bar<T> {}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 2,
          column: 11,
        },
      ],
      output: `
type Foo<T> = Bar<T>
      `,
    },
    {
      code: `
declare module FooBar {
  type Baz = typeof baz;
  export interface Bar extends Baz {}
}
      `,
      errors: [
        {
          messageId: 'noEmptyWithSuper',
          line: 4,
          column: 20,
          endColumn: 23,
          endLine: 4,
          suggestions: [
            {
              messageId: 'noEmptyWithSuper',
              output: `
declare module FooBar {
  type Baz = typeof baz;
  export type Bar = Baz
}
      `,
            },
          ],
        },
      ],
      filename: 'test.d.ts',
      // output matches input because a suggestion was made
      output: null,
    },
  ],
});