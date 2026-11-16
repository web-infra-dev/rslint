import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('no-unused-expressions', {
  valid: [
    'function f(){}',
    'a = b',
    'new a',
    '{}',
    'f(); g()',
    'i++',
    'a()',
    {
      code: 'a && a()',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: 'a() || (b = c)',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: 'a ? b() : c()',
      options: [
        {
          allowTernary: true,
        },
      ],
    },
    {
      code: 'a ? b() || (c = d) : e()',
      options: [
        {
          allowShortCircuit: true,
          allowTernary: true,
        },
      ],
    },
    'delete foo.bar',
    'void new C',
    // JSX upstream cases live in Go tests because this JS harness uses virtual.ts.
    '"use strict";',
    '"directive one"; "directive two"; f();',
    'function foo() {"use strict"; return true; }',
    'var foo = () => {"use strict"; return true; }',
    'function foo() {"directive one"; "directive two"; f(); }',
    'function foo() { var foo = "use strict"; return true; }',
    'function* foo(){ yield 0; }',
    'async function foo() { await 5; }',
    'async function foo() { await foo.bar; }',
    {
      code: 'async function foo() { bar && await baz; }',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: 'async function foo() { foo ? await bar : await baz; }',
      options: [
        {
          allowTernary: true,
        },
      ],
    },
    {
      code: 'tag`tagged template literal`',
      options: [
        {
          allowTaggedTemplates: true,
        },
      ],
    },
    {
      code: 'shouldNotBeAffectedByAllowTemplateTagsOption()',
      options: [
        {
          allowTaggedTemplates: true,
        },
      ],
    },
    'import("foo")',
    'func?.("foo")',
    'obj?.foo("bar")',
    {
      code: '"use strict";',
      options: [
        {
          ignoreDirectives: true,
        },
      ],
    },
    {
      code: '"directive one"; "directive two"; f();',
      options: [
        {
          ignoreDirectives: true,
        },
      ],
    },
    {
      code: 'function foo() {"use strict"; return true; }',
      options: [
        {
          ignoreDirectives: true,
        },
      ],
    },
    {
      code: 'function foo() {"directive one"; "directive two"; f(); }',
      options: [
        {
          ignoreDirectives: true,
        },
      ],
    },
    'test.age?.toLocaleString();',
    'let a = (a?.b).c;',
    "let b = a?.['b'];",
    'let c = one[2]?.[3][4];',
    'one[2]?.[3][4]?.();',
    "a?.['b']?.c();",
    `module Foo {
  'use strict';
}`,
    `namespace Foo {
  'use strict';

  export class Foo {}
  export class Bar {}
}`,
    `function foo() {
  'use strict';

  return null;
}`,
    "import('./foo');",
    "import('./foo').then(() => {});",
    `class Foo<T> {}
new Foo<string>();`,
    {
      code: 'foo && foo?.();',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: "foo && import('./foo');",
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: "foo ? import('./foo') : import('./bar');",
      options: [
        {
          allowTernary: true,
        },
      ],
    },
    {
      code: 'foo && foo()!;',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: `declare const foo: Function | undefined;
<any>(foo && foo()!)`,
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
    {
      code: '(Foo && Foo())<string, number>;',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
    },
  ],
  invalid: [
    {
      code: '0',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'a',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'f(), 0',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '{0}',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '[]',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'a && b();',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'a() || false',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'a || (b = c)',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'a ? b() || (c = d) : e',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '`untagged template literal`',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'tag`tagged template literal`',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'a && b()',
      options: [
        {
          allowTernary: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'a ? b() : c()',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'a || b',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'a() && b',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'a ? b : 0',
      options: [
        {
          allowTernary: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'a ? b : c()',
      options: [
        {
          allowTernary: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'foo.bar;',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '!a',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '+a',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '"directive one"; f(); "directive two";',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'function foo() {"directive one"; f(); "directive two"; }',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'if (0) { "not a directive"; f(); }',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'function foo() { var foo = true; "use strict"; }',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'var foo = () => { var foo = true; "use strict"; }',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '`untagged template literal`',
      options: [
        {
          allowTaggedTemplates: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '`untagged template literal`',
      options: [
        {
          allowTaggedTemplates: false,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'tag`tagged template literal`',
      options: [
        {
          allowTaggedTemplates: false,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'obj?.foo',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'obj?.foo.bar',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'obj?.foo().bar',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    // JSX upstream invalid cases live in Go tests.
    {
      code: "class C { static { 'use strict'; } }",
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `class C { static {
'foo'
'bar'
} }`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 6,
        },
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'foo;',
      options: [
        {
          ignoreDirectives: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      skip: 'Rslint intentionally uses modern directive semantics; ES3 directive behavior is outside the supported compatibility scope.',
      code: '"use strict";',
      languageOptions: {
        ecmaVersion: 3,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      skip: 'Rslint intentionally uses modern directive semantics; ES3 directive behavior is outside the supported compatibility scope.',
      code: '"directive one"; "directive two"; f();',
      languageOptions: {
        ecmaVersion: 3,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      skip: 'Rslint intentionally uses modern directive semantics; ES3 directive behavior is outside the supported compatibility scope.',
      code: 'function foo() {"use strict"; return true; }',
      languageOptions: {
        ecmaVersion: 3,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      skip: 'Rslint intentionally uses modern directive semantics; ES3 directive behavior is outside the supported compatibility scope.',
      code: 'function foo() {"directive one"; "directive two"; f(); }',
      languageOptions: {
        ecmaVersion: 3,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 33,
        },
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: `
  if (0) 0;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
  f(0), {};
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
  a, b();
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
  a() &&
	function namedFunctionInExpressionContext() {
	  f();
	};
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 5,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
  a?.b;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: `
  (a?.b).c;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
  a?.['b'];
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
  (a?.['b']).c;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
  a?.b()?.c;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
  (a?.b()).c;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
  one[2]?.[3][4];
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
  one.two?.three.four;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 23,
        },
      ],
    },
    {
      code: `module Foo {
  const foo = true;
  'use strict';
}`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: `namespace Foo {
  export class Foo {}
  export class Bar {}

  'use strict';
}`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 16,
        },
      ],
    },
    {
      code: `function foo() {
  const foo = true;

  'use strict';
}`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'foo && foo?.bar;',
      options: [
        {
          allowShortCircuit: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'foo ? foo?.bar : bar.baz;',
      options: [
        {
          allowTernary: true,
        },
      ],
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
  class Foo<T> {}
  Foo<string>;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'Map<string, string>;',
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
  declare const foo: number | undefined;
  foo;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
  declare const foo: number | undefined;
  foo as any;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
  declare const foo: number | undefined;
  <any>foo;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
  declare const foo: number | undefined;
  foo!;
`,
      errors: [
        {
          messageId: 'unusedExpression',
          message:
            'Expected an assignment or function call and instead saw an expression.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
  ],
});
