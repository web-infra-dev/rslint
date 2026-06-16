/**
 * @fileoverview Tests for lines-between-class-members rule.
 * @author 薛定谔的猫<hh_2013@foxmail.com>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/lines-between-class-members/lines-between-class-members._js_.test.ts
 *   packages/eslint-plugin/rules/lines-between-class-members/lines-between-class-members._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('lines-between-class-members', null as never, { valid, invalid })`
 *  - The local error helpers `alwaysError` / `neverError` (`{ messageId: 'always' | 'never' }`)
 *    are inlined to their final object form.
 *  - The `$` unindent template tag (used only in the _ts_ file) is evaluated to
 *    its real multi-line string. The _js_ file uses plain backtick templates whose
 *    literal indentation is significant (error `column` values are computed against
 *    it) and is preserved byte-for-byte.
 *  - `type` fields: none present upstream, nothing dropped.
 *
 * The _js_ cases come first, then the _ts_ cases. No Babel/Flow cases and no
 * `if (!skipBabel)` block exist for this rule. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * Every invalid case upstream pins both `errors` and `output`, so there are NO
 * output-only cases. No case surfaces a parser-level (syntax) incompatibility.
 * Exactly ONE case is isolated into the `KNOWN GAPS` block at the bottom — a
 * fix-application difference (rslint's multi-pass `--fix` fixpoint vs ESLint's
 * single-pass RuleTester `output`), annotated there with both behaviours.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('lines-between-class-members', null as never, {
  valid: [
    // ---- from lines-between-class-members._js_.test.ts ----
    'class foo{}',
    'class foo{;;}',
    'class foo{\n\n}',
    'class foo{constructor(){}\n}',
    'class foo{\nconstructor(){}}',

    'class foo{ bar(){}\n\nbaz(){}}',
    'class foo{ bar(){}\n\n/*comments*/baz(){}}',
    'class foo{ bar(){}\n\n//comments\nbaz(){}}',
    'class foo{ bar(){}\n//comments\n\nbaz(){}}',
    'class A{ foo() {} // a comment\n\nbar() {}}',
    'class A{ foo() {}\n/* a */ /* b */\n\nbar() {}}',
    'class A{ foo() {}/* a */ \n\n /* b */bar() {}}',

    'class A {\nfoo() {}\n/* comment */;\n;\n\nbar() {}\n}',
    'class A {\nfoo() {}\n// comment\n\n;\n;\nbar() {}\n}',

    'class foo{ bar(){}\n\n;;baz(){}}',
    'class foo{ bar(){};\n\nbaz(){}}',

    'class C {\naaa;\n\n#bbb;\n\nccc(){}\n\n#ddd(){}\n}',

    { code: 'class foo{ bar(){}\nbaz(){}}', options: ['never'] },
    {
      code: 'class foo{ bar(){}\n/*comments*/baz(){}}',
      options: ['never'],
    },
    {
      code: 'class foo{ bar(){}\n//comments\nbaz(){}}',
      options: ['never'],
    },
    {
      code: 'class foo{ bar(){}/* comments\n\n*/baz(){}}',
      options: ['never'],
    },
    {
      code: 'class foo{ bar(){}/* \ncomments\n*/baz(){}}',
      options: ['never'],
    },
    {
      code: 'class foo{ bar(){}\n/* \ncomments\n*/\nbaz(){}}',
      options: ['never'],
    },

    { code: 'class foo{ bar(){}\n\nbaz(){}}', options: ['always'] },
    {
      code: 'class foo{ bar(){}\n\n/*comments*/baz(){}}',
      options: ['always'],
    },
    {
      code: 'class foo{ bar(){}\n\n//comments\nbaz(){}}',
      options: ['always'],
    },

    {
      code: 'class foo{ bar(){}\nbaz(){}}',
      options: ['always', { exceptAfterSingleLine: true }],
    },
    {
      code: 'class foo{ bar(){\n}\n\nbaz(){}}',
      options: ['always', { exceptAfterSingleLine: true }],
    },
    {
      code: 'class foo{\naaa;\n#bbb;\nccc(){\n}\n\n#ddd(){\n}\n}',
      options: ['always', { exceptAfterSingleLine: true }],
    },

    // semicolon-less style (semicolons are at the beginning of lines)
    { code: 'class C { foo\n\n;bar }', options: ['always'] },
    {
      code: 'class C { foo\n;bar }',
      options: ['always', { exceptAfterSingleLine: true }],
    },
    { code: 'class C { foo\n;bar }', options: ['never'] },

    // enforce option with blankLine: "always"
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: '*' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: 'method' },
          ],
        },
      ],
    },

    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: '*' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: '*', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        { enforce: [{ blankLine: 'always', prev: '*', next: '*' }] },
      ],
    },

    // enforce option - blankLine: "never"
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: '*' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'field', next: 'method' },
          ],
        },
      ],
    },

    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'field', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [{ blankLine: 'never', prev: 'field', next: '*' }],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: '*', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [{ blankLine: 'never', prev: '*', next: 'field' }],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        { enforce: [{ blankLine: 'never', prev: '*', next: '*' }] },
      ],
    },

    // enforce option - multiple configurations
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods, disallows blank lines between fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'never', prev: 'field', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around fields, disallows blank lines between methods
          enforce: [
            { blankLine: 'always', prev: '*', next: 'field' },
            { blankLine: 'always', prev: 'field', next: '*' },
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }
                
                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }
                
                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'never', prev: '*', next: 'method' },
            { blankLine: 'never', prev: 'method', next: '*' },
            { blankLine: 'never', prev: 'field', next: 'field' },

            // This should take precedence over the above
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
    },

    // enforce with exceptAfterSingleLine option
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
        {
          exceptAfterSingleLine: true,
        },
      ],
    },
    // ---- from lines-between-class-members._ts_.test.ts ----
    {
      code: "class foo {\nbaz1() { }\n\nbaz2() { }\n\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nqux1() { }\n\nqux2() { }\n};",
      options: ['always'],
    },
    {
      code: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\n\nqux() { }\n};",
      options: ['always', { exceptAfterOverload: true }],
    },
    {
      code: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\nqux() { }\n};",
      options: ['always', { exceptAfterOverload: true, exceptAfterSingleLine: true }],
    },
    {
      code: "class foo{\nbar(a: string):void;\n\nbar(a: string, b:string):void;\n\nbar(a: string, b:string){\n\n}\n\nbaz() { }\n\nqux() { }\n};",
      options: ['always', { exceptAfterOverload: false, exceptAfterSingleLine: false }],
    },
    {
      code: "class foo {\nbar(a: string):void\nbar(a: string, b:string):void;\nbar(a: string, b:string){\n\n}\nbaz() { }\nqux() { }\n};",
      options: ['never', { exceptAfterOverload: true, exceptAfterSingleLine: true }],
    },
    {
      code: "class foo{\nbar(a: string):void\nbar(a: string, b:string):void;\nbar(a: string, b:string){\n\n}\nbaz() { }\nqux() { }\n};",
      options: ['never', { exceptAfterOverload: true, exceptAfterSingleLine: true }],
    },
    {
      code: "abstract class foo {\nabstract bar(a: string): void;\nabstract bar(a: string, b: string): void;\n};",
      options: ['always'],
    },
    // https://github.com/eslint-stylistic/eslint-stylistic/issues/240
    {
      code: "class foo {\n  bar(a: string): void;\n  bar(a: string, b:string): void;\n  bar(a: string, b:string) {\n\n  }\n\n  baz() { }\n\n  qux() { }\n};",
      options: [{ enforce: [{ blankLine: 'always', prev: 'method', next: 'method' }] }, { exceptAfterOverload: true }],
    },
  ],
  invalid: [
    // ---- from lines-between-class-members._js_.test.ts ----
    {
      code: 'class foo{ bar(){}\nbaz(){}}',
      output: 'class foo{ bar(){}\n\nbaz(){}}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class foo{ bar(){}\n\nbaz(){}}',
      output: 'class foo{ bar(){}\nbaz(){}}',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){\n}\nbaz(){}}',
      output: 'class foo{ bar(){\n}\n\nbaz(){}}',
      options: ['always', { exceptAfterSingleLine: true }],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class foo{ bar(){\n}\n/* comment */\nbaz(){}}',
      output: 'class foo{ bar(){\n}\n\n/* comment */\nbaz(){}}',
      options: ['always', { exceptAfterSingleLine: true }],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class foo{ bar(){}\n\n// comment\nbaz(){}}',
      output: 'class foo{ bar(){}\n// comment\nbaz(){}}',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n\n/* comment */\nbaz(){}}',
      output: 'class foo{ bar(){}\n/* comment */\nbaz(){}}',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n/* comment-1 */\n\n/* comment-2 */\nbaz(){}}',
      output: 'class foo{ bar(){}\n/* comment-1 */\n/* comment-2 */\nbaz(){}}',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n\n/* comment */\n\nbaz(){}}',
      output: null,
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n\n// comment\n\nbaz(){}}',
      output: null,
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n/* comment-1 */\n\n/* comment-2 */\n\n/* comment-3 */\nbaz(){}}',
      output: null,
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class foo{ bar(){}\n/* comment-1 */\n\n;\n\n/* comment-3 */\nbaz(){}}',
      output: null,
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class A {\nfoo() {}// comment\n;\n/* comment */\nbar() {}\n}',
      output: 'class A {\nfoo() {}// comment\n\n;\n/* comment */\nbar() {}\n}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class A {\nfoo() {}\n/* comment */;\n;\n/* comment */\nbar() {}\n}',
      output: 'class A {\nfoo() {}\n\n/* comment */;\n;\n/* comment */\nbar() {}\n}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class foo{ bar(){};\nbaz(){}}',
      output: 'class foo{ bar(){};\n\nbaz(){}}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class foo{ bar(){} // comment \nbaz(){}}',
      output: 'class foo{ bar(){} // comment \n\nbaz(){}}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class A {\nfoo() {}\n/* comment */;\n;\nbar() {}\n}',
      output: 'class A {\nfoo() {}\n\n/* comment */;\n;\nbar() {}\n}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class C {\nfield1\nfield2\n}',
      output: 'class C {\nfield1\n\nfield2\n}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class C {\n#field1\n#field2\n}',
      output: 'class C {\n#field1\n\n#field2\n}',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class C {\nfield1\n\nfield2\n}',
      output: 'class C {\nfield1\nfield2\n}',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class C {\nfield1 = () => {\n}\nfield2\nfield3\n}',
      output: 'class C {\nfield1 = () => {\n}\n\nfield2\nfield3\n}',
      options: ['always', { exceptAfterSingleLine: true }],
      errors: [{ messageId: 'always' }],
    },
    // NOTE: the `'class C { foo;bar }'` case is in KNOWN GAPS below — its upstream
    // `output` captures a SINGLE ESLint fix pass (`foo;\nbar`), whereas rslint's
    // `--fix` runs to a stable fixpoint and inserts the blank line in a second pass
    // (`foo;\n\nbar`). A documented multi-pass-vs-single-pass difference, not a bug.
    {
      code: 'class C { foo;\nbar; }',
      output: 'class C { foo;\n\nbar; }',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class C { foo;\n;bar }',
      output: 'class C { foo;\n\n;bar }',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },

    // semicolon-less style (semicolons are at the beginning of lines)
    {
      code: 'class C { foo\n;bar }',
      output: 'class C { foo\n\n;bar }',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },
    {
      code: 'class C { foo\n\n;bar }',
      output: 'class C { foo\n;bar }',
      options: ['never'],
      errors: [{ messageId: 'never' }],
    },
    {
      code: 'class C { foo\n;;bar }',
      output: 'class C { foo\n\n;;bar }',
      options: ['always'],
      errors: [{ messageId: 'always' }],
    },

    // enforce option with blankLine: "always"
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 11,
          column: 17,
        },
        {
          messageId: 'always',
          line: 14,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
        {
          messageId: 'always',
          line: 16,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'method', next: '*' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 11,
          column: 17,
        },
        {
          messageId: 'always',
          line: 14,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: 'field', next: '*' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
        {
          messageId: 'always',
          line: 10,
          column: 17,
        },
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'always', prev: '*', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        { enforce: [{ blankLine: 'always', prev: '*', next: '*' }] },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
        {
          messageId: 'always',
          line: 10,
          column: 17,
        },
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
      ],
    },

    // enforce option - blankLine: "never"
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                
                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 11,
          column: 17,
        },
        {
          messageId: 'never',
          line: 15,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 14,
          column: 17,
        },
        {
          messageId: 'never',
          line: 18,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
fieldA = 'Field A';
                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 8,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                
                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 8,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'method', next: '*' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 8,
          column: 17,
        },
        {
          messageId: 'never',
          line: 14,
          column: 17,
        },
        {
          messageId: 'never',
          line: 18,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'field', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 12,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
#fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'field', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 10,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
#fieldB = 'Field B';
method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: 'field', next: '*' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 10,
          column: 17,
        },
        {
          messageId: 'never',
          line: 12,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';
method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: '*', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 12,
          column: 17,
        },
        {
          messageId: 'never',
          line: 14,
          column: 17,
        },
        {
          messageId: 'never',
          line: 18,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
fieldA = 'Field A';
#fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      options: [
        {
          enforce: [
            { blankLine: 'never', prev: '*', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'never',
          line: 8,
          column: 17,
        },
        {
          messageId: 'never',
          line: 10,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
fieldA = 'Field A';
#fieldB = 'Field B';
method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        { enforce: [{ blankLine: 'never', prev: '*', next: '*' }] },
      ],
      errors: [
        {
          messageId: 'never',
          line: 8,
          column: 17,
        },
        {
          messageId: 'never',
          line: 10,
          column: 17,
        },
        {
          messageId: 'never',
          line: 12,
          column: 17,
        },
        {
          messageId: 'never',
          line: 14,
          column: 17,
        },
        {
          messageId: 'never',
          line: 18,
          column: 17,
        },
      ],
    },

    // enforce option - multiple configurations
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';

                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
#fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods, disallows blank lines between fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'never', prev: 'field', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'never',
          line: 9,
          column: 17,
        },
        {
          messageId: 'always',
          line: 10,
          column: 17,
        },
        {
          messageId: 'always',
          line: 11,
          column: 17,
        },
        {
          messageId: 'always',
          line: 14,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}
get area() {
                    return this.method1();
                }
method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around fields, disallows blank lines between methods
          enforce: [
            { blankLine: 'always', prev: '*', next: 'field' },
            { blankLine: 'always', prev: 'field', next: '*' },
            { blankLine: 'never', prev: 'method', next: 'method' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
        {
          messageId: 'never',
          line: 11,
          column: 17,
        },
        {
          messageId: 'never',
          line: 15,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
        {
          messageId: 'always',
          line: 10,
          column: 17,
        },
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
      ],
    },
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';

                #fieldB = 'Field B';

                method1() {}

                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'never', prev: '*', next: 'method' },
            { blankLine: 'never', prev: 'method', next: '*' },
            { blankLine: 'never', prev: 'field', next: 'field' },

            // This should take precedence over the above
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 8,
          column: 17,
        },
        {
          messageId: 'always',
          line: 9,
          column: 17,
        },
        {
          messageId: 'always',
          line: 10,
          column: 17,
        },
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
      ],
    },

    // enforce with exceptAfterSingleLine option
    {
      code: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }
                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }
                method2() {}
              }
            `,
      output: `
              class MyClass {
                constructor(height, width) {
                    this.height = height;
                    this.width = width;
                }

                fieldA = 'Field A';
                #fieldB = 'Field B';
                method1() {}
                get area() {
                    return this.method1();
                }

                method2() {}
              }
            `,
      options: [
        {

          // requires blank lines around methods and fields
          enforce: [
            { blankLine: 'always', prev: '*', next: 'method' },
            { blankLine: 'always', prev: 'method', next: '*' },
            { blankLine: 'always', prev: 'field', next: 'field' },
          ],
        },
        {
          exceptAfterSingleLine: true,
        },
      ],
      errors: [
        {
          messageId: 'always',
          line: 7,
          column: 17,
        },
        {
          messageId: 'always',
          line: 13,
          column: 17,
        },
      ],
    },
    // ---- from lines-between-class-members._ts_.test.ts ----
    {
      code: "class foo {\nbaz1() { }\nbaz2() { }\n\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nqux1() { }\nqux2() { }\n};",
      output: "class foo {\nbaz1() { }\n\nbaz2() { }\n\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nqux1() { }\n\nqux2() { }\n};",
      options: ['always'],
      errors: [
        { messageId: 'always' },
        { messageId: 'always' },
      ],
    },
    {
      code: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\nbaz() { }\nqux() { }\n}",
      output: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\n\nqux() { }\n}",
      options: ['always', { exceptAfterOverload: true }],
      errors: [
        { messageId: 'always' },
        { messageId: 'always' },
      ],
    },
    {
      code: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\nbaz() { }\nqux() { }\n}",
      output: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\nqux() { }\n}",
      options: ['always', { exceptAfterOverload: true, exceptAfterSingleLine: true }],
      errors: [
        { messageId: 'always' },
      ],
    },
    {
      code: "class foo {\nbar(a: string): void;\nbar(a: string, b:string): void;\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\nqux() { }\n}",
      output: "class foo {\nbar(a: string): void;\n\nbar(a: string, b:string): void;\n\nbar(a: string, b:string) {\n\n}\n\nbaz() { }\n\nqux() { }\n}",
      options: ['always', { exceptAfterOverload: false, exceptAfterSingleLine: false }],
      errors: [
        { messageId: 'always' },
        { messageId: 'always' },
        { messageId: 'always' },
      ],
    },
    {
      code: "class foo{\nbar(a: string):void;\n\nbar(a: string, b:string):void;\n\nbar(a: string, b:string){\n\n}\n\nbaz() { }\n\nqux() { }\n};",
      output: "class foo{\nbar(a: string):void;\nbar(a: string, b:string):void;\nbar(a: string, b:string){\n\n}\nbaz() { }\nqux() { }\n};",
      options: ['never', { exceptAfterOverload: true, exceptAfterSingleLine: true }],
      errors: [
        { messageId: 'never' },
        { messageId: 'never' },
        { messageId: 'never' },
        { messageId: 'never' },
      ],
    },
  ],
});

/**
 * ====================== lines-between-class-members — KNOWN GAPS ======================
 *
 * One invalid case is isolated here. It is NOT a parser-level incompatibility but a
 * fix-application difference: ESLint's RuleTester `output` captures a SINGLE fix
 * pass, while rslint's `--fix` (the engine this RuleTester drives) rewrites to a
 * stable fixpoint (multi-pass). The rule's `'always'` fixer does
 * `fixer.insertTextAfter(curLineLastToken, '\n')` — it adds exactly ONE newline per
 * pass (lines-between-class-members.ts:267). The other green invalid cases reach
 * their final `output` in a single pass, so only the fixture below diverges.
 *
 * ---- invalid (upstream expects 1 `always` diagnostic + a single-pass fix) ----
 *
 *   { code: 'class C { foo;bar }', output: 'class C { foo;\nbar }', options: ['always'], errors: [{ messageId: 'always' }] }
 *
 *   upstream (1 pass):  'class C { foo;bar }'  ->  'class C { foo;\nbar }'
 *   rslint (fixpoint):  'class C { foo;bar }'  ->  'class C { foo;\n\nbar }'
 *
 *   The members start on the SAME line with no newline between them, so upstream's
 *   first pass only splits them onto two lines (`foo;\nbar`); a notional second pass
 *   would then insert the required blank line. rslint runs that second pass and emits
 *   `foo;\n\nbar`. The diagnostic itself matches (1x `always`); only the captured
 *   fix output differs. This is the documented multi-pass-vs-single-pass gap, not a
 *   rule-logic bug, and the expected upstream behaviour is preserved above for the
 *   record.
 *
 * All other upstream fixtures for this rule are valid TypeScript under rslint's ts-go
 * parser (plain class bodies, fields, methods, overload signatures — no octal
 * escapes, no `assert` import attributes, no sloppy-mode-only syntax, no JSX), and
 * every invalid case pins both `errors` and `output`. There are NO output-only cases.
 * ====================================================================================
 */
