// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-useless-spread.js
// Includes every upstream case and documentation example, with documented safety differences.
// Go tests also assert exact ranges and edits.
import { RuleTester } from '../rule-tester';

const tester = new RuleTester();

// SpreadInList
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: 'const array = [[]]',
      filename: 'src/virtual.js',
    },
    {
      code: 'const array = [{}]',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = ({...[]})',
      filename: 'src/virtual.js',
    },
    {
      code: 'foo([])',
      filename: 'src/virtual.js',
    },
    {
      code: 'foo({})',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Foo([])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Foo({})',
      filename: 'src/virtual.js',
    },
    {
      code: 'const array = [...a]',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {...a}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const [first, ...rest] = []',
      filename: 'src/virtual.js',
    },
    {
      code: 'const {foo, ...rest} = {}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function a(foo, ...rest) {}',
      filename: 'src/virtual.js',
    },
    {
      code: '({\n\tget a() {},\n\tset a(v) {},\n\t...{\n\t\tget a() {}\n\t}\n})',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'const array = [...[a]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a]',
    },
    {
      code: 'const object = {...{a}}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = {a}',
    },
    {
      code: 'foo(...[a])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a)',
    },
    {
      code: 'new Foo(...[a])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a)',
    },
    {
      code: 'const array = [...[a,]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a]',
    },
    {
      code: 'const object = {...{a,}}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = {a}',
    },
    {
      code: 'foo(...[a,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a)',
    },
    {
      code: 'new Foo(...[a,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a)',
    },
    {
      code: 'const array = [...[a,],]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a,]',
    },
    {
      code: 'const object = {...{a,},}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = {a,}',
    },
    {
      code: 'foo(...[a,],)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a,)',
    },
    {
      code: 'new Foo(...[a,],)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a,)',
    },
    {
      code: 'const array = [...(( [a] ))]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [ a ]',
    },
    {
      code: 'const object = {...(( {a} ))}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = { a }',
    },
    {
      code: 'foo(...(( [a] )))',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo( a )',
    },
    {
      code: 'new Foo(...(( [a] )))',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo( a )',
    },
    {
      code: 'const array = [...[]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = []',
    },
    {
      code: 'const object = {...{}}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = {}',
    },
    {
      code: 'foo(...[])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo()',
    },
    {
      code: 'new Foo(...[])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo()',
    },
    {
      code: 'const array = [...[,]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = []',
    },
    {
      code: 'foo(...[,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(undefined)',
    },
    {
      code: 'new Foo(...[,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(undefined)',
    },
    {
      code: 'const array = [...[,,]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [,]',
    },
    {
      code: 'foo(...[,,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(undefined,undefined)',
    },
    {
      code: 'new Foo(...[,,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(undefined,undefined)',
    },
    {
      code: 'const array = [...[a, , b,]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a, , b]',
    },
    {
      code: 'foo(...[a, , b,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a, undefined, b)',
    },
    {
      code: 'new Foo(...[a, , b,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a, undefined, b)',
    },
    {
      code: 'const array = [...[a, , b,],]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a, , b,]',
    },
    {
      code: 'foo(...[a, , b,],)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a, undefined, b,)',
    },
    {
      code: 'new Foo(...[a, , b,],)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a, undefined, b,)',
    },
    {
      code: 'foo(...[,, ,(( a )), ,,(0, b), ,,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output:
        'foo(undefined,undefined, undefined,(( a )), undefined,undefined,(0, b), undefined,undefined)',
    },
    {
      code: 'const array = [a, ...[a, b]]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a, a, b]',
    },
    {
      code: 'const object = {a, ...{a, b}}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'const object = {a, a, b}',
    },
    {
      code: 'foo(a, ...[a, b])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
      output: 'foo(a, a, b)',
    },
    {
      code: 'new Foo(a, ...[a, b])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
      output: 'new Foo(a, a, b)',
    },
    {
      code: 'const array = [...[a, b], b,]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const array = [a, b, b,]',
    },
    {
      code: 'const object = {...{a, b}, b,}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'const object = {a, b, b,}',
    },
    {
      code: 'foo(...[a, b], b,)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      output: 'foo(a, b, b,)',
    },
    {
      code: 'new Foo(...[a, b], b,)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Foo(a, b, b,)',
    },
    {
      code: 'const array = [a, ...[a, b], b,]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a, a, b, b,]',
    },
    {
      code: 'const object = {a, ...{a, b}, b,}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'const object = {a, a, b, b,}',
    },
    {
      code: 'foo(a, ...[a, b], b,)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
      output: 'foo(a, a, b, b,)',
    },
    {
      code: 'new Foo(a, ...[a, b], b,)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
      output: 'new Foo(a, a, b, b,)',
    },
    {
      code: 'const array = [a, ...[], b]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a,  b]',
    },
    {
      code: 'const array = [a, ...(( [] )),]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a,   ]',
    },
    {
      code: 'const array = [a, ...(( [] ))]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a,   ]',
    },
    {
      code: 'const array = [a, ...[b], c]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'const array = [a, b, c]',
    },
    {
      code: 'const object = {a, ...(({})), b,}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'const object = {a,  b,}',
    },
    {
      code: '({a:1, ...{a: 2}})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
      output: '({a:1, a: 2})',
    },
    {
      code: '({...{a:1}, ...{a: 2}})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: '({a:1, a: 2})',
    },
    {
      code: '({[a]:1, ...{[a]: 2}})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
      output: '({[a]:1, [a]: 2})',
    },
    {
      code: 'const object = {\n\ta: 1,\n\n\t...{\n\t\ttestKeys() {\n\t\t\tconsole.assert(Object.keys(this).length === 2)\n\t\t}\n\t}\n}\nobject.testKeys();',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 4,
          column: 2,
          endLine: 4,
          endColumn: 5,
          suggestions: [],
        },
      ],
      output:
        'const object = {\n\ta: 1,\n\n\t\n\t\ttestKeys() {\n\t\t\tconsole.assert(Object.keys(this).length === 2)\n\t\t}\n\t\n}\nobject.testKeys();',
    },
    {
      code: 'new Foo(\n\tfoo(\n\t\ta,\n\t\t...[a, b],\n\t\tb,\n\t),\n\t...[\n\t\ta,\n\t\t...[\n\t\t\ta,\n\t\t\tb,\n\t\t],\n\t\tb,\n\t],\n\t{\n\t\ta: [...[a, b]],\n\t\t...{\n\t\t\ta,\n\t\t\tb,\n\t\t},\n\t}\n)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 7,
          column: 2,
          endLine: 7,
          endColumn: 5,
          suggestions: [],
        },
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 9,
          column: 3,
          endLine: 9,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 16,
          column: 7,
          endLine: 16,
          endColumn: 10,
          suggestions: [],
        },
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 17,
          column: 3,
          endLine: 17,
          endColumn: 6,
          suggestions: [],
        },
      ],
      output:
        'new Foo(\n\tfoo(\n\t\ta,\n\t\ta, b,\n\t\tb,\n\t),\n\t\n\t\ta,\n\t\t...[\n\t\t\ta,\n\t\t\tb,\n\t\t],\n\t\tb\n\t,\n\t{\n\t\ta: [a, b],\n\t\t\n\t\t\ta,\n\t\t\tb\n\t\t,\n\t}\n)',
    },
    {
      code: 'const baz = [2];\ncall(foo, ...[bar, ...baz]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 14,
          suggestions: [],
        },
      ],
      output: 'const baz = [2];\ncall(foo, bar, ...baz);',
    },
    {
      code: 'Promise.all(...[...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'Promise.all(...iterable)',
    },
    {
      code: 'new Map(...[...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'new Set(...[iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
      output: 'new Set(iterable)',
    },
    {
      code: 'Object.assign(target, {...{a}})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: 'Object.assign(target, {a})',
    },
  ],
});

// ObjectAssign
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: 'Object.assign(target, source)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(target, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(target, {foo})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(target, {foo, ...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(target, {...source, foo})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign?.(target, {...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object?.assign(target, {...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object[assign](target, {...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'NotObject.assign(target, {...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign({...source})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign({...target}, source)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(...args, {...source})',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'Object.assign(target, {...source})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, source)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {...a}, {...b})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, a, {...b})',
            },
          ],
        },
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 35,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, {...a}, b)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, first, {...second})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, first, second)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(...args, target, {...source})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(...args, target, source)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {...first, ...second}, third)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, first, second, third)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {...source,}, third)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, source, third)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {...(( source ))})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, (( source )))',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {...(foo, bar)})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, (foo, bar))',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'Object.assign(target, {/* keep */ ...source})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
      output: null,
    },
  ],
});

// CollectionConstructor
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: 'new Set(iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new NotSet(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new namespace.Set(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Set(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Uint8Array(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Set(...iterable, extraArgument)',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'new Set(...iterable)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new Set(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'new Map(...iterable)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new Map(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'new WeakSet(...iterable)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new WeakSet(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'new WeakMap(...iterable)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new WeakMap(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'new Set(...getNames())',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new Set(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: null,
    },
  ],
});

// IterableConversion
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: 'new NotMatchedConstructor([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new foo.Map([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Map([...iterable], extraArgument)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Map()',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Map([,...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Map([...iterable, extraElement])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Map({...iterable})',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Uint8Array(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Uint8Array(before, [...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Uint8Array([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Float64Array([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Int32Array([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new BigInt64Array([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'new Uint8Array([...iterable], 0)',
      filename: 'src/virtual.js',
    },
    {
      code: 'const typed = new BigUint64Array([...iterable], byteOffset, length)',
      filename: 'src/virtual.js',
    },
    {
      code: 'const typed = new BigUint64Array([...iterable], ...args)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.notMatchedMethod([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'NotPromise.all([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'foo.Promise.all([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all?.([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise?.all([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise[all]([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all()',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all([...iterable], extraArgument)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all([,...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all([...iterable, extraElement])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Promise.all({...iterable})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.notFromEntries([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'NotObject.fromEntries([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.fromEntries?.([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object?.fromEntries([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object[fromEntries]([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.fromEntries()',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.fromEntries([...iterable], extraArgument)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.fromEntries(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.fromEntries({...iterable})',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.notFrom([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'NotTypedArray.from([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.from?.([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array?.from([...iterable])',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.from([...iterable], extraArgument)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.from(...iterable)',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.from({...iterable})',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const [...iterable] of foo);',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of bar) [...iterable];',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of [,...iterable]);',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of [...iterable, extraElement]);',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of {...iterable});',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo in [...iterable]);',
      filename: 'src/virtual.js',
    },
    {
      code: 'function * fn() {yield [...iterable];}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function * fn() {yield* [...iterable, extraElement];}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function * fn() {yield* {...iterable};}',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'const map = new Map([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Map(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: 'const map = new Map(iterable)',
    },
    {
      code: 'const weakMap = new WeakMap([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new WeakMap(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'const weakMap = new WeakMap(iterable)',
    },
    {
      code: 'const set = new Set([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Set(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: 'const set = new Set(iterable)',
    },
    {
      code: 'const weakSet = new WeakSet([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new WeakSet(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'const weakSet = new WeakSet(iterable)',
    },
    {
      code: 'const promise = Promise.all([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Promise.all(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'const promise = Promise.all(iterable)',
    },
    {
      code: 'const promise = Promise.allSettled([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Promise.allSettled(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
      output: 'const promise = Promise.allSettled(iterable)',
    },
    {
      code: 'const promise = Promise.any([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Promise.any(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'const promise = Promise.any(iterable)',
    },
    {
      code: 'const promise = Promise.race([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Promise.race(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
      output: 'const promise = Promise.race(iterable)',
    },
    {
      code: 'const array = Array.from([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: 'const array = Array.from(iterable)',
    },
    {
      code: 'const typed = BigUint64Array.from([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`BigUint64Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
      output: 'const typed = BigUint64Array.from(iterable)',
    },
    {
      code: 'const object = Object.fromEntries([...iterable])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Object.fromEntries(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
      output: 'const object = Object.fromEntries(iterable)',
    },
    {
      code: 'for (const foo of [...iterable]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of iterable);',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'async () => {for await (const foo of [...iterable]);}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'async () => {for await (const foo of iterable);}',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'const map = new Map([...iterable,])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Map(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'const map = new Map(iterable)',
    },
    {
      code: 'for (const foo of [...iterable,]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of iterable);',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'const map = new Map([...iterable,],)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Map(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'const map = new Map(iterable,)',
    },
    {
      code: 'const map = new Map([...(( iterable ))])',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Map(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 40,
          suggestions: [],
        },
      ],
      output: 'const map = new Map((( iterable )))',
    },
    {
      code: 'for (const foo of [...(( iterable ))]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of (( iterable )));',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'const map = new Map((( [...(( iterable ))] )))',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Map(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
      output: 'const map = new Map((( (( iterable )) )))',
    },
    {
      code: 'for (const foo of (( [...(( iterable ))] )));',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 41,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of (( (( iterable )) )));',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'for (const foo of[...iterable]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of iterable);',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'function * fn() {\n\tyield * [...iterable];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-yield-star',
          message:
            '`yield*` can delegate directly when materializing the iterable is unnecessary.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'function * fn() {\n\tyield * iterable;\n}',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'function * fn() {\n\tyield * [...iterable,];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-yield-star',
          message:
            '`yield*` can delegate directly when materializing the iterable is unnecessary.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'function * fn() {\n\tyield * iterable;\n}',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'function * fn() {\n\tyield * (( [...iterable] ));\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-yield-star',
          message:
            '`yield*` can delegate directly when materializing the iterable is unnecessary.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'function * fn() {\n\tyield * (( iterable ));\n}',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'function * fn() {\n\tyield * (( [...(( iterable ))] ));\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-yield-star',
          message:
            '`yield*` can delegate directly when materializing the iterable is unnecessary.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'function * fn() {\n\tyield * (( (( iterable )) ));\n}',
            },
          ],
        },
      ],
      output: null,
    },
  ],
});

// ArrayClone
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: '[...not.array]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...not.array()]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...array.unknown()]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...Object.notReturningArray(foo)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...NotObject.keys(foo)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...Int8Array.from(foo)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...Int8Array.of()]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...new Int8Array(3)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...Promise.all(foo)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...Promise.allSettled(foo)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...await Promise.all(foo, extraArgument)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...foo.filter(bar)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...foo.flatMap(bar)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...foo.map(bar)]',
      filename: 'src/virtual.js',
    },
    {
      code: 'function foo(array: number[]) { return [...array]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(array: number[]) { return [...array]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'const data = new Map([["a", 1], ["b", 2], ["c", 3]]); const foo = [...data.values().map(value => value * 2)];',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: string) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: {slice(start: number): string}) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: {split(separator: string): string}) { return [...value.split("|")]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: {concat(value: string): string}) { return [...value.concat("bar")]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: {map(): number[]}) { return [...value.map()]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: {slice(start: number): number[]}) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: Int32Array) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'class ArraySubclass extends Array<number> {} function foo(value: ArraySubclass) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: string) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function foo(value: Int32Array) { return [...value.slice(1)]; }',
      filename: 'src/virtual.ts',
    },
    {
      code: '[...Iterator.concat(bar)]',
      filename: 'src/virtual.js',
    },
    {
      code: '[...foo.copyWithin(-2)]',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: '[...foo.concat(bar)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'foo.concat(bar)',
    },
    {
      code: '[...foo.flat()]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'foo.flat()',
    },
    {
      code: '[...foo.slice(1)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: '[...foo.splice(1)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'foo.splice(1)',
    },
    {
      code: '[...foo.toReversed()]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'foo.toReversed()',
    },
    {
      code: '[...foo.toSorted()]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'foo.toSorted()',
    },
    {
      code: '[...foo.toSpliced(0, 1)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
      output: 'foo.toSpliced(0, 1)',
    },
    {
      code: '[...foo.with(0, bar)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'foo.with(0, bar)',
    },
    {
      code: '[...foo.split("|")]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'foo.split("|")',
    },
    {
      code: '[...Object.keys(foo)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output: 'Object.keys(foo)',
    },
    {
      code: '[...Object.values(foo)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'Object.values(foo)',
    },
    {
      code: '[...Array.from(foo)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'Array.from(foo)',
    },
    {
      code: '[...Array.of()]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: 'Array.of()',
    },
    {
      code: '[...new Array(3)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: '[...await Promise.all(foo)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: '(await Promise.all(foo))',
    },
    {
      code: '[...await Promise.allSettled(foo)]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: '(await Promise.allSettled(foo))',
    },
    {
      code: 'function foo(array: number[]) { return [...array.map(value => value * 2)]; }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 74,
          suggestions: [],
        },
      ],
      output:
        'function foo(array: number[]) { return array.map(value => value * 2); }',
    },
    {
      code: 'function foo(array: number[]) { return [...array.filter(Boolean)]; }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 66,
          suggestions: [],
        },
      ],
      output: 'function foo(array: number[]) { return array.filter(Boolean); }',
    },
    {
      code: 'function foo(array: number[]) { return [...array.flatMap(value => value)]; }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 74,
          suggestions: [],
        },
      ],
      output:
        'function foo(array: number[]) { return array.flatMap(value => value); }',
    },
    {
      code: 'function foo(bar) {\n\treturn[...Object.keys(bar)];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 29,
          suggestions: [],
        },
      ],
      output: 'function foo(bar) {\n\treturn Object.keys(bar);\n}',
    },
    {
      code: 'function foo(bar) {\n\treturn[\n\t\t...Object.keys(bar)\n\t];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 2,
          column: 8,
          endLine: 4,
          endColumn: 3,
          suggestions: [],
        },
      ],
      output: 'function foo(bar) {\n\treturn (\n\t\tObject.keys(bar)\n\t);\n}',
    },
    {
      code: 'function foo(bar) {\n\treturn[\n\t\t...(\n\t\t\tObject.keys(bar)\n\t\t)\n\t];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 2,
          column: 8,
          endLine: 6,
          endColumn: 3,
          suggestions: [],
        },
      ],
      output:
        'function foo(bar) {\n\treturn (\n\t\t(\n\t\t\tObject.keys(bar)\n\t\t)\n\t);\n}',
    },
    {
      code: 'function foo(bar) {\n\treturn([\n\t\t...Object.keys(bar)\n\t]);\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 2,
          column: 9,
          endLine: 4,
          endColumn: 3,
          suggestions: [],
        },
      ],
      output: 'function foo(bar) {\n\treturn (\n\t\tObject.keys(bar)\n\t);\n}',
    },
  ],
});

// TypeScript
tester.run('no-useless-spread', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'for (const foo of[...iterable2]);',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of iterable2);',
            },
          ],
        },
      ],
      output: null,
    },
  ],
});

// Precedence
tester.run('no-useless-spread', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const n = [...await Promise.all(x)].length;',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'const n = (await Promise.all(x)).length;',
    },
    {
      code: 'const n = [...await Promise.all(x)][0];',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'const n = (await Promise.all(x))[0];',
    },
    {
      code: 'const n = [...await Promise.all(x)].map(f);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'const n = (await Promise.all(x)).map(f);',
    },
    {
      code: 'const n = [...await Promise.all(x)]?.length;',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'clone-array',
          message: 'Unnecessarily cloning an array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      output: 'const n = (await Promise.all(x))?.length;',
    },
  ],
});

// Documentation
tester.run('no-useless-spread', {} as never, {
  valid: [
    {
      code: 'const array = [firstElement, secondElement, thirdElement];',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {firstProperty, secondProperty, thirdProperty};',
      filename: 'src/virtual.js',
    },
    {
      code: 'foo(firstArgument, secondArgument, thirdArgument);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = new Foo(firstArgument, secondArgument, thirdArgument);',
      filename: 'src/virtual.js',
    },
    {
      code: 'Object.assign(target, source);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const set = new Set(iterable);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const results = await Promise.all(iterable);',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of set);',
      filename: 'src/virtual.js',
    },
    {
      code: 'function * foo() {\n\tyield * anotherGenerator();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const array = [...foo, bar];',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {...foo, bar};',
      filename: 'src/virtual.js',
    },
    {
      code: 'foo(foo, ...bar);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = new Foo(...foo, bar);',
      filename: 'src/virtual.js',
    },
    {
      code: "new Uint8Array([...'ab'])",
      filename: 'src/virtual.js',
    },
    {
      code: "new Uint8Array('ab')",
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'const array = [firstElement, ...[secondElement], thirdElement];',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in array literal is unnecessary.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
      output: 'const array = [firstElement, secondElement, thirdElement];',
    },
    {
      code: 'const object = {firstProperty, ...{secondProperty}, thirdProperty};',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an object literal in object literal is unnecessary.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
      output: 'const object = {firstProperty, secondProperty, thirdProperty};',
    },
    {
      code: 'foo(firstArgument, ...[secondArgument], thirdArgument);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'foo(firstArgument, secondArgument, thirdArgument);',
    },
    {
      code: 'const object = new Foo(firstArgument, ...[secondArgument], thirdArgument);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-list',
          message: 'Spread an array literal in arguments is unnecessary.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output:
        'const object = new Foo(firstArgument, secondArgument, thirdArgument);',
    },
    {
      code: 'Object.assign(target, {...source});',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-object-assign',
          message:
            '`Object.assign(…)` source object with only spread properties is unnecessary.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-object-assign-spread',
              output: 'Object.assign(target, source);',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'const set = new Set([...iterable]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`new Set(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: 'const set = new Set(iterable);',
    },
    {
      code: 'const set = new Set(...iterable);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'spread-in-collection-constructor',
          message:
            '`new Set(…)` accepts a single iterable argument, spreading is misleading.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'const results = await Promise.all([...iterable]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array',
          message:
            "`Promise.all(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 48,
          suggestions: [],
        },
      ],
      output: 'const results = await Promise.all(iterable);',
    },
    {
      code: 'for (const foo of [...set]);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-for-of',
          message:
            '`for…of` can iterate directly when an array snapshot is not needed.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'for (const foo of set);',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'function * foo() {\n\tyield * [...anotherGenerator()];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'iterable-to-array-in-yield-star',
          message:
            '`yield*` can delegate directly when materializing the iterable is unnecessary.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion/remove-iterable-to-array',
              output: 'function * foo() {\n\tyield * anotherGenerator();\n}',
            },
          ],
        },
      ],
      output: null,
    },
  ],
});
