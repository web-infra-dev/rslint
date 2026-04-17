import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('accessor-pairs', {
  valid: [
    // Destructuring patterns — not flagged
    {
      code: 'var { get: foo } = bar; ({ set: foo } = bar);',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: 'var { set } = foo; ({ get } = foo);',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },

    // Default: getWithoutSet=false, so lone getter is OK
    'var o = { get a() {} }',
    { code: 'var o = { get a() {} }', options: [{}] as any },

    // Both options disabled
    {
      code: 'var o = { get a() {} };',
      options: [{ setWithoutGet: false, getWithoutSet: false }] as any,
    },
    {
      code: 'var o = { set a(foo) {} };',
      options: [{ setWithoutGet: false, getWithoutSet: false }] as any,
    },

    // Valid pairs
    {
      code: 'var o = { get a() {}, set a(foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: 'var o = { set a(foo) {}, get a() {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: "var o = { get 'a'() {}, set 'a'(foo) {} };",
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: 'var o = { get [a]() {}, set [a](foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: 'var o = { get [(a)]() {}, set [a](foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },
    {
      code: 'var o = { get [f(a)]() {}, set [f(a)](foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
    },

    // Property descriptors — complete pairs or non-descriptor calls
    "var o = {a: 1};\n Object.defineProperty(o, 'b', \n{set: function(value) {\n val = value; \n},\n get: function() {\n return val; \n} \n});",
    'var o = {set: function() {}}',
    'Object.defineProperties(obj, {set: {value: function() {}}});',
    'Object.create(null, {set: {value: function() {}}});',

    // Classes — default (no errors without options)
    {
      code: 'class A { get a() {} }',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: 'class A { get #a() {} }',
      options: [{ enforceForClassMembers: true }] as any,
    },
    {
      code: 'class A { set a(foo) {} }',
      options: [{ enforceForClassMembers: false }] as any,
    },
    {
      code: 'class A { get a() {} set a(foo) {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
    },
    {
      code: 'class A { static get a() {} static set a(foo) {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
    },

    // TS — default off
    'interface I { get prop(): any }',
    'type T = { set prop(value: any): void }',
    {
      code: 'interface I { get prop(): any, set prop(value: any): void }',
      options: [{ enforceForTSTypes: true }] as any,
    },
    {
      code: 'type T = { get prop(): any, set prop(value: any): void }',
      options: [{ enforceForTSTypes: true }] as any,
    },
    { code: 'interface I {}', options: [{ enforceForTSTypes: true }] as any },
    {
      code: 'interface I { method(): any }',
      options: [{ enforceForTSTypes: true }] as any,
    },
  ],
  invalid: [
    // Default — setter without getter
    {
      code: 'var o = { set a(value) {} };',
      errors: [{ messageId: 'missingGetterInObjectLiteral' }],
    },
    {
      code: 'var o = { get a() {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
      errors: [{ messageId: 'missingSetterInObjectLiteral' }],
    },

    // Different keys
    {
      code: 'var o = { get a() {}, set b(foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
      errors: [
        { messageId: 'missingSetterInObjectLiteral' },
        { messageId: 'missingGetterInObjectLiteral' },
      ],
    },

    // Computed different keys
    {
      code: 'var o = { get [a]() {}, set [b](foo) {} };',
      options: [{ setWithoutGet: true, getWithoutSet: true }] as any,
      errors: [
        { messageId: 'missingSetterInObjectLiteral' },
        { messageId: 'missingGetterInObjectLiteral' },
      ],
    },

    // Property descriptors
    {
      code: "var o = {d: 1};\n Object.defineProperty(o, 'c', \n{set: function(value) {\n val = value; \n} \n});",
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },
    {
      code: "Reflect.defineProperty(obj, 'foo', {set: function(value) {}});",
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },
    {
      code: 'Object.defineProperties(obj, {foo: {set: function(value) {}}});',
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },
    {
      code: 'Object.create(null, {foo: {set: function(value) {}}});',
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },

    // Optional chaining on descriptor-creating calls
    {
      code: "Object?.defineProperty(obj, 'foo', {set: function(value) {}});",
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },
    {
      code: "(Object?.defineProperty)(obj, 'foo', {set: function(value) {}});",
      errors: [{ messageId: 'missingGetterInPropertyDescriptor' }],
    },

    // Class — default (setWithoutGet only)
    {
      code: 'class A { set a(foo) {} }',
      errors: [{ messageId: 'missingGetterInClass' }],
    },
    {
      code: 'class A { get a() {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
      errors: [{ messageId: 'missingSetterInClass' }],
    },
    {
      code: 'class A { static set a(foo) {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
      errors: [{ messageId: 'missingGetterInClass' }],
    },

    // Class — different keys
    {
      code: 'class A { get a() {} set b(foo) {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
      errors: [
        { messageId: 'missingSetterInClass' },
        { messageId: 'missingGetterInClass' },
      ],
    },

    // Class — same key on static vs instance is not a pair
    {
      code: 'class A { get a() {} static set a(foo) {} }',
      options: [
        {
          setWithoutGet: true,
          getWithoutSet: true,
          enforceForClassMembers: true,
        },
      ] as any,
      errors: [
        { messageId: 'missingSetterInClass' },
        { messageId: 'missingGetterInClass' },
      ],
    },

    // Class — private identifier
    {
      code: 'class A { set #a(foo) {} }',
      errors: [{ messageId: 'missingGetterInClass' }],
    },

    // TS types
    {
      code: 'interface I { set prop(value: any): any }',
      options: [{ enforceForTSTypes: true }] as any,
      errors: [{ messageId: 'missingGetterInType' }],
    },
    {
      code: 'type T = { set prop(value: any): any }',
      options: [{ enforceForTSTypes: true }] as any,
      errors: [{ messageId: 'missingGetterInType' }],
    },
    {
      code: 'type T = { get prop(): any }',
      options: [{ enforceForTSTypes: true, getWithoutSet: true }] as any,
      errors: [{ messageId: 'missingSetterInType' }],
    },
  ],
});
