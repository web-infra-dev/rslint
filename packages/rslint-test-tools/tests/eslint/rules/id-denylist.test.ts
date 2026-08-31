import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('id-denylist', {
  valid: [
    {
      code: 'foo = "bar"',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'bar = "bar"',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo = "bar"',
      options: ['f', 'fo', 'fooo', 'bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'function foo(){}',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo()',
      options: ['f', 'fo', 'fooo', 'bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: "import { foo as bar } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: "export { foo as bar } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'foo.bar()',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'baz'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var foo = bar.baz;',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var foo = bar.baz.bing;',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo.bar.baz = bing.bong.bash;',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'if (foo.bar) {}',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var obj = { key: foo.bar };',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'const {foo: bar} = baz',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'const {foo: {bar: baz}} = qux',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'function foo({ bar: baz }) {}',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'function foo({ bar: {baz: qux} }) {}',
      options: ['bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'function foo({baz} = obj.qux) {}',
      options: ['qux'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'function foo({ foo: {baz} = obj.qux }) {}',
      options: ['qux'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: '({a: bar = obj.baz});',
      options: ['baz'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: '({foo: {a: bar = obj.baz}} = qux);',
      options: ['baz'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'var arr = [foo.bar];',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: '[foo.bar]',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: '[foo.bar.nesting]',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'if (foo.bar === bar.baz) { [foo.bar] }',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var myArray = new Array(); var myDate = new Date();',
      options: ['array', 'date', 'mydate', 'myarray', 'new', 'var'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo()',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo.bar()',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo.bar',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: '({foo: obj.bar.bar.bar.baz} = {});',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: '({[obj.bar]: a = baz} = qux);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
    },

    // references to global variables
    {
      code: 'Number.parseInt()',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'x = Number.NaN;',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var foo = undefined;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'if (foo === undefined);',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'obj[undefined] = 5;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'foo = { [myGlobal]: 1 };',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 6, globals: { myGlobal: 'readonly' } },
    },
    {
      code: '({ myGlobal } = foo);',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 6, globals: { myGlobal: 'writable' } },
    },
    {
      code: '/* global myGlobal: readonly */ myGlobal = 5;',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 5 },
    },
    {
      code: 'var foo = [Map];',
      options: ['Map'] as any,
      languageOptions: { ecmaVersion: 6 },
    },
    {
      code: 'var foo = { bar: window.baz };',
      options: ['window'] as any,
      languageOptions: { ecmaVersion: 5, globals: { window: 'readonly' } },
    },

    // Class fields
    {
      code: 'class C { camelCase; #camelCase; #camelCase2() {} }',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 2022 },
    },
    {
      code: 'class C { snake_case; #snake_case; #snake_case2() {} }',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 2022 },
    },

    // Meta-properties
    {
      code: 'import.meta',
      options: ['import', 'meta'] as any,
      languageOptions: { ecmaVersion: 2020 },
    },
    {
      code: 'function foo() { new.target; }',
      options: ['new', 'target'] as any,
      languageOptions: { ecmaVersion: 6 },
    },

    // Import attribute keys
    {
      code: "import foo from 'foo.json' with { type: 'json' }",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
    {
      code: "export * from 'foo.json' with { type: 'json' }",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
    {
      code: "export { default } from 'foo.json' with { type: 'json' }",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
    {
      code: "import('foo.json', { with: { type: 'json' } })",
      options: ['with', 'type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
    {
      code: "import('foo.json', { 'with': { type: 'json' } })",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
    {
      code: "import('foo.json', { with: { type } })",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
    },
  ],
  invalid: [
    {
      code: 'foo = "bar"',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 1 }],
    },
    {
      code: 'bar = "bar"',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 1 }],
    },
    {
      code: 'foo = "bar"',
      options: ['f', 'fo', 'foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 1 }],
    },
    {
      code: 'function foo(){}',
      options: ['f', 'fo', 'foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: "import foo from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 8 }],
    },
    {
      code: "import * as foo from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: "export * as foo from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 2020 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: "import { foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: "import { foo as bar } from 'mod'",
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "import { foo as bar } from 'mod'",
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "import { foo as foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "import { foo, foo as bar } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: "import { foo as bar, foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 22 }],
    },
    {
      code: "import foo, { foo as bar } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 8 }],
    },
    {
      code: 'var foo; export { foo as bar };',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 26 }],
    },
    {
      code: 'var foo; export { foo };',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 19 },
      ],
    },
    {
      code: 'var foo; export { foo as bar };',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 19 },
      ],
    },
    {
      code: 'var foo; export { foo as foo };',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 19 },
        { messageId: 'restricted', line: 1, column: 26 },
      ],
    },
    {
      code: 'var foo; export { foo as bar };',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 19 },
        { messageId: 'restricted', line: 1, column: 26 },
      ],
    },
    {
      code: "export { foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: "export { foo as bar } from 'mod'",
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "export { foo as bar } from 'mod'",
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "export { foo as foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: "export { foo, foo as bar } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: "export { foo as bar, foo } from 'mod'",
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 22 }],
    },
    {
      code: 'foo.bar()',
      options: ['f', 'fo', 'foo', 'b', 'ba', 'baz'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 1 }],
    },
    {
      code: 'foo[bar] = baz;',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'baz = foo[bar];',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 11 }],
    },
    {
      code: 'var foo = bar.baz;',
      options: ['f', 'fo', 'foo', 'b', 'ba', 'barr', 'bazz'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'var foo = bar.baz;',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'bar', 'bazz'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 11 }],
    },
    {
      code: 'if (foo.bar) {}',
      options: ['f', 'fo', 'foo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'var obj = { key: foo.bar };',
      options: ['obj'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'var obj = { key: foo.bar };',
      options: ['key'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'var obj = { key: foo.bar };',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 18 }],
    },
    {
      code: 'var arr = [foo.bar];',
      options: ['arr'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'var arr = [foo.bar];',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 12 }],
    },
    {
      code: '[foo.bar]',
      options: ['f', 'fo', 'foo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 2 }],
    },
    {
      code: 'if (foo.bar === bar.baz) { [bing.baz] }',
      options: ['f', 'fo', 'foo', 'b', 'ba', 'barr', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'if (foo.bar === bar.baz) { [foo.bar] }',
      options: ['f', 'fo', 'fooo', 'b', 'ba', 'bar', 'bazz', 'bingg'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: 'var myArray = new Array(); var myDate = new Date();',
      options: ['array', 'date', 'myDate', 'myarray', 'new', 'var'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 32 }],
    },
    {
      code: 'var myArray = new Array(); var myDate = new Date();',
      options: ['array', 'date', 'mydate', 'myArray', 'new', 'var'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'foo.bar = 1',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'foo.bar.baz = 1',
      options: ['bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 9 }],
    },
    {
      code: 'const {foo} = baz',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 8 }],
    },
    {
      code: 'const {foo: bar} = baz',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'const {[foo]: bar} = baz',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 9 },
        { messageId: 'restricted', line: 1, column: 15 },
      ],
    },
    {
      code: 'const {foo: {bar: baz}} = qux',
      options: ['foo', 'bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 19 }],
    },
    {
      code: 'const {foo: {[bar]: baz}} = qux',
      options: ['foo', 'bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 15 },
        { messageId: 'restricted', line: 1, column: 21 },
      ],
    },
    {
      code: 'const {[foo]: {[bar]: baz}} = qux',
      options: ['foo', 'bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 9 },
        { messageId: 'restricted', line: 1, column: 17 },
        { messageId: 'restricted', line: 1, column: 23 },
      ],
    },
    {
      code: 'function foo({ bar: baz }) {}',
      options: ['bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 21 }],
    },
    {
      code: 'function foo({ bar: {baz: qux} }) {}',
      options: ['bar', 'baz', 'qux'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 27 }],
    },
    {
      code: '({foo: obj.bar} = baz);',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 12 }],
    },
    {
      code: '({foo: obj.bar.bar.bar.baz} = {});',
      options: ['foo', 'bar', 'baz'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 24 }],
    },
    {
      code: '({[foo]: obj.bar} = baz);',
      options: ['foo', 'bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 4 },
        { messageId: 'restricted', line: 1, column: 14 },
      ],
    },
    {
      code: '({foo: { a: obj.bar }} = baz);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: '({a: obj.bar = baz} = qux);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: '({a: obj.bar.bar.baz = obj.qux} = obj.qux);',
      options: ['a', 'bar', 'baz', 'qux'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 18 }],
    },
    {
      code: '({a: obj[bar] = obj.qux} = obj.qux);',
      options: ['a', 'bar', 'baz', 'qux'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: '({a: [obj.bar] = baz} = qux);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 11 }],
    },
    {
      code: '({foo: { a: obj.bar = baz}} = qux);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 17 }],
    },
    {
      code: '({foo: { [a]: obj.bar }} = baz);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 19 }],
    },
    {
      code: '({...obj.bar} = baz);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 9 },
      errors: [{ messageId: 'restricted', line: 1, column: 10 }],
    },
    {
      code: '([obj.bar] = baz);',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 7 }],
    },
    {
      code: 'const [bar] = baz;',
      options: ['bar'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 8 }],
    },

    // not a reference to a global variable, because it isn't a reference to a variable
    {
      code: 'foo.undefined = 1;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 5 }],
    },
    {
      code: 'var foo = { undefined: 1 };',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'var foo = { undefined: undefined };',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'var foo = { Number() {} };',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'class Foo { Number() {} }',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },
    {
      code: 'myGlobal: while(foo) { break myGlobal; } ',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 5, globals: { myGlobal: 'readonly' } },
      errors: [
        { messageId: 'restricted', line: 1, column: 1 },
        { messageId: 'restricted', line: 1, column: 30 },
      ],
    },

    // globals declared in the given source code are not excluded from consideration
    {
      code: 'const foo = 1; bar = foo;',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 7 },
        { messageId: 'restricted', line: 1, column: 22 },
      ],
    },
    {
      code: 'let foo; foo = bar;',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 10 },
      ],
    },
    {
      code: 'bar = foo; var foo;',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [
        { messageId: 'restricted', line: 1, column: 7 },
        { messageId: 'restricted', line: 1, column: 16 },
      ],
    },
    {
      code: 'function foo() {} var bar = foo;',
      options: ['foo'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [
        { messageId: 'restricted', line: 1, column: 10 },
        { messageId: 'restricted', line: 1, column: 29 },
      ],
    },
    {
      code: 'class Foo {} var bar = Foo;',
      options: ['Foo'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 7 },
        { messageId: 'restricted', line: 1, column: 24 },
      ],
    },

    // redeclared globals are not excluded from consideration
    {
      code: 'let undefined; undefined = 1;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 5 },
        { messageId: 'restricted', line: 1, column: 16 },
      ],
    },
    {
      code: 'foo = undefined; var undefined;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [
        { messageId: 'restricted', line: 1, column: 7 },
        { messageId: 'restricted', line: 1, column: 22 },
      ],
    },
    {
      code: 'function undefined(){} x = undefined;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [
        { messageId: 'restricted', line: 1, column: 10 },
        { messageId: 'restricted', line: 1, column: 28 },
      ],
    },
    {
      code: 'class Number {} x = Number.NaN;',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 7 },
        { messageId: 'restricted', line: 1, column: 21 },
      ],
    },

    // assignment to a property with a restricted name creates a global with that name
    {
      code: '/* globals myGlobal */ window.myGlobal = 5; foo = myGlobal;',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 5, globals: { window: 'readonly' } },
      errors: [{ messageId: 'restricted', line: 1, column: 31 }],
    },

    // disabled global variables
    {
      code: 'var foo = undefined;',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5, globals: { undefined: 'off' } },
      errors: [{ messageId: 'restricted', line: 1, column: 11 }],
    },
    {
      code: '/* globals Number: off */ Number.parseInt()',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 27 }],
    },
    {
      code: 'var foo = [Map];',
      options: ['Map'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 12 }],
    },

    // shadowed global variables
    {
      code: 'if (foo) { let undefined; bar = undefined; }',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 16 },
        { messageId: 'restricted', line: 1, column: 33 },
      ],
    },
    {
      code: 'function foo(Number) { var x = Number.NaN; }',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [
        { messageId: 'restricted', line: 1, column: 14 },
        { messageId: 'restricted', line: 1, column: 32 },
      ],
    },
    {
      code: 'function foo() { var myGlobal; x = myGlobal; }',
      options: ['myGlobal'] as any,
      languageOptions: { ecmaVersion: 5, globals: { myGlobal: 'readonly' } },
      errors: [
        { messageId: 'restricted', line: 1, column: 22 },
        { messageId: 'restricted', line: 1, column: 36 },
      ],
    },
    {
      code: 'function foo(bar) { return Number.parseInt(bar); } const Number = 1;',
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 28 },
        { messageId: 'restricted', line: 1, column: 58 },
      ],
    },
    {
      code: "import Number from 'myNumber'; const foo = Number.parseInt(bar);",
      options: ['Number'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [
        { messageId: 'restricted', line: 1, column: 8 },
        { messageId: 'restricted', line: 1, column: 44 },
      ],
    },
    {
      code: 'var foo = function undefined() {};',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 5 },
      errors: [{ messageId: 'restricted', line: 1, column: 20 }],
    },

    // a reference to a global variable that also creates a property with a restricted name
    {
      code: 'var foo = { undefined }',
      options: ['undefined'] as any,
      languageOptions: { ecmaVersion: 6 },
      errors: [{ messageId: 'restricted', line: 1, column: 13 }],
    },

    // Class fields
    {
      code: 'class C { camelCase; #camelCase; #camelCase2() {} }',
      options: ['camelCase'] as any,
      languageOptions: { ecmaVersion: 2022 },
      errors: [
        { messageId: 'restricted', line: 1, column: 11 },
        { messageId: 'restrictedPrivate', line: 1, column: 22 },
      ],
    },
    {
      code: 'class C { snake_case; #snake_case() {}; #snake_case2() {} }',
      options: ['snake_case'] as any,
      languageOptions: { ecmaVersion: 2022 },
      errors: [
        { messageId: 'restricted', line: 1, column: 11 },
        { messageId: 'restrictedPrivate', line: 1, column: 23 },
      ],
    },

    // Not an import attribute key
    {
      code: "import('foo.json', { with: { [type]: 'json' } })",
      options: ['type'] as any,
      languageOptions: { ecmaVersion: 2025 },
      errors: [{ messageId: 'restricted', line: 1, column: 31 }],
    },
    {
      code: "import('foo.json', { with: { type: json } })",
      options: ['json'] as any,
      languageOptions: { ecmaVersion: 2025 },
      errors: [{ messageId: 'restricted', line: 1, column: 36 }],
    },
  ],
});
