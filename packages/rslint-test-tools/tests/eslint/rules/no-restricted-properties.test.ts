import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-properties', {
  valid: [
    {
      code: 'someObject.someProperty',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
    },
    {
      code: 'anotherObject.disallowedProperty',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
    },
    {
      code: 'someObject.someProperty()',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
    },
    {
      code: 'anotherObject.disallowedProperty()',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
    },
    {
      code: 'anotherObject.disallowedProperty()',
      options: [
        {
          object: 'someObject',
          property: 'disallowedProperty',
          message: 'Please use someObject.allowedProperty instead.',
        },
      ] as any,
    },
    {
      code: "anotherObject['disallowedProperty']()",
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
    },
    {
      code: 'obj.toString',
      options: [{ object: 'obj', property: '__proto__' }] as any,
    },
    {
      code: 'toString.toString',
      options: [{ object: 'obj', property: 'foo' }] as any,
    },
    {
      code: 'obj.toString',
      options: [{ object: 'obj', property: 'foo' }] as any,
    },
    { code: 'foo.bar', options: [{ property: 'baz' }] as any },
    { code: 'foo.bar', options: [{ object: 'baz' }] as any },
    { code: 'foo()', options: [{ object: 'foo' }] as any },
    { code: 'foo;', options: [{ object: 'foo' }] as any },
    { code: 'foo[/(?<zero>0)/]', options: [{ property: 'null' }] as any },
    {
      code: 'let bar = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let {baz: bar} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let {unrelated} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let {baz: {bar: qux}} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let {bar} = foo.baz;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    { code: 'let {baz: bar} = foo;', options: [{ property: 'bar' }] as any },
    {
      code: 'let baz; ({baz: bar} = foo)',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let bar;',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let bar; ([bar = 5] = foo);',
      options: [{ object: 'foo', property: '1' }] as any,
    },
    {
      code: 'function qux({baz: bar} = foo) {}',
      options: [{ object: 'foo', property: 'bar' }] as any,
    },
    {
      code: 'let [bar, baz] = foo;',
      options: [{ object: 'foo', property: '1' }] as any,
    },
    {
      code: 'let [, bar] = foo;',
      options: [{ object: 'foo', property: '0' }] as any,
    },
    {
      code: 'let [, bar = 5] = foo;',
      options: [{ object: 'foo', property: '1' }] as any,
    },
    {
      code: 'let bar; ([bar = 5] = foo);',
      options: [{ object: 'foo', property: '0' }] as any,
    },
    {
      code: 'function qux([bar] = foo) {}',
      options: [{ object: 'foo', property: '0' }] as any,
    },
    {
      code: 'function qux([, bar] = foo) {}',
      options: [{ object: 'foo', property: '0' }] as any,
    },
    {
      code: 'function qux([, bar] = foo) {}',
      options: [{ object: 'foo', property: '1' }] as any,
    },
    {
      code: 'class C { #foo; foo() { this.#foo; } }',
      options: [{ property: '#foo' }] as any,
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        { property: 'disallowedProperty', allowObjects: ['someObject'] },
      ] as any,
    },
    {
      code: 'someObject.disallowedProperty; anotherObject.disallowedProperty();',
      options: [
        {
          property: 'disallowedProperty',
          allowObjects: ['someObject', 'anotherObject'],
        },
      ] as any,
    },
    {
      code: 'someObject.disallowedProperty()',
      options: [
        { property: 'disallowedProperty', allowObjects: ['someObject'] },
      ] as any,
    },
    {
      code: "someObject['disallowedProperty']()",
      options: [
        { property: 'disallowedProperty', allowObjects: ['someObject'] },
      ] as any,
    },
    {
      code: 'let {bar} = foo;',
      options: [{ property: 'bar', allowObjects: ['foo'] }] as any,
    },
    {
      code: 'let {baz: bar} = foo;',
      options: [{ property: 'baz', allowObjects: ['foo'] }] as any,
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        { object: 'someObject', allowProperties: ['disallowedProperty'] },
      ] as any,
    },
    {
      code: 'someObject.disallowedProperty; someObject.anotherDisallowedProperty();',
      options: [
        {
          object: 'someObject',
          allowProperties: ['disallowedProperty', 'anotherDisallowedProperty'],
        },
      ] as any,
    },
    {
      code: 'someObject.disallowedProperty()',
      options: [
        { object: 'someObject', allowProperties: ['disallowedProperty'] },
      ] as any,
    },
    {
      code: "someObject['disallowedProperty']()",
      options: [
        { object: 'someObject', allowProperties: ['disallowedProperty'] },
      ] as any,
    },
    {
      code: 'let {bar} = foo;',
      options: [{ object: 'foo', allowProperties: ['bar'] }] as any,
    },
    {
      code: 'let {baz: bar} = foo;',
      options: [{ object: 'foo', allowProperties: ['baz'] }] as any,
    },
  ],

  invalid: [
    {
      code: 'someObject.disallowedProperty',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
      ] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        {
          object: 'someObject',
          property: 'disallowedProperty',
          message: 'Please use someObject.allowedProperty instead.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()',
      options: [
        { object: 'someObject', property: 'disallowedProperty' },
        { object: 'anotherObject', property: 'anotherDisallowedProperty' },
      ] as any,
      errors: [
        { messageId: 'restrictedObjectProperty' },
        { messageId: 'restrictedObjectProperty' },
      ],
    },
    {
      code: 'foo.__proto__',
      options: [
        {
          property: '__proto__',
          message: 'Please use Object.getPrototypeOf instead.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: "foo['__proto__']",
      options: [
        {
          property: '__proto__',
          message: 'Please use Object.getPrototypeOf instead.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'foo.bar.baz;',
      options: [{ object: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'foo.bar();',
      options: [{ object: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'foo.bar.baz();',
      options: [{ object: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'foo.bar.baz;',
      options: [{ property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'foo.bar();',
      options: [{ property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'foo.bar.baz();',
      options: [{ property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'foo[/(?<zero>0)/]',
      options: [{ property: '/(?<zero>0)/' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: "require.call({}, 'foo')",
      options: [
        { object: 'require', message: 'Please call require() directly.' },
      ] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: "require['resolve']",
      options: [{ object: 'require' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar: baz} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: "let {'bar': baz} = foo;",
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar: {baz: qux}} = foo;',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar} = foo;',
      options: [{ object: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar: baz} = foo;',
      options: [{ object: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let {bar} = foo;',
      options: [{ property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'let bar; ({bar} = foo);',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'let bar; ({bar: baz = 1} = foo);',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'function qux({bar} = foo) {}',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'function qux({bar: baz} = foo) {}',
      options: [{ object: 'foo', property: 'bar' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: "var {['foo']: qux, bar} = baz",
      options: [{ object: 'baz', property: 'foo' }] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: "obj['#foo']",
      options: [{ property: '#foo' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'const { bar: { bad } = {} } = foo;',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'const { bar: { bad } } = foo;',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'const { bad } = foo();',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bad } = foo());',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bar: { bad } } = foo);',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bar: { bad } = {} } = foo);',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bad }) => {};',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bad } = {}) => {};',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bad: bar }) => {};',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '({ bar: { bad } = {} }) => {};',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: '[{ bad }] = foo;',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'const [{ bad }] = foo;',
      options: [{ property: 'bad' }] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        {
          property: 'disallowedProperty',
          allowObjects: ['anotherObject'],
        },
      ] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        {
          property: 'disallowedProperty',
          allowObjects: ['anotherObject'],
          message: 'Please use someObject.allowedProperty instead.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedProperty' }],
    },
    {
      code: 'someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()',
      options: [
        { property: 'disallowedProperty', allowObjects: ['anotherObject'] },
        {
          property: 'anotherDisallowedProperty',
          allowObjects: ['someObject'],
        },
      ] as any,
      errors: [
        { messageId: 'restrictedProperty' },
        { messageId: 'restrictedProperty' },
      ],
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        { object: 'someObject', allowProperties: ['allowedProperty'] },
      ] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'someObject.disallowedProperty',
      options: [
        {
          object: 'someObject',
          allowProperties: ['allowedProperty'],
          message: 'Please use someObject.allowedProperty instead.',
        },
      ] as any,
      errors: [{ messageId: 'restrictedObjectProperty' }],
    },
    {
      code: 'someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()',
      options: [
        {
          object: 'someObject',
          allowProperties: ['anotherDisallowedProperty'],
        },
        { object: 'anotherObject', allowProperties: ['disallowedProperty'] },
      ] as any,
      errors: [
        { messageId: 'restrictedObjectProperty' },
        { messageId: 'restrictedObjectProperty' },
      ],
    },
  ],
});
