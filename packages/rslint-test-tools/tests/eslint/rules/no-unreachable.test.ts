import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unreachable', {
  valid: [
    // Function declaration after return is hoisted
    'function foo() { return bar(); function bar() { return 1; } }',
    // Normal code before return
    'function foo() { var x = 1; return x; }',
    // if without else: not fully terminal
    'function foo() { if (x) { return; } bar(); }',
    // var without initializer after return is hoisted
    'function foo() { return; var x; }',
    // Empty statement after return is allowed
    'function foo() { return; ; }',
    // Multiple var declarations without initializers
    'function foo() { return; var x, y, z; }',
    // Function declaration after throw
    'function foo() { throw new Error(); function bar() {} }',
    // Switch without default is not terminal
    'function foo() { switch(x) { case 1: return; } bar(); }',
    // Try with finally (no catch) - try returns
    'function foo() { try { return; } finally { cleanup(); } }',
  ],
  invalid: [
    // Unreachable after return
    {
      code: 'function foo() { return; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after throw
    {
      code: 'function foo() { throw error; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after break
    {
      code: 'while (true) { break; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after continue
    {
      code: 'while (true) { continue; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // var with initializer after return IS reported
    {
      code: 'function foo() { return; var x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Multiple unreachable statements - grouped into one report
    {
      code: 'function foo() { return; x = 1; y = 2; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // let after return
    {
      code: 'function foo() { return; let x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // const after return
    {
      code: 'function foo() { return; const x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Class declaration after return
    {
      code: 'function foo() { return; class Bar {} }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after if/else where both return
    {
      code: 'function foo() { if (x) { return 1; } else { return 2; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after try/catch where both return
    {
      code: 'function foo() { try { return 1; } catch(e) { return 2; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after while(true) infinite loop
    {
      code: 'function foo() { while(true) { doSomething(); } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after for(;;) infinite loop
    {
      code: 'function foo() { for(;;) { doSomething(); } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after try/finally where finally returns
    {
      code: "function foo() { try { console.log('x'); } finally { return 1; } bar(); }",
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after labeled break
    {
      code: 'outer: while(true) { while(true) { break outer; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after switch with default where all cases return
    {
      code: 'function foo() { switch(x) { case 1: return 1; default: return 2; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after nested if/else all returning
    {
      code: 'function foo() { if (a) { if (b) { return; } else { return; } } else { return; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
  ],
});
