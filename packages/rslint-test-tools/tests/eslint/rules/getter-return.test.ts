import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('getter-return', {
  valid: [
    // Object getters with return
    'var foo = { get bar(){return true;} };',
    // Class getters with return
    'class foo { get bar(){return true;} }',
    'class foo { get bar(){if(baz){return true;} else {return false;} } }',
    'class foo { get(){return true;} }',
    // Object.defineProperty with getter returning value
    'Object.defineProperty(foo, "bar", { get: function () {return true;}});',
    'Object.defineProperty(foo, "bar", { get: function () { ~function (){ return true; }();return true;}});',
    // Object.defineProperties
    'Object.defineProperties(foo, { bar: { get: function () {return true;}} });',
    // Reflect.defineProperty
    'Reflect.defineProperty(foo, "bar", { get: function () {return true;}});',
    // Object.create
    'Object.create(foo, { bar: { get() {return true;} } });',
    'Object.create(foo, { bar: { get: function () {return true;} } });',
    // Non-getter functions
    'var get = function(){};',
    'var foo = { bar(){} };',
    'var foo = { get: function () {} }',
    // With allowImplicit option
    {
      code: 'var foo = { get bar() {return;} };',
      options: { allowImplicit: true },
    },
    {
      code: 'var foo = { get bar(){return true;} };',
      options: { allowImplicit: true },
    },
    {
      code: 'var foo = { get bar(){if(bar) {return;} return true;} };',
      options: { allowImplicit: true },
    },
    {
      code: 'class foo { get bar(){return true;} }',
      options: { allowImplicit: true },
    },
    {
      code: 'class foo { get bar(){return;} }',
      options: { allowImplicit: true },
    },
    // Throw statements as valid exit paths
    'var foo = { get bar(){ throw new Error("not implemented"); } };',
    'class foo { get bar(){ if(baz) { throw new Error(); } return true; } }',
    'class foo { get bar(){ if(baz) { return true; } else { throw new Error(); } } }',
    'class foo { get bar(){ if(baz) { throw new Error(); } else { throw new Error(); } } }',
    // Try/catch with return
    'class foo { get bar(){ try { return 1; } catch(e) { return 2; } } }',
    'class foo { get bar(){ try { return 1; } catch(e) { throw e; } } }',
    'class foo { get bar(){ try { throw new Error(); } catch(e) { return 1; } } }',
    'class foo { get bar(){ try { return 1; } finally { } } }',
    // Switch with return
    'class foo { get bar(){ switch(x) { case 1: return 1; default: return 2; } } }',
    'class foo { get bar(){ switch(x) { case 1: return 1; case 2: return 2; default: throw new Error(); } } }',
    // Object.defineProperty with throw
    'Object.defineProperty(foo, "bar", { get: function () { throw new Error("not implemented"); }});',
  ],

  invalid: [
    // Object getters without return
    {
      code: 'var foo = { get bar() {} };',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'var foo = { get bar(){if(baz) {return true;}} };',
      errors: [{ messageId: 'expectedAlways' }],
    },
    {
      code: 'var foo = { get bar() { return; } };',
      errors: [{ messageId: 'expected' }],
    },
    // Class getters without return
    {
      code: 'class foo { get bar(){} }',
      errors: [{ messageId: 'expected' }],
    },
    {
      code: 'class foo { get bar(){ if (baz) { return true; }}}',
      errors: [{ messageId: 'expectedAlways' }],
    },
    // Object.defineProperty without return
    {
      code: "Object.defineProperty(foo, 'bar', { get: function (){}});",
      errors: [{ messageId: 'expected' }],
    },
    {
      code: "Object.defineProperty(foo, 'bar', { get(){} });",
      errors: [{ messageId: 'expected' }],
    },
    // Optional chaining (ES2020)
    {
      code: "Object?.defineProperty(foo, 'bar', { get: function (){} });",
      errors: [{ messageId: 'expected' }],
    },
    {
      code: "(Object?.defineProperty)(foo, 'bar', { get: function (){} });",
      errors: [{ messageId: 'expected' }],
    },
    // if-throw without else (not all paths covered)
    {
      code: 'class foo { get bar(){ if(baz) { throw new Error(); } } }',
      errors: [{ messageId: 'expected' }],
    },
    // Switch without default
    {
      code: 'class foo { get bar(){ switch(x) { case 1: return 1; } } }',
      errors: [{ messageId: 'expectedAlways' }],
    },
    // Try/catch where not all paths return
    {
      code: 'class foo { get bar(){ try { return 1; } catch(e) { } } }',
      errors: [{ messageId: 'expectedAlways' }],
    },
  ],
});
