import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('object-shorthand', {
  valid: [
    'var x = {y() {}}',
    'var x = {y}',
    'var x = {a: b}',
    `var x = {a: 'a'}`,
    `var x = {'a': 'a'}`,
    `var x = {'a': b}`,
    'var x = {y(x) {}}',
    'var {x,y,z} = x',
    'var {x: {y}} = z',
    'var x = {*x() {}}',
    'var x = {x: y}',
    'var x = {x: y, y: z}',
    'var x = {x() {}, y: z, l(){}}',
    'var x = {[y]: y}',
    'doSomething({x: y})',
    'doSomething({y() {}})',
    'doSomething({x: y, y() {}})',
    '!{ a: function a(){} };',

    // arrow functions are allowed by default
    'var x = {y: (x)=>x}',
    'doSomething({y: (x)=>x})',

    // getters and setters
    'var x = {get y() {}}',
    'var x = {set y(z) {}}',
    'var x = {get y() {}, set y(z) {}}',

    // properties option
    { code: 'var x = {[y]: y}', options: ['properties'] as any },
    { code: `var x = {['y']: 'y'}`, options: ['properties'] as any },
    { code: `var x = {['y']: y}`, options: ['properties'] as any },
    { code: 'var x = {y}', options: ['properties'] as any },
    { code: 'var x = {y: {b}}', options: ['properties'] as any },

    // methods option
    { code: 'var x = {[y]() {}}', options: ['methods'] as any },
    { code: 'var x = {[y]: function x() {}}', options: ['methods'] as any },
    { code: 'var x = {[y]: y}', options: ['methods'] as any },
    { code: 'var x = {y() {}}', options: ['methods'] as any },
    { code: 'var x = {x, y() {}, a:b}', options: ['methods'] as any },

    // never
    { code: 'var x = {a: n, c: d, f: g}', options: ['never'] as any },
    { code: 'var x = {a: function(){}, b: {c: d}}', options: ['never'] as any },

    // ignoreConstructors
    {
      code: 'var x = {ConstructorFunction: function(){}, a: b}',
      options: ['always', { ignoreConstructors: true }] as any,
    },
    {
      code: 'var x = {_ConstructorFunction: function(){}, a: b}',
      options: ['always', { ignoreConstructors: true }] as any,
    },
    {
      code: 'var x = {$ConstructorFunction: function(){}, a: b}',
      options: ['always', { ignoreConstructors: true }] as any,
    },
    {
      code: 'var x = {__ConstructorFunction: function(){}, a: b}',
      options: ['always', { ignoreConstructors: true }] as any,
    },
    {
      code: 'var x = {_0ConstructorFunction: function(){}, a: b}',
      options: ['always', { ignoreConstructors: true }] as any,
    },
    {
      code: 'var x = {notConstructorFunction(){}, b: c}',
      options: ['always', { ignoreConstructors: true }] as any,
    },

    // methodsIgnorePattern
    {
      code: 'var x = { foo: function() {}  }',
      options: ['always', { methodsIgnorePattern: '^foo$' }] as any,
    },
    {
      code: `var x = { 'foo': function() {}  }`,
      options: ['always', { methodsIgnorePattern: '^foo$' }] as any,
    },
    {
      code: `var x = { ['foo']: function() {}  }`,
      options: ['always', { methodsIgnorePattern: '^foo$' }] as any,
    },
    {
      code: 'var x = { 123: function() {}  }',
      options: ['always', { methodsIgnorePattern: '^123$' }] as any,
    },
    {
      code: 'var x = { afoob: function() {}  }',
      options: ['always', { methodsIgnorePattern: 'foo' }] as any,
    },

    // avoidQuotes
    {
      code: `var x = {'a': function(){}}`,
      options: ['always', { avoidQuotes: true }] as any,
    },
    {
      code: `var x = {['a']: function(){}}`,
      options: ['methods', { avoidQuotes: true }] as any,
    },
    {
      code: `var x = {'y': y}`,
      options: ['properties', { avoidQuotes: true }] as any,
    },

    // consistent
    { code: 'var x = {a: a, b: b}', options: ['consistent'] as any },
    { code: 'var x = {a: b, c: d, f: g}', options: ['consistent'] as any },
    { code: 'var x = {a, b}', options: ['consistent'] as any },
    {
      code: 'var x = {a, b, get test() { return 1; }}',
      options: ['consistent'] as any,
    },

    // consistent-as-needed
    { code: 'var x = {a, b}', options: ['consistent-as-needed'] as any },
    { code: `var x = {0: 'foo'}`, options: ['consistent-as-needed'] as any },
    {
      code: `var x = {'key': 'baz'}`,
      options: ['consistent-as-needed'] as any,
    },
    { code: `var x = {foo: 'foo'}`, options: ['consistent-as-needed'] as any },
    { code: 'var x = {[foo]: foo}', options: ['consistent-as-needed'] as any },
    {
      code: 'var x = {foo: function foo() {}}',
      options: ['consistent-as-needed'] as any,
    },

    // avoidExplicitReturnArrows
    {
      code: '({ x: () => foo })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
    },
    {
      code: '({ x() { return; } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
    },
    {
      code: '({ x: () => { this; } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
    },
    {
      code: 'function foo() { ({ x: () => { arguments; } }) }',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
    },
  ],
  invalid: [
    {
      code: 'var x = {x: x}',
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },
    {
      code: `var x = {'x': x}`,
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },
    {
      code: 'var x = {y: y, x: x}',
      errors: [
        { messageId: 'expectedPropertyShorthand' },
        { messageId: 'expectedPropertyShorthand' },
      ],
    },
    {
      code: 'var x = {y: function() {}}',
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: 'var x = {y: function*() {}}',
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: 'doSomething({x: x})',
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },
    {
      code: 'doSomething({y: function() {}})',
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: 'doSomething({[y]: function() {}})',
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },

    // never
    {
      code: 'var x = {y() {}}',
      options: ['never'] as any,
      errors: [{ messageId: 'expectedMethodLongform' }],
    },
    {
      code: 'var x = {*y() {}}',
      options: ['never'] as any,
      errors: [{ messageId: 'expectedMethodLongform' }],
    },
    {
      code: 'var x = {y}',
      options: ['never'] as any,
      errors: [{ messageId: 'expectedPropertyLongform' }],
    },

    // properties
    {
      code: 'var x = {x: x}',
      options: ['properties'] as any,
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },

    // methods
    {
      code: 'var x = {y: function() {}}',
      options: ['methods'] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },

    // avoidQuotes
    {
      code: 'var x = {a: function(){}}',
      options: ['methods', { avoidQuotes: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: `var x = {'a'(){}}`,
      options: ['always', { avoidQuotes: true }] as any,
      errors: [{ messageId: 'expectedLiteralMethodLongform' }],
    },

    // consistent
    {
      code: 'var x = {a: a, b}',
      options: ['consistent'] as any,
      errors: [{ messageId: 'unexpectedMix' }],
    },

    // consistent-as-needed
    {
      code: 'var x = {a: a, b: b}',
      options: ['consistent-as-needed'] as any,
      errors: [{ messageId: 'expectedAllPropertiesShorthanded' }],
    },
    {
      code: 'var x = {a, z: function z(){}}',
      options: ['consistent-as-needed'] as any,
      errors: [{ messageId: 'unexpectedMix' }],
    },

    // avoidExplicitReturnArrows
    {
      code: '({ x: () => { return; } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: '({ x: foo => { return; } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: '({ x: (foo = 1) => { return; } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },

    // Parenthesized values (tsgo preserves ParenthesizedExpression).
    {
      code: 'var x = {a: (a)}',
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },
    {
      code: 'var x = {a: (((a)))}',
      errors: [{ messageId: 'expectedPropertyShorthand' }],
    },
    {
      code: 'var x = {a: (function(){ return 1 })}',
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
    {
      code: '({ a: (() => { return 1 }) })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },

    // `arguments` at module scope is NOT a lexical identifier (no enclosing
    // function provides it). Arrow must still be convertible.
    {
      code: '({ x: () => { arguments } })',
      options: ['always', { avoidExplicitReturnArrows: true }] as any,
      errors: [{ messageId: 'expectedMethodShorthand' }],
    },
  ],
});
