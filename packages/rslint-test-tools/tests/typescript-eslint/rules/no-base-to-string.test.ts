// All 275 valid and 81 invalid cases from typescript-eslint v8.28.0.
// Original inputs, options, generators and messages are preserved; complete
// diagnostic ranges are additionally pinned from ESLint at the same version.
// This version has no autofixes or suggestions; the shared tester asserts their absence.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.28.0/packages/eslint-plugin/tests/rules/no-base-to-string.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': `{
  "compilerOptions": {
    "jsx": "preserve",
    "target": "es5",
    "module": "commonjs",
    "strict": true,
    "esModuleInterop": true,
    "lib": ["es2015", "es2017", "esnext"],
    "experimentalDecorators": true
  },
  "include": [
    "file.ts",
    "consistent-type-exports.ts",
    "deprecated.ts",
    "mixed-enums-decl.ts",
    "react.tsx",
    "var-declaration.ts",
    "errors.ts",
    "switch-exhaustiveness-check.ts"
  ]
}
`,
    'tsconfig.noUncheckedIndexedAccess.json': `{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "noUncheckedIndexedAccess": true
  }
}
`,
  },
  languageOptions: { parserOptions: { project: './tsconfig.json' } },
});

const literalListBasic: string[] = [
  "''",
  "'text'",
  'true',
  'false',
  '1',
  '1n',
  '[]',
  '/regex/',
];

const literalListNeedParen: string[] = [
  "__dirname === 'foobar'",
  '{}.constructor()',
  '() => {}',
  'function() {}',
];

const literalList = [...literalListBasic, ...literalListNeedParen];

const literalListWrapped = [
  ...literalListBasic,
  ...literalListNeedParen.map(i => `(${i})`),
];

ruleTester.run('no-base-to-string', {
  valid: [
    // template
    ...literalList.map(i => `\`\${${i}}\`;`),

    // operator + +=
    ...literalListWrapped.flatMap(l =>
      literalListWrapped.map(r => `${l} + ${r};`),
    ),

    // toString()
    ...literalListWrapped.map(i => `${i === '1' ? `(${i})` : i}.toString();`),

    // variable toString() and template
    ...literalList.map(
      i => `
        let value = ${i};
        value.toString();
        let text = \`\${value}\`;
      `,
    ),

    // String()
    ...literalList.map(i => `String(${i});`),
    `
function someFunction() {}
someFunction.toString();
let text = \`\${someFunction}\`;
    `,
    `
function someFunction() {}
someFunction.toLocaleString();
let text = \`\${someFunction}\`;
    `,
    'unknownObject.toString();',
    'unknownObject.toLocaleString();',
    'unknownObject.someOtherMethod();',
    `
class CustomToString {
  toString() {
    return 'Hello, world!';
  }
}
'' + new CustomToString();
    `,
    `
const literalWithToString = {
  toString: () => 'Hello, world!',
};
'' + literalWithToString;
    `,
    `
const printer = (inVar: string | number | boolean) => {
  inVar.toString();
};
printer('');
printer(1);
printer(true);
    `,
    `
const printer = (inVar: string | number | boolean) => {
  inVar.toLocaleString();
};
printer('');
printer(1);
printer(true);
    `,
    'let _ = {} * {};',
    'let _ = {} / {};',
    'let _ = ({} *= {});',
    'let _ = ({} /= {});',
    'let _ = ({} = {});',
    'let _ = {} == {};',
    'let _ = {} === {};',
    'let _ = {} in {};',
    'let _ = {} & {};',
    'let _ = {} ^ {};',
    'let _ = {} << {};',
    'let _ = {} >> {};',
    `
function tag() {}
tag\`\${{}}\`;
    `,
    `
      function tag() {}
      tag\`\${{}}\`;
    `,
    `
      interface Brand {}
      function test(v: string & Brand): string {
        return \`\${v}\`;
      }
    `,
    "'' += new Error();",
    "'' += new URL();",
    "'' += new URLSearchParams();",
    `
Number(1);
    `,
    {
      code: 'String(/regex/);',
      options: [{ ignoredTypeNames: ['RegExp'] }],
    },
    {
      code: `
type Foo = { a: string } | { b: string };
declare const foo: Foo;
String(foo);
      `,
      options: [{ ignoredTypeNames: ['Foo'] }],
    },
    `
function String(value) {
  return value;
}
declare const myValue: object;
String(myValue);
    `,
    `
import { String } from 'foo';
String({});
    `,
    `
['foo', 'bar'].join('');
    `,

    `
([{ foo: 'foo' }, 'bar'] as string[]).join('');
    `,
    `
function foo<T extends string>(array: T[]) {
  return array.join();
}
    `,
    `
class Foo {
  toString() {
    return '';
  }
}
[new Foo()].join();
    `,
    `
class Foo {
  join() {}
}
const foo = new Foo();
foo.join();
    `,
    `
declare const array: string[];
array.join('');
    `,
    `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
array.join('');
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
array.join('');
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
array.join('');
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
tuple.join('');
    `,
    `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
tuple.join('');
    `,

    `
String(['foo', 'bar']);
    `,

    `
String([{ foo: 'foo' }, 'bar'] as string[]);
    `,
    `
function foo<T extends string>(array: T[]) {
  return String(array);
}
    `,
    `
class Foo {
  toString() {
    return '';
  }
}
String([new Foo()]);
    `,
    `
declare const array: string[];
String(array);
    `,
    `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
String(array);
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
String(array);
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
String(array);
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
String(tuple);
    `,
    `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
String(tuple);
    `,

    `
['foo', 'bar'].toString();
    `,

    `
([{ foo: 'foo' }, 'bar'] as string[]).toString();
    `,
    `
function foo<T extends string>(array: T[]) {
  return array.toString();
}
    `,
    `
class Foo {
  toString() {
    return '';
  }
}
[new Foo()].toString();
    `,
    `
declare const array: string[];
array.toString();
    `,
    `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
array.toString();
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
array.toString();
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
array.toString();
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
tuple.toString();
    `,
    `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
tuple.toString();
    `,

    `
\`\${['foo', 'bar']}\`;
    `,

    `
\`\${[{ foo: 'foo' }, 'bar'] as string[]}\`;
    `,
    `
function foo<T extends string>(array: T[]) {
  return \`\${array}\`;
}
    `,
    `
class Foo {
  toString() {
    return '';
  }
}
\`\${[new Foo()]}\`;
    `,
    `
declare const array: string[];
\`\${array}\`;
    `,
    `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
\`\${array}\`;
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
\`\${array}\`;
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
\`\${array}\`;
    `,
    `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
\`\${tuple}\`;
    `,
    `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
\`\${tuple}\`;
    `,

    // don't bother trying to interpret spread args.
    `
let objects = [{}, {}];
String(...objects);
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/8585
    `
type Constructable<Entity> = abstract new (...args: any[]) => Entity;

interface GuildChannel {
  toString(): \`<#\${string}>\`;
}

declare const foo: Constructable<GuildChannel & { bar: 1 }>;
class ExtendedGuildChannel extends foo {}
declare const bb: ExtendedGuildChannel;
bb.toString();
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/8585 with intersection order reversed.
    `
type Constructable<Entity> = abstract new (...args: any[]) => Entity;

interface GuildChannel {
  toString(): \`<#\${string}>\`;
}

declare const foo: Constructable<{ bar: 1 } & GuildChannel>;
class ExtendedGuildChannel extends foo {}
declare const bb: ExtendedGuildChannel;
bb.toString();
    `,
    `
function foo<T>(x: T) {
  String(x);
}
    `,
    `
declare const u: unknown;
String(u);
    `,
    `
type Value = string | Value[];
declare const v: Value;

String(v);
    `,
    `
type Value = (string | Value)[];
declare const v: Value;

String(v);
    `,
    `
type Value = Value[];
declare const v: Value;

String(v);
    `,
    `
type Value = [Value];
declare const v: Value;

String(v);
    `,
    `
declare const v: ('foo' | 'bar')[][];
String(v);
    `,
  ],
  invalid: [
    {
      code: '`${{}})`;',
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '({}).toString();',
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '({}).toLocaleString();',
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: "'' + {};",
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'String({});',
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: "'' += {};",
      errors: [
        {
          messageId: 'baseToString',
          message: "'{}' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString.toLocaleString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString + '';
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrString' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject.toLocaleString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject + '';
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'someObjectOrObject' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
        interface A {}
        interface B {}
        function test(intersection: A & B): string {
          return \`\${intersection}\`;
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'intersection' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
class Foo {
  foo: string;
}
declare const foo: string | Foo;
\`\${foo}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'foo' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 4,
          endLine: 6,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const foo: Bar | Foo;
\`\${foo}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'foo' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 4,
          endLine: 9,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const foo: Bar & Foo;
\`\${foo}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'foo' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 4,
          endLine: 9,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
        [{}, {}].join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for [{}, {}] will use Object's default stringification format ('[object Object]') when stringified.",
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        const array = [{}, {}];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class A {
          a: string;
        }
        [new A(), 'str'].join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for [new A(), 'str'] may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        tuple.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        tuple.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        tuple.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for tuple may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        tuple.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for tuple may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        tuple.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for tuple will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        const array = ['string', { foo: 'bar' }];
        array.join('');
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 14,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
        },
      },
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array.join();
        }
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for array may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },

    {
      code: `
        String([{}, {}]);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
        const array = [{}, {}];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class A {
          a: string;
        }
        String([new A(), 'str']);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 32,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 16,
          endLine: 9,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        String(tuple);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        String(tuple);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        String(tuple);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        String(tuple);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        String(tuple);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 16,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
        const array = ['string', { foo: 'bar' }];
        String(array);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 21,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
        },
      },
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return String(array);
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 25,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },

    {
      code: `
        [{}, {}].toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        const array = [{}, {}];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class A {
          a: string;
        }
        [new A(), 'str'].toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        tuple.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        tuple.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        tuple.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        tuple.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        tuple.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
        const array = ['string', { foo: 'bar' }];
        array.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 14,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
        },
      },
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array.toString();
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },

    {
      code: `
        \`\${[{}, {}]}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[{}, {}]' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
        const array = [{}, {}];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class A {
          a: string;
        }
        \`\${[new A(), 'str']}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'[new A(), 'str']' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 9,
          column: 12,
          endLine: 9,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        \`\${tuple}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        \`\${tuple}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        \`\${tuple}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        \`\${tuple}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        \`\${tuple}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'tuple' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        const array = ['string', { foo: 'bar' }];
        \`\${array}\`;
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 17,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noUncheckedIndexedAccess.json',
        },
      },
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return \`\${array}\`;
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },

    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          array[0].toString();
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'array[0]' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(value: T) {
          value.toString();
        }
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'value' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
type Bar = Record<string, string>;
declare const foo: Bar | string;
foo.toString();
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'foo' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array;
        }
        foo([{ foo: 'foo' }]).join();
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for foo([{ foo: 'foo' }]) will use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array;
        }
        foo([{ foo: 'foo' }, 'bar']).join();
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for foo([{ foo: 'foo' }, 'bar']) may use Object's default stringification format ('[object Object]') when stringified.",
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 37,
        },
      ],
    },
    {
      code: `
type Value = { foo: string } | Value[];
declare const v: Value;

String(v);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'v' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
type Value = ({ foo: string } | Value)[];
declare const v: Value;

String(v);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'v' may use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
type Value = [{ foo: string }, Value];
declare const v: Value;

String(v);
      `,
      errors: [
        {
          messageId: 'baseToString',
          message: "'v' will use Object's default stringification format ('[object Object]') when stringified.",
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
declare const v: { foo: string }[][];
v.join();
      `,
      errors: [
        {
          messageId: 'baseArrayJoin',
          message: "Using `join()` for v will use Object's default stringification format ('[object Object]') when stringified.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },
  ],
});
