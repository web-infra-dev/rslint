import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-expressions', {
  valid: [
    'function f(){}',
    'a = b',
    'new a',
    '{}',
    'f(); g()',
    'i++',
    'a()',
    { code: 'a && a()', options: [{ allowShortCircuit: true }] as any },
    {
      code: 'a() || (b = c)',
      options: [{ allowShortCircuit: true }] as any,
    },
    { code: 'a ? b() : c()', options: [{ allowTernary: true }] as any },
    {
      code: 'a ? b() || (c = d) : e()',
      options: [{ allowShortCircuit: true, allowTernary: true }] as any,
    },
    'delete foo.bar',
    'void new C',
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
      options: [{ allowShortCircuit: true }] as any,
    },
    {
      code: 'async function foo() { foo ? await bar : await baz; }',
      options: [{ allowTernary: true }] as any,
    },
    {
      code: 'tag`tagged template literal`',
      options: [{ allowTaggedTemplates: true }] as any,
    },
    {
      code: 'shouldNotBeAffectedByAllowTemplateTagsOption()',
      options: [{ allowTaggedTemplates: true }] as any,
    },
    'import("foo")',
    'func?.("foo")',
    'obj?.foo("bar")',

    // JSX upstream cases live in Go tests because this JS harness uses virtual.ts.

    { code: '"use strict";', options: [{ ignoreDirectives: true }] as any },
    {
      code: '"directive one"; "directive two"; f();',
      options: [{ ignoreDirectives: true }] as any,
    },
    {
      code: 'function foo() {"use strict"; return true; }',
      options: [{ ignoreDirectives: true }] as any,
    },
    {
      code: 'function foo() {"directive one"; "directive two"; f(); }',
      options: [{ ignoreDirectives: true }] as any,
    },

    'test.age?.toLocaleString();',
    'let a = (a?.b).c;',
    "let b = a?.['b'];",
    'let c = one[2]?.[3][4];',
    'one[2]?.[3][4]?.();',
    "a?.['b']?.c();",
    "module Foo {\n  'use strict';\n}",
    "namespace Foo {\n  'use strict';\n\n  export class Foo {}\n  export class Bar {}\n}",
    "function foo() {\n  'use strict';\n\n  return null;\n}",
    "import('./foo');",
    "import('./foo').then(() => {});",
    'class Foo<T> {}\nnew Foo<string>();',
    {
      code: 'foo && foo?.();',
      options: [{ allowShortCircuit: true }] as any,
    },
    {
      code: "foo && import('./foo');",
      options: [{ allowShortCircuit: true }] as any,
    },
    {
      code: "foo ? import('./foo') : import('./bar');",
      options: [{ allowTernary: true }] as any,
    },
    {
      code: 'foo && foo()!;',
      options: [{ allowShortCircuit: true }] as any,
    },
    {
      code: 'declare const foo: Function | undefined;\n<any>(foo && foo()!)',
      options: [{ allowShortCircuit: true }] as any,
    },
    {
      code: '(Foo && Foo())<string, number>;',
      options: [{ allowShortCircuit: true }] as any,
    },
  ],
  invalid: [
    { code: '0', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'a', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'f(), 0', errors: [{ messageId: 'unusedExpression' }] },
    { code: '{0}', errors: [{ messageId: 'unusedExpression' }] },
    { code: '[]', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'a && b();', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'a() || false', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'a || (b = c)', errors: [{ messageId: 'unusedExpression' }] },
    {
      code: 'a ? b() || (c = d) : e',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '`untagged template literal`',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'tag`tagged template literal`',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a && b()',
      options: [{ allowTernary: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a ? b() : c()',
      options: [{ allowShortCircuit: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a || b',
      options: [{ allowShortCircuit: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a() && b',
      options: [{ allowShortCircuit: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a ? b : 0',
      options: [{ allowTernary: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'a ? b : c()',
      options: [{ allowTernary: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    { code: 'foo.bar;', errors: [{ messageId: 'unusedExpression' }] },
    { code: '!a', errors: [{ messageId: 'unusedExpression' }] },
    { code: '+a', errors: [{ messageId: 'unusedExpression' }] },
    {
      code: '"directive one"; f(); "directive two";',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'function foo() {"directive one"; f(); "directive two"; }',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'if (0) { "not a directive"; f(); }',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'function foo() { var foo = true; "use strict"; }',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'var foo = () => { var foo = true; "use strict"; }',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '`untagged template literal`',
      options: [{ allowTaggedTemplates: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '`untagged template literal`',
      options: [{ allowTaggedTemplates: false }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'tag`tagged template literal`',
      options: [{ allowTaggedTemplates: false }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    { code: 'obj?.foo', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'obj?.foo.bar', errors: [{ messageId: 'unusedExpression' }] },
    { code: 'obj?.foo().bar', errors: [{ messageId: 'unusedExpression' }] },

    // JSX upstream invalid cases live in Go tests.

    {
      code: "class C { static { 'use strict'; } }",
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: "class C { static {\n'foo'\n'bar'\n} }",
      errors: [
        { messageId: 'unusedExpression' },
        { messageId: 'unusedExpression' },
      ],
    },
    {
      code: 'foo;',
      options: [{ ignoreDirectives: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    // SKIP: rslint's JS harness does not expose ESLint's ecmaVersion: 3 mode.
    { code: '"use strict";', skip: true, errors: [] },
    // SKIP: rslint's JS harness does not expose ESLint's ecmaVersion: 3 mode.
    { code: '"directive one"; "directive two"; f();', skip: true, errors: [] },
    // SKIP: rslint's JS harness does not expose ESLint's ecmaVersion: 3 mode.
    {
      code: 'function foo() {"use strict"; return true; }',
      skip: true,
      errors: [],
    },
    // SKIP: rslint's JS harness does not expose ESLint's ecmaVersion: 3 mode.
    {
      code: 'function foo() {"directive one"; "directive two"; f(); }',
      skip: true,
      errors: [],
    },

    { code: '\n  if (0) 0;\n', errors: [{ messageId: 'unusedExpression' }] },
    { code: '\n  f(0), {};\n', errors: [{ messageId: 'unusedExpression' }] },
    { code: '\n  a, b();\n', errors: [{ messageId: 'unusedExpression' }] },
    {
      code: '\n  a() &&\n\tfunction namedFunctionInExpressionContext() {\n\t  f();\n\t};\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    { code: '\n  a?.b;\n', errors: [{ messageId: 'unusedExpression' }] },
    { code: '\n  (a?.b).c;\n', errors: [{ messageId: 'unusedExpression' }] },
    { code: "\n  a?.['b'];\n", errors: [{ messageId: 'unusedExpression' }] },
    {
      code: "\n  (a?.['b']).c;\n",
      errors: [{ messageId: 'unusedExpression' }],
    },
    { code: '\n  a?.b()?.c;\n', errors: [{ messageId: 'unusedExpression' }] },
    { code: '\n  (a?.b()).c;\n', errors: [{ messageId: 'unusedExpression' }] },
    {
      code: '\n  one[2]?.[3][4];\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  one.two?.three.four;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: "module Foo {\n  const foo = true;\n  'use strict';\n}",
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: "namespace Foo {\n  export class Foo {}\n  export class Bar {}\n\n  'use strict';\n}",
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: "function foo() {\n  const foo = true;\n\n  'use strict';\n}",
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'foo && foo?.bar;',
      options: [{ allowShortCircuit: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'foo ? foo?.bar : bar.baz;',
      options: [{ allowTernary: true }] as any,
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  class Foo<T> {}\n  Foo<string>;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: 'Map<string, string>;',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  declare const foo: number | undefined;\n  foo;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  declare const foo: number | undefined;\n  foo as any;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  declare const foo: number | undefined;\n  <any>foo;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
    {
      code: '\n  declare const foo: number | undefined;\n  foo!;\n',
      errors: [{ messageId: 'unusedExpression' }],
    },
  ],
});
