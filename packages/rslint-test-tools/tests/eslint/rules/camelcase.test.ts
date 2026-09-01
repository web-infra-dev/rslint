import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('camelcase', {
  valid: [
    {
      code: 'firstName = "Nicholas"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'FIRST_NAME = "Nicholas"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '__myPrivateVariable = "Patrick"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'myPrivateVariable_ = "Patrick"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'function doSomething(){}',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'do_something()',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'new do_something',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'new do_something()',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'foo.do_something()',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var foo = bar.baz_boom;',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var foo = bar.baz_boom.something;',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'foo.boom_pow.qux = bar.baz_boom.something;',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (bar.baz_boom) {}',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var obj = { key: foo.bar_baz };',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var arr = [foo.bar_baz];',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '[foo.bar_baz]',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var arr = [foo.bar_baz.qux];',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '[foo.bar_baz.nesting]',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (foo.bar_baz === boom.bam_pow) { [foo.baz_boom] }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {key: 1}',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {_leading: 1}',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {trailing_: 1}',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {bar_baz: 1}',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {_leading: 1}',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var o = {trailing_: 1}',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj.a_b = 2;',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj._a = 2;',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj.a_ = 2;',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj._a = 2;',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj.a_ = 2;',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var obj = {\n a_a: 1 \n};\n obj.a_b = 2;',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'obj.foo_bar = function(){};',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: "const { ['foo']: _foo } = obj;",
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'const { [_foo_]: foo } = obj;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { category_id } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { category_id: category_id } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { category_id = 1 } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { [{category_id} = query]: categoryId } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { category_id: category } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { _leading } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'var { trailing_ } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'import { camelCased } from "external module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { _leading } from "external module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { trailing_ } from "external module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { no_camelcased as camelCased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { no_camelcased as _leading } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { no_camelcased as trailing_ } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: 'import { no_camelcased as camelCased, anotherCamelCased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: "import { snake_cased } from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: "import { snake_cased as snake_cased } from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    {
      code: "import { 'snake_cased' as snake_cased } from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    {
      code: "import { camelCased } from 'mod'",
      options: {
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
    },
    {
      code: "export { a as 'snake_cased' } from 'mod'",
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    {
      code: "export * as 'snake_cased' from 'mod'",
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    {
      code: 'var _camelCased = aGlobalVariable',
      options: {
        ignoreGlobals: false,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          aGlobalVariable: 'readonly',
        },
      },
    },
    {
      code: 'var camelCased = _aGlobalVariable',
      options: {
        ignoreGlobals: false,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          _aGlobalVariable: 'readonly',
        },
      },
    },
    {
      code: 'var camelCased = a_global_variable',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'a_global_variable.foo()',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'a_global_variable[undefined]',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'var foo = a_global_variable.bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'a_global_variable.foo = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: '( { foo: a_global_variable.bar } = baz )',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'a_global_variable = foo',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: 'a_global_variable = foo',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: '({ a_global_variable } = foo)',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: '({ snake_cased: a_global_variable } = foo)',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: '({ snake_cased: a_global_variable = foo } = bar)',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: '[a_global_variable] = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: '[a_global_variable = foo] = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
    },
    {
      code: 'foo[a_global_variable] = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'var foo = { [a_global_variable]: bar }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'var { [a_global_variable]: foo } = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
    },
    {
      code: 'function foo({ no_camelcased: camelCased }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo({ no_camelcased: _leading }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo({ no_camelcased: trailing_ }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: "function foo({ camelCased = 'default value' }) {};",
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: "function foo({ _leading = 'default value' }) {};",
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: "function foo({ trailing_ = 'default value' }) {};",
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo({ camelCased }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo({ _leading }) {}',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo({ trailing_ }) {}',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'ignored_foo = 0;',
      options: {
        allow: ['ignored_foo'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'ignored_foo = 0; ignored_bar = 1;',
      options: {
        allow: ['ignored_foo', 'ignored_bar'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'user_id = 0;',
      options: {
        allow: ['_id$'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '__option_foo__ = 0;',
      options: {
        allow: ['__option_foo__'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '__option_foo__ = 0; user_id = 0; foo = 1',
      options: {
        allow: ['__option_foo__', '_id$'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'fo_o = 0;',
      options: {
        allow: ['__option_foo__', 'fo_o'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'user = 0;',
      options: {
        allow: [],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'foo = { [computedBar]: 0 };',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ a: obj.fo_o } = bar);',
      options: {
        allow: ['fo_o'],
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ a: obj.foo } = bar);',
      options: {
        allow: ['fo_o'],
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ a: obj.fo_o } = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ a: obj.fo_o.b_ar } = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ a: { b: obj.fo_o } } = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '([obj.fo_o] = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({ c: [ob.fo_o]} = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '([obj.fo_o.b_ar] = bar);',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '({obj} = baz.fo_o);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '([obj] = baz.fo_o);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: '([obj.foo = obj.fo_o] = bar);',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { camelCase; #camelCase; #camelCase2() {} }',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { snake_case; #snake_case; #snake_case2() {} }',
      options: {
        properties: 'never',
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: '\n            const { some_property } = obj;\n\n            const bar = { some_property };\n\n            obj.some_property = 10;\n\n            const xyz = { some_property: obj.some_property };\n\n            const foo = ({ some_property }) => {\n                console.log(some_property)\n            };\n            ',
      options: {
        properties: 'never',
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: '\n            const { some_property } = obj;\n            doSomething({ some_property });\n            ',
      options: {
        properties: 'never',
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: "import foo from 'foo.json' with { my_type: 'json' }",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'module',
      },
    },
    {
      code: "export * from 'foo.json' with { my_type: 'json' }",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'module',
      },
    },
    {
      code: "export { default } from 'foo.json' with { my_type: 'json' }",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'module',
      },
    },
    {
      code: "import('foo.json', { my_with: { my_type: 'json' } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
      },
    },
    {
      code: "import('foo.json', { 'with': { my_type: 'json' } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
      },
    },
    {
      code: "import('foo.json', { my_with: { my_type } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
      },
    },
    {
      code: "import('foo.json', { my_with: { my_type } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
        globals: {
          my_type: true,
        },
      },
    },
  ],
  invalid: [
    {
      code: 'first_name = "Nicholas"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '__private_first_name = "Patrick"',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'function foo_bar(){}',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'obj.foo_bar = function(){};',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'bar_baz.foo = function(){};',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '[foo_bar.baz]',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'if (foo.bar_baz === boom.bam_pow) { [foo_bar.baz] }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: 'foo.bar_baz = boom.bam_pow',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var foo = { bar_baz: boom.bam_pow }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'var foo = { bar_baz: boom.bam_pow }',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'foo.qux.boom_pow = { bar: boom.bam_pow }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'var o = {bar_baz: 1}',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'obj.a_b = 2;',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'var { category_id: category_alias } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'var { category_id: category_alias } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'var { [category_id]: categoryId } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var { [category_id]: categoryId } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var { category_id: categoryId, ...other_props } = query;',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2018,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'var { category_id } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var { category_id: category_id } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'var { category_id = 1 } = query;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'import no_camelcased from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'import * as no_camelcased from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'import { no_camelcased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'import { no_camelcased as no_camel_cased } from "external module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'import { camelCased as no_camel_cased } from "external module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: "import { 'snake_cased' as snake_cased } from 'mod'",
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: "import { 'snake_cased' as another_snake_cased } from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'import { camelCased, no_camelcased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'import { no_camelcased as camelCased, another_no_camelcased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'import camelCased, { no_camelcased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'import no_camelcased, { another_no_camelcased as camelCased } from "external-module";',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "import snake_cased from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "import * as snake_cased from 'mod'",
      options: {
        ignoreImports: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "import snake_cased from 'mod'",
      options: {
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "import * as snake_cased from 'mod'",
      options: {
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var camelCased = snake_cased',
      options: {
        ignoreGlobals: false,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          snake_cased: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'a_global_variable.foo()',
      options: {
        ignoreGlobals: false,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          snake_cased: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'a_global_variable[undefined]',
      options: {
        ignoreGlobals: false,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          snake_cased: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var camelCased = snake_cased',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          snake_cased: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'var camelCased = snake_cased',
      options: {},
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          snake_cased: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'foo.a_global_variable = bar',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'var foo = { a_global_variable: bar }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var foo = { a_global_variable: a_global_variable }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var foo = { a_global_variable() {} }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'class Foo { a_global_variable() {} }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'a_global_variable: for (;;);',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'if (foo) { let a_global_variable; a_global_variable = bar; }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 33,
        },
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'function foo(a_global_variable) { foo = a_global_variable; }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: 'var a_global_variable',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'function a_global_variable () {}',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'const a_global_variable = foo; bar = a_global_variable',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'bar = a_global_variable; var a_global_variable;',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'writable',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'var foo = { a_global_variable }',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
        globals: {
          a_global_variable: 'readonly',
        },
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'undefined_variable;',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'implicit_global = 1;',
      options: {
        ignoreGlobals: true,
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "export * as snake_cased from 'mod'",
      languageOptions: {
        ecmaVersion: 2020,
        sourceType: 'module',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'function foo({ no_camelcased }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: "function foo({ no_camelcased = 'default value' }) {};",
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'const no_camelcased = 0; function foo({ camelcased_value = no_camelcased}) {}',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'const { bar: no_camelcased } = foo;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'function foo({ value_1: my_default }) {}',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'function foo({ isCamelcased: no_camelcased }) {};',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'var { foo: bar_baz = 1 } = quz;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'const { no_camelcased = false } = bar;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'const { no_camelcased = foo_bar } = bar;',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'not_ignored_foo = 0;',
      options: {
        allow: ['ignored_bar'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'not_ignored_foo = 0;',
      options: {
        allow: ['_id$'],
      },
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'foo = { [computed_bar]: 0 };',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '({ a: obj.fo_o } = bar);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '({ a: obj.fo_o } = bar);',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '({ a: obj.fo_o.b_ar } = baz);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '({ a: { b: { c: obj.fo_o } } } = bar);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '({ a: { b: { c: obj.fo_o.b_ar } } } = baz);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '([obj.fo_o] = bar);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '([obj.fo_o] = bar);',
      options: {
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '([obj.fo_o = 1] = bar);',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '({ a: [obj.fo_o] } = bar);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '({ a: { b: [obj.fo_o] } } = bar);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '([obj.fo_o.ba_r] = baz);',
      languageOptions: {
        ecmaVersion: 6,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '({...obj.fo_o} = baz);',
      languageOptions: {
        ecmaVersion: 9,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '({...obj.fo_o.ba_r} = baz);',
      languageOptions: {
        ecmaVersion: 9,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '({c: {...obj.fo_o }} = baz);',
      languageOptions: {
        ecmaVersion: 9,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'obj.o_k.non_camelcase = 0',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2020,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '(obj?.o_k).non_camelcase = 0',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2020,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class C { snake_case; }',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'class C { #snake_case; foo() { this.#snake_case; } }',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCasePrivate',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'class C { #snake_case() {} }',
      options: {
        properties: 'always',
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCasePrivate',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '\n            const { some_property } = obj;\n            doSomething({ some_property });\n            ',
      options: {
        properties: 'always',
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 40,
        },
      ],
    },
    {
      code: '\n            const { some_property } = obj;\n            doSomething({ some_property });\n            doSomething({ [some_property]: "bar" });\n            ',
      options: {
        properties: 'never',
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 4,
          column: 28,
          endLine: 4,
          endColumn: 41,
        },
      ],
    },
    {
      code: '\n            const { some_property } = obj;\n\n            const bar = { some_property };\n\n            obj.some_property = 10;\n\n            const xyz = { some_property: obj.some_property };\n\n            const foo = ({ some_property }) => {\n                console.log(some_property)\n            };\n            ',
      options: {
        properties: 'always',
        ignoreDestructuring: true,
      },
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 4,
          column: 27,
          endLine: 4,
          endColumn: 40,
        },
        {
          messageId: 'notCamelCase',
          line: 6,
          column: 17,
          endLine: 6,
          endColumn: 30,
        },
        {
          messageId: 'notCamelCase',
          line: 8,
          column: 27,
          endLine: 8,
          endColumn: 40,
        },
      ],
    },
    {
      code: "import('foo.json', { my_with: { [my_type]: 'json' } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: "import('foo.json', { my_with: { my_type: my_json } })",
      options: {
        properties: 'always',
        ignoreImports: false,
      },
      languageOptions: {
        ecmaVersion: 2025,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'notCamelCase',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
  ],
});
