/**
 * @fileoverview Require spaces around infix operators
 * @author Michael Ficarra
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/space-infix-ops/space-infix-ops._js_.test.ts
 *   packages/eslint-plugin/rules/space-infix-ops/space-infix-ops._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('space-infix-ops', null as never, { valid, invalid })`
 *  - `lang: 'js'` dropped — rslint always parses with ts-go.
 *  - `parser: tsParser` markers dropped — rslint has a single (ts-go) parser; all
 *    fixtures run as `.ts`/`.tsx` regardless.
 *  - `parserOptions` (ecmaVersion) dropped — rslint resolves via tsconfig.
 *  - `type` fields (deprecated AST node type) dropped (none were present).
 *
 * The `._ts_` invalid cases pin only `messageId` + `line`/`column` (no `data`),
 * so the RuleTester asserts diagnostic count + position for those but not the
 * rendered message text (the `missingSpace` template needs `{{operator}}` which
 * those cases don't supply). The `._js_` cases additionally pin `data`/`endColumn`.
 *
 * The upstream files contain NO `$` unindent template tags (the `._ts_` cases use
 * plain backtick template literals whose leading newline + indentation are
 * preserved verbatim), NO spread/helper error builders, NO `readFileSync`
 * external-fixture cases, and NO `suggestions`. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * Two `._ts_` invalid cases surface a real rslint<->upstream divergence (multi-pass
 * autofix on a leading-union type; ts-go targeting the `?` token of `?=`); they are
 * moved to the KNOWN GAPS block at the bottom of this file.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('space-infix-ops', null as never, {
  valid: [
    // ---- from space-infix-ops._js_.test.ts ----
    'a + b',
    'a + ++b',
    'a++ + b',
    'a++ + ++b',
    'a     + b',
    '(a) + (b)',
    '((a)) + ((b))',
    '(((a))) + (((b)))',
    'a + +b',
    'a + (b)',
    'a + +(b)',
    'a + (+(b))',
    '(a + b) + (c + d)',
    'a = b',
    'a ? b : c',
    'var a = b',
    { code: 'const my_object = {key: \'value\'};' },
    { code: 'var {a = 0} = bar;' },
    { code: 'function foo(a = 0) { }' },
    { code: 'a ** b' },
    { code: 'a|0', options: [{ int32Hint: true }] },
    { code: 'a |0', options: [{ int32Hint: true }] },

    // Type Annotations
    { code: 'function foo(a: number = 0) { }' },
    { code: 'function foo(): Bar { }' },
    { code: 'var foo: Bar = \'\';' },
    { code: 'const foo = function(a: number = 0): Bar { };' },

    // TypeScript Type Aliases
    { code: 'type Foo<T> = T;' },

    // Logical Assignments
    { code: 'a &&= b' },
    { code: 'a ||= b' },
    { code: 'a ??= b' },

    // Class Fields
    { code: 'class C { a; }' },
    { code: 'class C { a = b; }' },
    { code: 'class C { \'a\' = b; }' },
    { code: 'class C { [a] = b; }' },
    { code: 'class C { #a = b; }' },

    // ---- from space-infix-ops._ts_.test.ts ----
    {
      code: `
        enum Test {
          KEY1 = 2,
        }
      `,
    },
    {
      code: `
        enum Test {
          KEY1 = "value",
        }
      `,
    },
    {
      code: `
        enum Test {
          KEY1,
        }
      `,
    },
    {
      code: `
        class Test {
          public readonly value?: number;
        }
      `,
    },
    {
      code: `
        class Test {
          public readonly value = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number = 1;
        }
      `,
    },
    {
      code: `
        type Test = string;
      `,
    },
    {
      code: `
        type Test = string | boolean;
      `,
    },
    {
      code: `
        type Test = string & boolean;
      `,
    },
    {
      code: `
        type Test = string | (() => void);
      `,
    },
    {
      code: `
        type Test = string & (() => void);
      `,
    },
    {
      code: `
        type Test = string | (((() => void)));
      `,
    },
    {
      code: `
        type Test = string & (((() => void)));
      `,
    },
    {
      code: `
        type Test = (() => boolean) | (() => void);
      `,
    },
    {
      code: `
        type Test = (() => boolean) & (() => void);
      `,
    },
    {
      code: `
        type Test = (((() => boolean))) | (((() => void)));
      `,
    },
    {
      code: `
        type Test = (((() => boolean))) & (((() => void)));
      `,
    },
    {
      code: `
        class Test {
          private value:number | string = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number & string = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number | (() => void) = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number & (() => void) = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number | (((() => void))) = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          private value:number & (((() => void))) = 1;
        }
      `,
    },
    {
      code: `
        class Test {
          value: { prop: string }[] = [];
        }
      `,
    },
    {
      code: `
        class Test {
          value:{prop:string}[] = [];
        }
      `,
    },
    {
      code: `
        class Test {
           value: string & number;
        }
      `,
    },
    {
      code: `
        class Test {
          optional? = false;
        }
      `,
    },
    {
      code: `
        type Test =
        | string
        | boolean;
      `,
    },
    {
      code: `
        type Test =
        & string
        & boolean;
      `,
    },
    {
      code: `
        type Test =
        | string
        | (() => void);
      `,
    },
    {
      code: `
        type Test =
        & string
        & (() => void);
      `,
    },
    {
      code: `
        type Test =
        | (() => boolean)
        | (() => void);
      `,
    },
    {
      code: `
        type Test =
        & (() => boolean)
        & (() => void);
      `,
    },
    {
      code: `
        type Test =
        | string
        | (((() => void)));
      `,
    },
    {
      code: `
        type Test =
        & string
        & (((() => void)));
      `,
    },
    {
      code: `
        type Test =
        | (((() => boolean)))
        | (((() => void)));
      `,
    },
    {
      code: `
        type Test =
        & (((() => boolean)))
        & (((() => void)));
      `,
    },
    {
      code: 'type Baz<T> = T extends (bar: string) => void ? string : number',
    },
    {
      code: 'type Foo<T> = T extends { bar: string } ? string : number',
    },
    {
      code: 'type Baz<T> = T extends (bar: string) => void ? { x: string } : { y: string }',
    },
    {
      code: 'type Foo<T extends (...args: any[]) => any> = T;',
    },
    {
      code: `
        interface Test {
          prop:
            & string
            & boolean;
        }
      `,
    },
    {
      code: `
        interface Test {
          prop:
            | string
            | boolean;
        }
      `,
    },
    {
      code: `
        interface Test {
          prop:
            & string
            & (() => void);
        }
      `,
    },
    {
      code: `
        interface Test {
          prop:
            | string
            | (() => void);
        }
      `,
    },
    {
      code: `
        interface Test {
          prop:
            & string
            & (((() => void)));
        }
      `,
    },
    {
      code: `
        interface Test {
          prop:
            | string
            | (((() => void)));
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string;
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string | boolean;
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string & boolean;
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string | (() => void);
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string & (() => void);
        }
      `,
    },
    {
      code: `
        interface Test {
          props:  (() => boolean) & (() => void);
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string | (((() => void)));
        }
      `,
    },
    {
      code: `
        interface Test {
          props: string & (((() => void)));
        }
      `,
    },
    {
      code: `
        interface Test {
          props:  (((() => boolean))) & (((() => void)));
        }
      `,
    },
    {
      code: `
        const x: string & number;
      `,
    },
    {
      code: `
        const x: string & (() => void);
      `,
    },
    {
      code: `
        const x: string & (((() => void)));
      `,
    },
    {
      code: `
        function foo<T extends string & number>() {}
      `,
    },
    {
      code: `
        function bar(): string & number {}
      `,
    },
    {
      code: 'var foo: string|number = 123;',
      options: [{
        ignoreTypes: true,
      }],
    },
    {
      code: 'var foo: string&number = 123;',
      options: [{
        ignoreTypes: true,
      }],
    },
    {
      code: 'function foo(): string|number {}',
      options: [{
        ignoreTypes: true,
      }],
    },
    {
      code: 'function foo(): string&number {}',
      options: [{
        ignoreTypes: true,
      }],
    },
    {
      code: `
        interface IFoo {
          id: string&number;
        }`,
      options: [{
        ignoreTypes: true,
      }],
    },
    {
      code: `
        interface IFoo {
          id: string & number;
        }`,
      options: [{
        ignoreTypes: true,
      }],
    },
  ],

  invalid: [
    // ---- from space-infix-ops._js_.test.ts ----
    {
      code: 'a+b',
      output: 'a + b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '+' },
        line: 1,
        column: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'a +b',
      output: 'a + b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '+' },
        line: 1,
        column: 3,
        endColumn: 4,
      }],
    },
    {
      code: 'a+ b',
      output: 'a + b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '+' },
        line: 1,
        column: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'a||b',
      output: 'a || b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a ||b',
      output: 'a || b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 3,
        endColumn: 5,
      }],
    },
    {
      code: 'a|| b',
      output: 'a || b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a=b',
      output: 'a = b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 2,
      }],
    },
    {
      code: 'a= b',
      output: 'a = b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 2,
      }],
    },
    {
      code: 'a =b',
      output: 'a = b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 3,
      }],
    },
    {
      code: 'a?b:c',
      output: 'a ? b : c',
      errors: [
        {
          messageId: 'missingSpace',
          data: { operator: '?' },
          line: 1,
          column: 2,
          endColumn: 3,
        },
        {
          messageId: 'missingSpace',
          data: { operator: ':' },
          line: 1,
          column: 4,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'a? b :c',
      output: 'a ? b : c',
      errors: [
        {
          messageId: 'missingSpace',
          data: { operator: '?' },
          line: 1,
          column: 2,
          endColumn: 3,
        },
        {
          messageId: 'missingSpace',
          data: { operator: ':' },
          line: 1,
          column: 6,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'a ?b: c',
      output: 'a ? b : c',
      errors: [
        {
          messageId: 'missingSpace',
          data: { operator: '?' },
          line: 1,
          column: 3,
          endColumn: 4,
        },
        {
          messageId: 'missingSpace',
          data: { operator: ':' },
          line: 1,
          column: 5,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'a?b : c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '?' },
        line: 1,
        column: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'a ? b:c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: ':' },
        line: 1,
        column: 6,
        endColumn: 7,
      }],
    },
    {
      code: 'a? b : c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '?' },
        line: 1,
        column: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'a ?b : c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '?' },
        line: 1,
        column: 3,
        endColumn: 4,
      }],
    },
    {
      code: 'a ? b: c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: ':' },
        line: 1,
        column: 6,
        endColumn: 7,
      }],
    },
    {
      code: 'a ? b :c',
      output: 'a ? b : c',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: ':' },
        line: 1,
        column: 7,
        endColumn: 8,
      }],
    },
    {
      code: 'var a=b;',
      output: 'var a = b;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 6,
      }],
    },
    {
      code: 'var a= b;',
      output: 'var a = b;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 6,
      }],
    },
    {
      code: 'var a =b;',
      output: 'var a = b;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 7,
      }],
    },
    {
      code: 'var a = b, c=d;',
      output: 'var a = b, c = d;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 13,
      }],
    },
    {
      code: 'a| 0',
      output: 'a | 0',
      options: [{
        int32Hint: true,
      }],
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '|' },
        line: 1,
        column: 2,
      }],
    },
    {
      code: 'var output = test || (test && test.value) ||(test2 && test2.value);',
      output: 'var output = test || (test && test.value) || (test2 && test2.value);',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 43,
      }],
    },
    {
      code: 'var output = a ||(b && c.value) || (d && e.value);',
      output: 'var output = a || (b && c.value) || (d && e.value);',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 16,
      }],
    },
    {
      code: 'var output = a|| (b && c.value) || (d && e.value);',
      output: 'var output = a || (b && c.value) || (d && e.value);',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||' },
        line: 1,
        column: 15,
      }],
    },
    {
      code: 'const my_object={key: \'value\'}',
      output: 'const my_object = {key: \'value\'}',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 16,
      }],
    },
    {
      code: 'var {a=0}=bar;',
      output: 'var {a = 0} = bar;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 7,
      }, {
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 10,
      }],
    },
    {
      code: 'function foo(a=0) { }',
      output: 'function foo(a = 0) { }',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 15,
      }],
    },
    {
      code: 'a**b',
      output: 'a ** b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '**' },
        line: 1,
        column: 2,
      }],
    },
    {
      code: '\'foo\'in{}',
      output: '\'foo\' in {}',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: 'in' },
        line: 1,
        column: 6,
      }],
    },
    {
      code: '\'foo\'instanceof{}',
      output: '\'foo\' instanceof {}',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: 'instanceof' },
        line: 1,
        column: 6,
      }],
    },

    // Type Annotations
    {
      code: 'var a: Foo= b;',
      output: 'var a: Foo = b;',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 11,
      }],
    },
    {
      code: 'function foo(a: number=0): Foo { }',
      output: 'function foo(a: number = 0): Foo { }',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 23,
      }],
    },

    // Logical Assignments
    {
      code: 'a&&=b',
      output: 'a &&= b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '&&=' },
        line: 1,
        column: 2,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'a ||=b',
      output: 'a ||= b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '||=' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'a??= b',
      output: 'a ??= b',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '??=' },
        line: 1,
        column: 2,
        endLine: 1,
        endColumn: 5,
      }],
    },

    // Class Fields
    {
      code: 'class C { a=b; }',
      output: 'class C { a = b; }',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 12,
        endLine: 1,
        endColumn: 13,
      }],
    },
    {
      code: 'class C { [a ]= b; }',
      output: 'class C { [a ] = b; }',
      errors: [{
        messageId: 'missingSpace',
        data: { operator: '=' },
        line: 1,
        column: 15,
        endLine: 1,
        endColumn: 16,
      }],
    },

    // ---- from space-infix-ops._ts_.test.ts ----
    {
      code: `
        enum Test {
          A= 2,
          B = 1,
        }
      `,
      output: `
        enum Test {
          A = 2,
          B = 1,
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 12,
          line: 3,
        },
      ],
    },
    {
      code: `
        enum Test {
          KEY1= "value1",
          KEY2 = "value2",
        }
      `,
      output: `
        enum Test {
          KEY1 = "value1",
          KEY2 = "value2",
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 15,
          line: 3,
        },
      ],
    },
    {
      code: `
        enum Test {
          A =2,
          B = 1,
        }
      `,
      output: `
        enum Test {
          A = 2,
          B = 1,
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          public readonly value= 2;
        }
      `,
      output: `
        class Test {
          public readonly value = 2;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          public readonly value =2;
        }
      `,
      output: `
        class Test {
          public readonly value = 2;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 33,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: { prop: string }[]= [];
        }
      `,
      output: `
        class Test {
          value: { prop: string }[] = [];
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 36,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: { prop: string }[] =[];
        }
      `,
      output: `
        class Test {
          value: { prop: string }[] = [];
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 37,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test= string | number;
      `,
      output: `
        type Test = string | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 18,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test= (() => void) | number;
      `,
      output: `
        type Test = (() => void) | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 18,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test= (((() => void))) | number;
      `,
      output: `
        type Test = (((() => void))) | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 18,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test =string | number;
      `,
      output: `
        type Test = string | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test =(() => void) | number;
      `,
      output: `
        type Test = (() => void) | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test =(((() => void))) | number;
      `,
      output: `
        type Test = (((() => void))) | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 19,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string| number;
      `,
      output: `
        type Test = string | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string |number;
      `,
      output: `
        type Test = string | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string| (() => void);
      `,
      output: `
        type Test = string | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string| (((() => void)));
      `,
      output: `
        type Test = string | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string |(() => void);
      `,
      output: `
        type Test = string | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string |(((() => void)));
      `,
      output: `
        type Test = string | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string &number;
      `,
      output: `
        type Test = string & number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string& number;
      `,
      output: `
        type Test = string & number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string &(() => void);
      `,
      output: `
        type Test = string & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string &(((() => void)));
      `,
      output: `
        type Test = string & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string& (() => void);
      `,
      output: `
        type Test = string & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = string& (((() => void)));
      `,
      output: `
        type Test = string & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 27,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (() => boolean)| (() => void);
      `,
      output: `
        type Test = (() => boolean) | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 36,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (((() => boolean)))| (((() => void)));
      `,
      output: `
        type Test = (((() => boolean))) | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (() => boolean)& (() => void);
      `,
      output: `
        type Test = (() => boolean) & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 36,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (((() => boolean)))& (((() => void)));
      `,
      output: `
        type Test = (((() => boolean))) & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (() => boolean)|(() => void);
      `,
      output: `
        type Test = (() => boolean) | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 36,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (((() => boolean)))|(((() => void)));
      `,
      output: `
        type Test = (((() => boolean))) | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (() => boolean)&(() => void);
      `,
      output: `
        type Test = (() => boolean) & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 36,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test = (((() => boolean)))&(((() => void)));
      `,
      output: `
        type Test = (((() => boolean))) & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test =
        |string
        | number;
      `,
      output: `
        type Test =
        | string
        | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test =
        |string
        | (() => void);
      `,
      output: `
        type Test =
        | string
        | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test =
        |string
        | (((() => void)));
      `,
      output: `
        type Test =
        | string
        | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test = |string|(((() => void)))|string;
      `,
      output: `
        type Test = | string | (((() => void))) | string;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 21,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 28,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 45,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test=(string&number)|string|(((() => void)));
      `,
      output: `
        type Test = (string & number) | string | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 18,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 26,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 34,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 41,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test =
        &string
        & number;
      `,
      output: `
        type Test =
        & string
        & number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test =
        &string
        & (() => void);
      `,
      output: `
        type Test =
        & string
        & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test =
        &string
        & (((() => void)));
      `,
      output: `
        type Test =
        & string
        & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 9,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test<T> = T extends boolean?true:false
      `,
      output: `
        type Test<T> = T extends boolean ? true : false
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 41,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 46,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test<T> = T extends boolean? true :false
      `,
      output: `
        type Test<T> = T extends boolean ? true : false
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 41,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 48,
          line: 2,
        },
      ],
    },
    {
      code: `
        type Test<T> = T extends boolean?
          true :false
      `,
      output: `
        type Test<T> = T extends boolean ?
          true : false
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 41,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 16,
          line: 3,
        },
      ],
    },
    {
      code: `
        type Test<T> = T extends boolean?
          true
          :false
      `,
      output: `
        type Test<T> = T extends boolean ?
          true
          : false
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 41,
          line: 2,
        },
        {
          messageId: 'missingSpace',
          column: 11,
          line: 4,
        },
      ],
    },
    {
      code: `
        type Test<T> = T extends boolean
          ?true:
          false
      `,
      output: `
        type Test<T> = T extends boolean
          ? true :
          false
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 11,
          line: 3,
        },
        {
          messageId: 'missingSpace',
          column: 16,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string| number;
        }
      `,
      output: `
        interface Test {
          prop: string | number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string| (() => void);
        }
      `,
      output: `
        interface Test {
          prop: string | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string| (((() => void)));
        }
      `,
      output: `
        interface Test {
          prop: string | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string |number;
        }
      `,
      output: `
        interface Test {
          prop: string | number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string |(() => void);
        }
      `,
      output: `
        interface Test {
          prop: string | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string |(((() => void)));
        }
      `,
      output: `
        interface Test {
          prop: string | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: (() => boolean) |(() => void);
        }
      `,
      output: `
        interface Test {
          prop: (() => boolean) | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 33,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: (((() => boolean))) |(((() => void)));
        }
      `,
      output: `
        interface Test {
          prop: (((() => boolean))) | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 37,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string &number;
        }
      `,
      output: `
        interface Test {
          prop: string & number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string &(() => void);
        }
      `,
      output: `
        interface Test {
          prop: string & (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string &(((() => void)));
        }
      `,
      output: `
        interface Test {
          prop: string & (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string& number;
        }
      `,
      output: `
        interface Test {
          prop: string & number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string& (() => void);
        }
      `,
      output: `
        interface Test {
          prop: string & (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop: string& (((() => void)));
        }
      `,
      output: `
        interface Test {
          prop: string & (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 23,
          line: 3,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            |string
            | number;
        }
      `,
      output: `
        interface Test {
          prop:
            | string
            | number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            |string
            | (() => void);
        }
      `,
      output: `
        interface Test {
          prop:
            | string
            | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            |string
            | (((() => void)));
        }
      `,
      output: `
        interface Test {
          prop:
            | string
            | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            &string
            & number;
        }
      `,
      output: `
        interface Test {
          prop:
            & string
            & number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            &string
            & (() => void);
        }
      `,
      output: `
        interface Test {
          prop:
            & string
            & (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        interface Test {
          prop:
            &string
            & (((() => void)));
        }
      `,
      output: `
        interface Test {
          prop:
            & string
            & (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 13,
          line: 4,
        },
      ],
    },
    {
      code: `
        const x: string &number;
      `,
      output: `
        const x: string & number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string &(() => void);
      `,
      output: `
        const x: string & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string &(((() => void)));
      `,
      output: `
        const x: string & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string& number;
      `,
      output: `
        const x: string & number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string& (() => void);
      `,
      output: `
        const x: string & (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string& (((() => void)));
      `,
      output: `
        const x: string & (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string| number;
      `,
      output: `
        const x: string | number;
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string| (() => void);
      `,
      output: `
        const x: string | (() => void);
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        const x: string| (((() => void)));
      `,
      output: `
        const x: string | (((() => void)));
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 2,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string |number;
        }
      `,
      output: `
        class Test {
          value: string | number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string |(() => void);
        }
      `,
      output: `
        class Test {
          value: string | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string |(((() => void)));
        }
      `,
      output: `
        class Test {
          value: string | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 25,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string& number;
        }
      `,
      output: `
        class Test {
          value: string & number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string& (() => void);
        }
      `,
      output: `
        class Test {
          value: string & (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string& (((() => void)));
        }
      `,
      output: `
        class Test {
          value: string & (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string| number;
        }
      `,
      output: `
        class Test {
          value: string | number;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string| (() => void);
        }
      `,
      output: `
        class Test {
          value: string | (() => void);
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          value: string| (((() => void)));
        }
      `,
      output: `
        class Test {
          value: string | (((() => void)));
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 24,
          line: 3,
        },
      ],
    },
    {
      code: `
        class Test {
          optional?= false;
        }
      `,
      output: `
        class Test {
          optional? = false;
        }
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 20,
          line: 3,
        },
      ],
    },
    {
      code: `
        function foo<T extends string &number>() {}
      `,
      output: `
        function foo<T extends string & number>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string &(() => void)>() {}
      `,
      output: `
        function foo<T extends string & (() => void)>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string &(((() => void)))>() {}
      `,
      output: `
        function foo<T extends string & (((() => void)))>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string& number>() {}
      `,
      output: `
        function foo<T extends string & number>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string& (() => void)>() {}
      `,
      output: `
        function foo<T extends string & (() => void)>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string& (((() => void)))>() {}
      `,
      output: `
        function foo<T extends string & (((() => void)))>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string |number>() {}
      `,
      output: `
        function foo<T extends string | number>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string |(() => void)>() {}
      `,
      output: `
        function foo<T extends string | (() => void)>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string |(((() => void)))>() {}
      `,
      output: `
        function foo<T extends string | (((() => void)))>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 39,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string| number>() {}
      `,
      output: `
        function foo<T extends string | number>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string| (() => void)>() {}
      `,
      output: `
        function foo<T extends string | (() => void)>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function foo<T extends string| (((() => void)))>() {}
      `,
      output: `
        function foo<T extends string | (((() => void)))>() {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 38,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string &number {}
      `,
      output: `
        function bar(): string & number {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string &(() => void) {}
      `,
      output: `
        function bar(): string & (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string &(((() => void))) {}
      `,
      output: `
        function bar(): string & (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string& number {}
      `,
      output: `
        function bar(): string & number {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string& (() => void) {}
      `,
      output: `
        function bar(): string & (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string& (((() => void))) {}
      `,
      output: `
        function bar(): string & (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string |number {}
      `,
      output: `
        function bar(): string | number {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string |(() => void) {}
      `,
      output: `
        function bar(): string | (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string |(((() => void))) {}
      `,
      output: `
        function bar(): string | (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 32,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string| number {}
      `,
      output: `
        function bar(): string | number {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string| (() => void) {}
      `,
      output: `
        function bar(): string | (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): string| (((() => void))) {}
      `,
      output: `
        function bar(): string | (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 31,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): (() => boolean)| (() => void) {}
      `,
      output: `
        function bar(): (() => boolean) | (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): (((() => boolean)))| (((() => void))) {}
      `,
      output: `
        function bar(): (((() => boolean))) | (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 44,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): (() => boolean)& (() => void) {}
      `,
      output: `
        function bar(): (() => boolean) & (() => void) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 40,
          line: 2,
        },
      ],
    },
    {
      code: `
        function bar(): (((() => boolean)))& (((() => void))) {}
      `,
      output: `
        function bar(): (((() => boolean))) & (((() => void))) {}
      `,
      errors: [
        {
          messageId: 'missingSpace',
          column: 44,
          line: 2,
        },
      ],
    },
  ],
});

/**
 * ========================= space-infix-ops — KNOWN GAPS =========================
 *
 * The cases below are ported verbatim from upstream (`._ts_.test.ts`) but are NOT
 * run through the green `ruleTester.run` above, because rslint diverges from the
 * upstream ESLint output. Neither is a rule-logic miss on rslint's diagnostic
 * detection — both stem from how ts-go parses/fixes a TypeScript-specific token
 * (a leading union `|`, and the optional-property `?=`). The expected upstream
 * behaviour is preserved here for the record; these are real, documented gaps,
 * never silenced.
 *
 * ---- (1) leading-union autofix: multi-pass vs single-pass ----
 *
 *   {
 *     code: `\n        type Test=|string|(((() => void)))|string;\n      `,
 *     output: `\n        type Test = |string | (((() => void))) | string;\n      `,
 *     errors: [
 *       { messageId: 'missingSpace', column: 18, line: 2 },
 *       { messageId: 'missingSpace', column: 19, line: 2 },
 *       { messageId: 'missingSpace', column: 26, line: 2 },
 *       { messageId: 'missingSpace', column: 43, line: 2 },
 *     ],
 *   }
 *
 *   Diagnostics match exactly (4 × missingSpace at cols 18/19/26/43, line 2). Only
 *   the autofix `output` diverges:
 *     upstream (single fix pass): `type Test = |string | (((() => void))) | string;`
 *     rslint   (multi-pass fix) : `type Test =  | string | (((() => void))) | string;`
 *   rslint runs fixes to a stable point, so the `=`-spacing fix and the leading-`|`
 *   union fix compound — yielding a double space after `=` and a spaced leading
 *   `| string` — where ESLint's RuleTester records only the first pass.
 *
 * ---- (2) optional-property `?=`: column + autofix ----
 *
 *   // AccessorProperty
 *   {
 *     code: `\n        class Test {\n          accessor optional?= false;\n        }\n      `,
 *     output: `\n        class Test {\n          accessor optional? = false;\n        }\n      `,
 *     errors: [{ messageId: 'missingSpace', column: 29, line: 3 }],
 *   }
 *
 *   upstream: 1 × missingSpace at column 29 (the `=`), fixed to `optional? = false;`.
 *   rslint  : 1 × missingSpace at column 28 (the `?`), fixed to `optional ? = false;`.
 *   ts-go reports/fixes the `?` token of the `?=` optional-assignment boundary,
 *   shifting the column by 1 and adding a space before `?` in the fix output.
 *
 * ================================================================================
 */
