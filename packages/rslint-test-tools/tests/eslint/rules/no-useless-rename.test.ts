import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-rename', {
  valid: [
    // ---- Destructuring (declarations) ----
    'let {foo} = obj;',
    'let {foo: bar} = obj;',
    'let {foo: bar, baz: qux} = obj;',
    'let {foo: {bar: baz}} = obj;',
    "let {'foo': bar} = obj;",
    "let {'foo': {'bar': baz}} = obj;",
    "let {['foo']: bar} = obj;",
    "let {['foo']: foo} = obj;",
    'let {[foo]: foo} = obj;',
    'function func({foo}) {}',
    'function func({foo: bar}) {}',
    '({foo}) => {};',
    '({foo: bar}) => {};',
    // rest elements cannot be renamed
    'const {...stuff} = myObject;',
    'const {foo: bar, ...stuff} = myObject;',

    // ---- Imports ----
    "import * as foo from 'foo';",
    "import foo from 'foo';",
    "import {foo} from 'foo';",
    "import {foo as bar} from 'foo';",
    "import {'foo' as bar} from 'baz';",

    // ---- Exports ----
    "export {foo} from 'foo';",
    'var foo = 0;export {foo as bar};',
    "export {foo as bar} from 'foo';",
    "var foo = 0; export {foo as 'bar'};",
    "export {'foo' as bar} from 'baz';",
    "export {'foo' as 'bar'} from 'baz';",
    "export {'foo'} from 'bar';",

    // ---- { ignoreDestructuring: true } ----
    { code: 'let {foo: foo} = obj;', options: { ignoreDestructuring: true } },
    {
      code: 'let {foo: foo, bar: bar} = obj;',
      options: { ignoreDestructuring: true },
    },

    // ---- { ignoreImport: true } ----
    {
      code: "import {foo as foo} from 'foo';",
      options: { ignoreImport: true },
    },
    {
      code: "import {foo as foo, bar as bar} from 'foo';",
      options: { ignoreImport: true },
    },

    // ---- { ignoreExport: true } ----
    {
      code: 'var foo = 0;export {foo as foo};',
      options: { ignoreExport: true },
    },
    {
      code: "export {foo as foo} from 'foo';",
      options: { ignoreExport: true },
    },
  ],
  invalid: [
    // ---- Destructuring (declarations) ----
    {
      code: 'let {foo: foo} = obj;',
      errors: [
        {
          messageId: 'unnecessarilyRenamed',
          line: 1,
          column: 6,
        },
      ],
    },
    {
      code: 'let {foo: foo, bar: bar} = obj;',
      errors: [
        { messageId: 'unnecessarilyRenamed', line: 1, column: 6 },
        { messageId: 'unnecessarilyRenamed', line: 1, column: 16 },
      ],
    },
    {
      code: 'let {foo: {bar: bar}} = obj;',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 12 }],
    },
    {
      code: "let {'foo': foo} = obj;",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 6 }],
    },
    {
      code: 'let {foo: foo = 1} = obj;',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 6 }],
    },
    {
      code: 'function func({foo: foo}) {}',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 16 }],
    },
    {
      code: '({foo: foo}) => {}',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 3 }],
    },

    // ---- Destructuring (assignment pattern) ----
    {
      code: '({foo: foo} = obj);',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 3 }],
    },
    {
      code: '({foo: (foo)} = obj);',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 3 }],
    },
    {
      code: '({foo: foo = 1} = obj);',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 3 }],
    },

    // ---- Rest mixed with useless rename ----
    {
      code: 'const {foo: foo, ...stuff} = myObject;',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 8 }],
    },

    // ---- Imports ----
    {
      code: "import {foo as foo} from 'foo';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: "import {'foo' as foo} from 'foo';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: "import {foo as foo, bar as baz} from 'foo';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: "import {foo as foo, bar as bar} from 'foo';",
      errors: [
        { messageId: 'unnecessarilyRenamed', line: 1, column: 9 },
        { messageId: 'unnecessarilyRenamed', line: 1, column: 21 },
      ],
    },

    // ---- Exports ----
    {
      code: 'var foo = 0; export {foo as foo};',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 22 }],
    },
    {
      code: "var foo = 0; export {foo as 'foo'};",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 22 }],
    },
    {
      code: "export {'foo' as 'foo'} from 'bar';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: "export {foo as foo} from 'foo';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: "export {foo as foo, bar as bar} from 'foo';",
      errors: [
        { messageId: 'unnecessarilyRenamed', line: 1, column: 9 },
        { messageId: 'unnecessarilyRenamed', line: 1, column: 21 },
      ],
    },

    // ---- Comment cases (fix is suppressed, diagnostic still fires) ----
    {
      code: '({foo/**/ : foo} = {});',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 3 }],
    },
    {
      code: "import {foo/**/ as foo} from 'foo';",
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 9 }],
    },
    {
      code: 'let foo; export {foo/**/as foo};',
      errors: [{ messageId: 'unnecessarilyRenamed', line: 1, column: 18 }],
    },
  ],
});
