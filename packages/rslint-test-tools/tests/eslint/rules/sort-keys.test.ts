import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('sort-keys', {
  valid: [
    "var obj = {'':1, [``]:2}",
    "var obj = {[``]:1, '':2}",
    "var obj = {'':1, a:2}",
    'var obj = {[``]:1, a:2}',
    'var obj = {_:2, a:1, b:3} // default',
    'var obj = {a:1, b:3, c:2}',
    'var obj = {a:2, b:3, b_:1}',
    'var obj = {C:3, b_:1, c:2}',
    'var obj = {$:1, A:3, _:2, a:4}',
    "var obj = {1:1, '11':2, 2:4, A:3}",
    "var obj = {'#':1, 'Z':2, À:3, è:4}",
    "var obj = { [/(?<zero>0)/]: 1, '/(?<zero>0)/': 2 }",
    'var obj = {a:1, b:3, [a + b]: -1, c:2}',
    "var obj = {'':1, [f()]:2, a:3}",
    {
      code: "var obj = {a:1, [b++]:2, '':3}",
      options: ['desc'] as any,
    },
    'var obj = {a:1, ...z, b:1}',
    'var obj = {b:1, ...z, a:1}',
    'var obj = {...a, b:1, ...c, d:1}',
    'var obj = {...a, b:1, ...d, ...c, e:2, z:5}',
    'var obj = {b:1, ...c, ...d, e:2}',
    "var obj = {a:1, ...z, '':2}",
    {
      code: "var obj = {'':1, ...z, 'a':2}",
      options: ['desc'] as any,
    },
    'var obj = {...z, a:1, b:1}',
    'var obj = {...z, ...c, a:1, b:1}',
    'var obj = {a:1, b:1, ...z}',
    {
      code: 'var obj = {...z, ...x, a:1, ...c, ...d, f:5, e:4}',
      options: ['desc'] as any,
    },
    'function fn(...args) { return [...args].length; }',
    'function g() {}; function f(...args) { return g(...args); }',
    'let {a, b} = {}',
    'var obj = {a:1, b:{x:1, y:1}, c:1}',
    {
      code: 'var obj = {_:2, a:1, b:3} // asc',
      options: ['asc'] as any,
    },
    {
      code: 'var obj = {a:1, b:3, c:2}',
      options: ['asc'] as any,
    },
    {
      code: 'var obj = {a:2, b:3, b_:1}',
      options: ['asc'] as any,
    },
    {
      code: 'var obj = {C:3, b_:1, c:2}',
      options: ['asc'] as any,
    },
    {
      code: 'var obj = {$:1, A:3, _:2, a:4}',
      options: ['asc'] as any,
    },
    {
      code: "var obj = {1:1, '11':2, 2:4, A:3}",
      options: ['asc'] as any,
    },
    {
      code: "var obj = {'#':1, 'Z':2, À:3, è:4}",
      options: ['asc'] as any,
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['asc', { minKeys: 4 }] as any,
    },
    {
      code: 'var obj = {_:2, a:1, b:3} // asc, insensitive',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:1, b:3, c:2}',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:2, b:3, b_:1}',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, C:3, c:2}',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, c:3, C:2}',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: "var obj = {1:1, '11':2, 2:4, A:3}",
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: "var obj = {'#':1, 'Z':2, À:3, è:4}",
      options: ['asc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {$:1, A:3, _:2, a:4}',
      options: ['asc', { caseSensitive: false, minKeys: 5 }] as any,
    },
    {
      code: 'var obj = {_:2, a:1, b:3} // asc, natural',
      options: ['asc', { natural: true }] as any,
    },
    {
      code: 'var obj = {a:1, b:3, c:2}',
      options: ['asc', { natural: true }] as any,
    },
    {
      code: 'var obj = {a:2, b:3, b_:1}',
      options: ['asc', { natural: true }] as any,
    },
    {
      code: 'var obj = {C:3, b_:1, c:2}',
      options: ['asc', { natural: true }] as any,
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['asc', { natural: true }] as any,
    },
    {
      code: "var obj = {1:1, 2:4, '11':2, A:3}",
      options: ['asc', { natural: true }] as any,
    },
    {
      code: "var obj = {'#':1, 'Z':2, À:3, è:4}",
      options: ['asc', { natural: true }] as any,
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['asc', { natural: true, minKeys: 4 }] as any,
    },
    {
      code: 'var obj = {_:2, a:1, b:3} // asc, natural, insensitive',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:1, b:3, c:2}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:2, b:3, b_:1}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, C:3, c:2}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, c:3, C:2}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: "var obj = {1:1, 2:4, '11':2, A:3}",
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: "var obj = {'#':1, 'Z':2, À:3, è:4}",
      options: ['asc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: [
        'asc',
        { natural: true, caseSensitive: false, minKeys: 4 },
      ] as any,
    },
    {
      code: 'var obj = {b:3, a:1, _:2} // desc',
      options: ['desc'] as any,
    },
    {
      code: 'var obj = {c:2, b:3, a:1}',
      options: ['desc'] as any,
    },
    {
      code: 'var obj = {b_:1, b:3, a:2}',
      options: ['desc'] as any,
    },
    {
      code: 'var obj = {c:2, b_:1, C:3}',
      options: ['desc'] as any,
    },
    {
      code: 'var obj = {a:4, _:2, A:3, $:1}',
      options: ['desc'] as any,
    },
    {
      code: "var obj = {A:3, 2:4, '11':2, 1:1}",
      options: ['desc'] as any,
    },
    {
      code: "var obj = {è:4, À:3, 'Z':2, '#':1}",
      options: ['desc'] as any,
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['desc', { minKeys: 4 }] as any,
    },
    {
      code: 'var obj = {b:3, a:1, _:2} // desc, insensitive',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {c:2, b:3, a:1}',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, b:3, a:2}',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {c:2, C:3, b_:1}',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {C:2, c:3, b_:1}',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:4, A:3, _:2, $:1}',
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: "var obj = {A:3, 2:4, '11':2, 1:1}",
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: "var obj = {è:4, À:3, 'Z':2, '#':1}",
      options: ['desc', { caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['desc', { caseSensitive: false, minKeys: 5 }] as any,
    },
    {
      code: 'var obj = {b:3, a:1, _:2} // desc, natural',
      options: ['desc', { natural: true }] as any,
    },
    {
      code: 'var obj = {c:2, b:3, a:1}',
      options: ['desc', { natural: true }] as any,
    },
    {
      code: 'var obj = {b_:1, b:3, a:2}',
      options: ['desc', { natural: true }] as any,
    },
    {
      code: 'var obj = {c:2, b_:1, C:3}',
      options: ['desc', { natural: true }] as any,
    },
    {
      code: 'var obj = {a:4, A:3, _:2, $:1}',
      options: ['desc', { natural: true }] as any,
    },
    {
      code: "var obj = {A:3, '11':2, 2:4, 1:1}",
      options: ['desc', { natural: true }] as any,
    },
    {
      code: "var obj = {è:4, À:3, 'Z':2, '#':1}",
      options: ['desc', { natural: true }] as any,
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['desc', { natural: true, minKeys: 4 }] as any,
    },
    {
      code: 'var obj = {b:3, a:1, _:2} // desc, natural, insensitive',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {c:2, b:3, a:1}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {b_:1, b:3, a:2}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {c:2, C:3, b_:1}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {C:2, c:3, b_:1}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:4, A:3, _:2, $:1}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: "var obj = {A:3, '11':2, 2:4, 1:1}",
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: "var obj = {è:4, À:3, 'Z':2, '#':1}",
      options: ['desc', { natural: true, caseSensitive: false }] as any,
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: [
        'desc',
        { natural: true, caseSensitive: false, minKeys: 4 },
      ] as any,
    },
    {
      code: '\n                var obj = {\n                    e: 1,\n                    f: 2,\n                    g: 3,\n\n                    a: 4,\n                    b: 5,\n                    c: 6\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b: 1,\n\n                    // comment\n                    a: 2,\n                    c: 3\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b: 1\n\n                    ,\n\n                    // comment\n                    a: 2,\n                    c: 3\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    b() {\n                    },\n                    e: 4\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    c: 1,\n                    d: 2,\n                    // comment\n\n                    // comment\n                    b() {\n                    },\n                    e: 4\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                  b,\n\n                  [a+b]: 1,\n                  a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    a() {\n\n                    },\n\n                    // abce\n                    f: 3,\n\n                    /*\n\n                    */\n                    [a+b]: 1,\n                    cc: 1,\n                    e: 2\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b: "/*",\n\n                    a: "*/",\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b,\n                    /*\n                    */ //\n\n                    a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b,\n\n                    /*\n                    */ //\n                    a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b: 1\n\n                    ,a: 2\n                };\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                    b: 1\n                // comment before comma\n\n                ,\n                a: 2\n                };\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                  b,\n\n                  a,\n                  ...z,\n                  c\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: '\n                var obj = {\n                  b,\n\n                  [foo()]: [\n\n                  ],\n                  a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
    },
    {
      code: "var obj = { ['b']: 1, a: 2 }",
      options: ['asc', { ignoreComputedKeys: true }] as any,
    },
    {
      code: 'var obj = { a: 1, [c]: 2, b: 3 }',
      options: ['asc', { ignoreComputedKeys: true }] as any,
    },
    {
      code: "var obj = { c: 1, ['b']: 2, a: 3 }",
      options: ['asc', { ignoreComputedKeys: true }] as any,
    },
  ],
  invalid: [
    {
      code: "var obj = {a:1, '':2} // default",
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {a:1, [``]:2} // default',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // default',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var obj = { null: 1, [/(?<zero>0)/]: 2 }',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'var obj = {...z, c:1, b:1}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {...z, ...c, d:4, b:1, ...y, ...f, e:2, a:1}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'var obj = {c:1, b:1, ...a}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {...z, ...a, c:1, b:1}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var obj = {...z, b:1, a:1, ...d, ...c}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {...z, a:2, b:0, ...x, ...c}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {...z, a:2, b:0, ...x}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "var obj = {...z, '':1, a:2}",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "var obj = {a:1, [b+c]:2, '':3}",
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "var obj = {'':1, [b+c]:2, a:3}",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "var obj = {b:1, [f()]:2, '':3, a:4}",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'var obj = {a:1, b:3, [a]: -1, c:2}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:{y:1, x:1}, b:1}',
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // asc',
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['asc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['asc', { minKeys: 3 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // asc, insensitive',
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, A:3, _:2, a:4}',
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['asc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['asc', { caseSensitive: false, minKeys: 3 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // asc, natural',
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {$:1, A:3, _:2, a:4}',
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['asc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['asc', { natural: true, minKeys: 2 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // asc, natural, insensitive',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, A:3, _:2, a:4}',
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, '11':2, 2:4, A:3}",
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['asc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: [
        'asc',
        { natural: true, caseSensitive: false, minKeys: 3 },
      ] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "var obj = {'':1, a:'2'} // desc",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "var obj = {[``]:1, a:'2'} // desc",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // desc',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['desc'] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['desc', { minKeys: 3 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // desc, insensitive',
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['desc', { caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['desc', { caseSensitive: false, minKeys: 2 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // desc, natural',
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, A:3, '11':2}",
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['desc', { natural: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: ['desc', { natural: true, minKeys: 3 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3} // desc, natural, insensitive',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var obj = {a:1, c:2, b:3}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, a:2, b:3}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var obj = {b_:1, c:2, C:3}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = {$:1, _:2, A:3, a:4}',
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "var obj = {1:1, 2:4, '11':2, A:3}",
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
      options: ['desc', { natural: true, caseSensitive: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'sortKeys',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {a:1, _:2, b:3}',
      options: [
        'desc',
        { natural: true, caseSensitive: false, minKeys: 2 },
      ] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n                var obj = {\n                    b: 1,\n                    c: 2,\n                    a: 3\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                let obj = {\n                    b\n\n                    ,a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: false }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 5,
          column: 22,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n                let obj = {\n                    b\n\n                    ,a\n                }\n            ',
      errors: [
        {
          messageId: 'sortKeys',
          line: 5,
          column: 22,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n                 var obj = {\n                    b: 1,\n                    c () {\n\n                    },\n                    a: 3\n                  }\n             ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 7,
          column: 21,
          endLine: 7,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                 var obj = {\n                    a: 1,\n                    b: 2,\n\n                    z () {\n\n                    },\n                    y: 3\n                  }\n             ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 9,
          column: 21,
          endLine: 9,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                 var obj = {\n                    b: 1,\n                    c () {\n                    },\n                    // comment\n                    a: 3\n                  }\n             ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 7,
          column: 21,
          endLine: 7,
          endColumn: 22,
        },
      ],
    },
    {
      code: "\n                var obj = {\n                  b,\n                  [a+b]: 1,\n                  a // sort-keys: 'a' should be before 'b'\n                }\n            ",
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 20,
        },
      ],
    },
    {
      code: '\n                var obj = {\n                    c: 1,\n                    d: 2,\n                    // comment\n                    // comment\n                    b() {\n                    },\n                    e: 4\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 7,
          column: 21,
          endLine: 7,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    z() {\n\n                    },\n                    f: 3,\n                    /*\n\n\n                    */\n                    [a+b]: 1,\n                    b: 1,\n                    e: 2\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 9,
          column: 21,
          endLine: 9,
          endColumn: 22,
        },
        {
          messageId: 'sortKeys',
          line: 15,
          column: 21,
          endLine: 15,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                var obj = {\n                    b: "/*",\n                    a: "*/",\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n                var obj = {\n                    b: 1\n                    // comment before comma\n                    , a: 2\n                };\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 5,
          column: 23,
          endLine: 5,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n                let obj = {\n                  b,\n                  [foo()]: [\n                  // ↓ this blank is inside a property and therefore should not count\n\n                  ],\n                  a\n                }\n            ',
      options: ['asc', { allowLineSeparatedGroups: true }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 8,
          column: 19,
          endLine: 8,
          endColumn: 20,
        },
      ],
    },
    {
      code: "var obj = { d: 1, ['c']: 2, b: 3, a: 4 }",
      options: ['asc', { ignoreComputedKeys: true, minKeys: 4 }] as any,
      errors: [
        {
          messageId: 'sortKeys',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
  ],
});
