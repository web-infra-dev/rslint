import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-dangle', null as never, {
  valid: [
    // default (never)
    { code: "var foo = { bar: 'baz' }" },
    { code: "var foo = {\nbar: 'baz'\n}" },
    { code: "var foo = [ 'baz' ]" },
    { code: "var foo = [\n'baz'\n]" },
    { code: '[,,]' },
    { code: '[\n,\n,\n]' },
    { code: '[,]' },
    { code: '[\n,\n]' },
    { code: '[]' },
    { code: '[\n]' },
    {
      code: 'var foo = [\n      (bar ? baz : qux),\n    ];',
      options: ['always-multiline'],
    },

    // never
    { code: "var foo = { bar: 'baz' }", options: ['never'] },
    { code: "var foo = {\nbar: 'baz'\n}", options: ['never'] },
    { code: "var foo = [ 'baz' ]", options: ['never'] },
    { code: 'var { a, b } = foo;', options: ['never'] },
    { code: 'var [ a, b ] = foo;', options: ['never'] },
    { code: 'var { a,\n b, \n} = foo;', options: ['only-multiline'] },
    { code: 'var [ a,\n b, \n] = foo;', options: ['only-multiline'] },

    // always
    { code: '[(1),]', options: ['always'] },
    { code: 'var x = { foo: (1),};', options: ['always'] },
    { code: "var foo = { bar: 'baz', }", options: ['always'] },
    { code: "var foo = {\nbar: 'baz',\n}", options: ['always'] },
    { code: "var foo = {\nbar: 'baz'\n,}", options: ['always'] },
    { code: "var foo = [ 'baz', ]", options: ['always'] },
    { code: "var foo = [\n'baz',\n]", options: ['always'] },
    { code: "var foo = [\n'baz'\n,]", options: ['always'] },

    // always-multiline / only-multiline
    { code: "var foo = { bar: 'baz' }", options: ['always-multiline'] },
    { code: "var foo = { bar: 'baz' }", options: ['only-multiline'] },
    { code: "var foo = {\nbar: 'baz',\n}", options: ['always-multiline'] },
    { code: "var foo = {\nbar: 'baz',\n}", options: ['only-multiline'] },
    { code: "var foo = [ 'baz' ]", options: ['always-multiline'] },
    { code: "var foo = [ 'baz' ]", options: ['only-multiline'] },
    { code: "var foo = [\n'baz',\n]", options: ['always-multiline'] },
    { code: "var foo = [\n'baz',\n]", options: ['only-multiline'] },

    // rest binding / spread
    { code: 'var [a, ...rest] = [];', options: ['always'] },
    {
      code: 'var [\n    a,\n    ...rest\n] = [];',
      options: ['always-multiline'],
    },
    { code: '[a, ...rest] = [];', options: ['always'] },
    { code: 'for ([a, ...rest] of []);', options: ['always'] },
    { code: 'var a = [b, ...spread,];', options: ['always'] },
    { code: 'var {foo, ...bar} = baz', options: ['always'] },

    // import / export
    { code: "import {foo,} from 'foo';", options: ['always'] },
    { code: "import foo from 'foo';", options: ['always'] },
    { code: "import foo, {abc,} from 'foo';", options: ['always'] },
    { code: "import * as foo from 'foo';", options: ['always'] },
    { code: "export {foo,} from 'foo';", options: ['always'] },
    { code: "import {foo} from 'foo';", options: ['never'] },
    { code: "import foo from 'foo';", options: ['never'] },
    { code: "import foo, {abc} from 'foo';", options: ['never'] },
    { code: "import * as foo from 'foo';", options: ['never'] },
    { code: "export {foo} from 'foo';", options: ['never'] },

    // functions (object form)
    {
      code: 'function foo(a) {} ',
      options: [{ functions: 'never' }],
    },
    {
      code: 'function foo(a,) {}',
      options: [{ functions: 'always' }],
    },
    {
      code: 'foo(a,)',
      options: [{ functions: 'always' }],
    },
    {
      code: 'function bar(a, ...b) {}',
      options: [{ functions: 'always' }],
    },

    // dynamic import
    { code: 'import(source)' },
    { code: 'import(source, )', options: ['always'] },
    { code: 'import(source)', options: ['never'] },
    { code: 'import(source, options)', options: ['never'] },

    // import attributes
    { code: 'import foo from "foo" with {type: "json"}' },
    { code: 'import foo from "foo" with {type: "json",}', options: ['always'] },
    { code: 'export {foo} from "foo" with {type: "json"}' },

    // TS default
    { code: 'enum Foo {}' },
    { code: 'enum Foo {Bar}' },
    { code: 'function Foo<T>() {}' },
    { code: 'type Foo = []' },

    // TS never / always
    { code: 'enum Foo {Bar}', options: ['never'] },
    { code: 'enum Foo {Bar,}', options: ['always'] },
    { code: 'function Foo<T>() {}', options: ['never'] },
    { code: 'function Foo<T,>() {}', options: ['always'] },
    { code: 'type Foo = [string]', options: ['never'] },
    { code: 'type Foo = [string,]', options: ['always'] },
  ],
  invalid: [
    // default (never)
    {
      code: "var foo = { bar: 'baz', }",
      output: "var foo = { bar: 'baz' }",
      errors: [{ messageId: 'unexpected', line: 1, column: 23 }],
    },
    {
      code: "var foo = {\nbar: 'baz',\n}",
      output: "var foo = {\nbar: 'baz'\n}",
      errors: [{ messageId: 'unexpected', line: 2, column: 11 }],
    },
    {
      code: "foo({ bar: 'baz', qux: 'quux', });",
      output: "foo({ bar: 'baz', qux: 'quux' });",
      errors: [{ messageId: 'unexpected', line: 1, column: 30 }],
    },
    {
      code: "var foo = [ 'baz', ]",
      output: "var foo = [ 'baz' ]",
      errors: [{ messageId: 'unexpected', line: 1, column: 18 }],
    },

    // never (explicit)
    {
      code: "var foo = { bar: 'baz', }",
      output: "var foo = { bar: 'baz' }",
      options: ['never'],
      errors: [{ messageId: 'unexpected', line: 1, column: 23 }],
    },

    // always: missing
    {
      code: "var foo = { bar: 'baz' }",
      output: "var foo = { bar: 'baz', }",
      options: ['always'],
      errors: [{ messageId: 'missing', line: 1, column: 23 }],
    },
    {
      code: "var foo = {\nbar: 'baz'\n}",
      output: "var foo = {\nbar: 'baz',\n}",
      options: ['always'],
      errors: [{ messageId: 'missing', line: 2, column: 11 }],
    },
    {
      code: "var foo = [ 'baz' ]",
      output: "var foo = [ 'baz', ]",
      options: ['always'],
      errors: [{ messageId: 'missing', line: 1, column: 18 }],
    },

    // always-multiline
    {
      code: "var foo = {\nbar: 'baz'\n}",
      output: "var foo = {\nbar: 'baz',\n}",
      options: ['always-multiline'],
      errors: [{ messageId: 'missing', line: 2, column: 11 }],
    },
    {
      code: "var foo = { bar: 'baz', }",
      output: "var foo = { bar: 'baz' }",
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected', line: 1, column: 23 }],
    },

    // destructuring
    {
      code: 'var { a, b, } = foo;',
      output: 'var { a, b } = foo;',
      options: ['never'],
      errors: [{ messageId: 'unexpected', line: 1, column: 11 }],
    },
    {
      code: 'var [ a, b, ] = foo;',
      output: 'var [ a, b ] = foo;',
      options: ['never'],
      errors: [{ messageId: 'unexpected', line: 1, column: 11 }],
    },

    // import / export
    {
      code: "import {foo} from 'foo';",
      output: "import {foo,} from 'foo';",
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: "import {foo,} from 'foo';",
      output: "import {foo} from 'foo';",
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "export {foo,} from 'foo';",
      output: "export {foo} from 'foo';",
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },

    // functions (object form)
    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(a) {}',
      output: 'function foo(a,) {}',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },

    // functions (string form)
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a)',
      output: 'foo(a,)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // default with trailing-comma call
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      errors: [{ messageId: 'unexpected', line: 1, column: 6 }],
    },

    // dynamic import
    {
      code: 'import(source,)',
      output: 'import(source)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import(source)',
      output: 'import(source,)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // import attributes
    {
      code: 'import foo from "foo" with {type: "json",}',
      output: 'import foo from "foo" with {type: "json"}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      output: 'import foo from "foo" with {type: "json",}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // TS
    {
      code: 'enum Foo {Bar,}',
      output: 'enum Foo {Bar}',
      errors: [{ messageId: 'unexpected' }],
    },
    // SKIP: rule-tester's default filename is `.tsx`; the rule applies the
    // upstream TSX carve-out (`<T,>` is the JSX-disambiguation form) when the
    // type-parameter list has a single param, so this case reports 0 errors
    // here. Covered by Go upstream/extras tests where filename is `.ts`.
    // { code: 'function Foo<T,>() {}', ... }
    {
      code: 'type Foo = [string,]',
      output: 'type Foo = [string]',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'enum Foo {Bar}',
      output: 'enum Foo {Bar,}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'type Foo<T> = Bar<T,>',
      output: 'type Foo<T> = Bar<T>',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
