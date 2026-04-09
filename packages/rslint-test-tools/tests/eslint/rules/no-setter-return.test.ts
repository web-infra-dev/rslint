import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-setter-return', {
  valid: [
    // not a setter
    'function foo() { return 1; }',
    'function set(val: any) { return 1; }',
    'var foo = function() { return 1; };',
    'var set = (val: any) => 1;',

    // setters do not affect other functions
    '({ set a(val) { }}); function foo() { return 1; }',
    '({ set a(val) { }}); (() => 1);',

    // return without a value is allowed
    '({ set foo(val) { return; } })',
    '({ set foo(val) { if (val) { return; } } })',
    'class A { set foo(val) { return; } }',
    '(class { set foo(val) { if (val) { return; } else { return; } return; } })',

    // not a setter
    '({ get foo() { return 1; } })',
    '({ get set() { return 1; } })',
    '({ set(val: any) { return 1; } })',
    '({ set: function(val: any) { return 1; } })',
    '({ set: (val: any) => 1 })',
    'class A { get foo() { return 1; } }',
    'class A { set(val: any) { return 1; } }',
    'class A { static set(val: any) { return 1; } }',

    // not returning from the setter
    '({ set foo(val) { function foo(val: any) { return 1; } } })',
    '({ set foo(val) { var foo = (val: any) => 1; } })',
    '(class { set foo(val) { var foo = (val: any) => 1; } })',
    // default parameter with return
    '({ set foo(val = function() { return 1; }) {} })',
    '(class { set foo(val = (v: any) => 1) {} })',
    // computed property key containing return
    '({ set [function() { return 1; } as any](val) {} })',

    // property descriptors: return without value is allowed
    "Object.defineProperty(foo, 'bar', { set(val) { return; } })",
    "Reflect.defineProperty(foo, 'bar', { set(val) { if (val) { return; } } })",
    'Object.defineProperties(foo, { bar: { set(val) { try { return; } catch(e){} } } })',
    'Object.create(foo, { bar: { set: function(val: any) { return; } } })',

    // property descriptors: not a setter
    'var x = { set(val: any) { return 1; } }',
    "Object.defineProperty(foo, 'bar', { value(val: any) { return 1; } })",
    "Reflect.defineProperty(foo, 'bar', { get(val: any) { return 1; } })",
    // computed property name is variable, not string "set"
    'declare var set: any; Object.defineProperties(foo, { bar: { [set](val: any) { return 1; } } })',
    // wrong structure for Object.create
    'Object.create(foo, { set: function(val: any) { return 1; } } as any)',
    // wrong arg count for Object.defineProperty
    'Object.defineProperty(foo, { set: (val: any) => 1 } as any)',
    // reversed arg positions
    'Object.defineProperties({ bar: { set(val) { return 1; } } }, foo)',
    'Object.create({ bar: { set(val) { return 1; } } }, foo)',

    // property descriptors: not returning from the setter
    "Object.defineProperty(foo, 'bar', { set(val) { function foo() { return 1; } } })",

    // property descriptors: wrong argument index
    "Object.defineProperty(foo, 'bar', 'baz' as any, { set(val) { return 1; } })",

    // wrong object name
    'declare var object: any; object.defineProperty(foo, "bar", { set(val) { return 1; } })',

    // global object is shadowed
    'function f(Object: any) { Object.defineProperties(foo, { bar: { set(val) { try { return 1; } catch(e){} } } }) }',

    // --- nesting / composition edge cases ---
    // class inside setter — inner method return is fine
    '({ set a(val) { class Inner { method() { return 1; } } } })',
    // generator/async function inside setter
    '({ set a(val) { function* gen() { return 1; } } })',
    '({ set a(val) { async function f() { return 1; } } })',
    // map callback inside setter
    '({ set a(val) { [1,2,3].map(function(x: any) { return x * 2; }); } })',
    // nested non-descriptor "set" method
    'var x = { outer: { set(val: any) { return 1; } } }',
    // Object.defineProperty inside setter — inner getter return is fine
    "({ set a(val) { Object.defineProperty(foo, 'bar', { get() { return 1; } }); } })",
  ],
  invalid: [
    // object literal setter
    {
      code: '({ set a(val) { return 1; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // class setter
    {
      code: 'class A { set a(val) { return 1; } }',
      errors: [{ messageId: 'returnsValue' }],
    },
    // static class setter
    {
      code: 'class A { static set a(val) { return 1; } }',
      errors: [{ messageId: 'returnsValue' }],
    },
    // anonymous class setter
    {
      code: '(class { set a(val) { return 1; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // return undefined explicitly
    {
      code: 'class A { set a(val) { return undefined; } }',
      errors: [{ messageId: 'returnsValue' }],
    },
    // conditional return
    {
      code: '({ set a(val) { if (foo) { return 1; }; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // multiple invalid in same object
    {
      code: '({ set a(val) { return 1; }, set b(val) { return 1; } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // multiple invalid in same setter
    {
      code: '({ set a(val) { if(val) { return 1; } else { return 2 }; } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // mixed valid and invalid
    {
      code: '({ set a(val) { if(val) { return 1; } else { return; }; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // while loop
    {
      code: '(class { set a(val) { while (foo){ if (bar) break; else return 1; } } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // static setter with multiple returns
    {
      code: '(class { static set a(val) { if (val > 0) { (this as any)._val = val; return val; } return false; } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // switch with mixed valid/invalid
    {
      code: 'class A { set a(val) { switch(val) { case 1: return x; case 2: return; default: return z } } }',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // inner function does not affect
    {
      code: '(class { set a(val) { function b(){ return 1; } return 2; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // inner function with bare return, outer with value
    {
      code: '({ set a(val) { function b(){ return; } return 1; } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // Object.defineProperty
    {
      code: "Object.defineProperty(foo, 'bar', { set(val) { return 1; } })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // Reflect.defineProperty
    {
      code: "Reflect.defineProperty(foo, 'bar', { set(val) { return 1; } })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // Object.defineProperties
    {
      code: 'Object.defineProperties(foo, { baz: { set(val) { return 1; } } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // Object.create
    {
      code: 'Object.create(null, { baz: { set(val) { return 1; } } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // arrow implicit return in descriptor
    {
      code: "Object.defineProperty(foo, 'bar', { set: (val: any) => val })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // Object.create with 3 returns, 2 errors
    {
      code: 'Object.create(null, { baz: { set(val) { return (this as any)._val; return; return undefined; } } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // multiple descriptors with arrow
    {
      code: 'Object.create({}, { baz: { set(val) { return 1; } }, bar: { set: (val: any) => 1 } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // bracket notation
    {
      code: "Object['defineProperty'](foo, 'bar', { set: function bar(val: any) { return 1; } })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // element access with template literal
    {
      code: "Object[`defineProperties`](foo, { baz: { ['set'](val) { return 1; } } })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // optional chaining
    {
      code: "Object?.defineProperty(foo, 'bar', { set(val) { return 1; } })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // parenthesized optional chaining
    {
      code: "(Object?.defineProperty)(foo, 'bar', { set(val) { return 1; } })",
      errors: [{ messageId: 'returnsValue' }],
    },

    // --- nesting / composition edge cases ---
    // setter inside setter — both flagged
    {
      code: '({ set a(val) { return { set b(v) { return 1; } }; } })',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // deeply nested return in setter
    {
      code: '({ set a(val) { if (x) { if (y) { if (z) { return 1; } } } } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // try/catch with returns in setter
    {
      code: 'class A { set a(val) { try { return 1; } catch(e) { return 2; } finally { } } }',
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // type assertion / non-null assertion — still value returns
    {
      code: 'class A { set a(val) { return val as any; } }',
      errors: [{ messageId: 'returnsValue' }],
    },
    {
      code: 'class A { set a(val) { return val!; } }',
      errors: [{ messageId: 'returnsValue' }],
    },
    // arrow setter with ternary expression body
    {
      code: "Object.defineProperty(foo, 'bar', { set: (val: any) => val > 0 ? val : -val })",
      errors: [{ messageId: 'returnsValue' }],
    },
    // nested Object.defineProperty — both setter returns flagged
    {
      code: "Object.defineProperty(foo, 'a', { set(val) { Object.defineProperty(bar, 'b', { set(v) { return 1; } }); return 2; } })",
      errors: [{ messageId: 'returnsValue' }, { messageId: 'returnsValue' }],
    },
    // setter returning map result (outer flagged, inner callback NOT)
    {
      code: '({ set a(val) { return [1,2,3].map(function(x: any) { return x * 2; }); } })',
      errors: [{ messageId: 'returnsValue' }],
    },
    // for loop inside setter
    {
      code: '({ set a(val) { for (var i = 0; i < 10; i++) { return i; } } })',
      errors: [{ messageId: 'returnsValue' }],
    },
  ],
});
