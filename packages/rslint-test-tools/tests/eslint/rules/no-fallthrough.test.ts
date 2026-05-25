import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-fallthrough', {
  valid: [
    // Break exits the case
    'switch(foo) { case 0: a(); break; case 1: b(); }',
    // Empty case (no statements), allowed
    'switch(foo) { case 0: case 1: a(); break; }',
    // Comment suppresses warning (various patterns)
    'switch(foo) { case 0: a(); /* falls through */ case 1: b(); }',
    'switch(foo) { case 0: a(); /* fall through */ case 1: b(); }',
    'switch(foo) { case 0: a(); /* fallsthrough */ case 1: b(); }',
    // Return exits the case
    'function foo() { switch(bar) { case 0: a(); return; case 1: b(); } }',
    // Throw exits the case
    'switch(foo) { case 0: a(); throw e; case 1: b(); }',
    // Continue exits the case
    'while(a) { switch(foo) { case 0: a(); continue; case 1: b(); } }',
    // Case-insensitive "Fall Through"
    'switch(foo) { case 0: a(); // Fall Through\ncase 1: b(); }',
    // Last case doesn't need break
    'switch(foo) { case 0: a(); }',
    // Multiple empty cases
    'switch(foo) { case 0: case 1: case 2: a(); break; }',
    // If/else both terminate
    'switch(foo) { case 0: if (a) { break; } else { break; } case 1: b(); }',
    // Block with break
    'switch(foo) { case 0: { a(); break; } case 1: b(); }',
    // Default with break
    'switch(foo) { case 0: a(); break; default: b(); break; case 1: c(); }',
    // Try/catch both terminate
    'switch(foo) { case 0: try { break; } catch(e) { break; } case 1: b(); }',
    // Nested switch with outer break
    'switch(foo) { case 0: switch(bar) { case 1: break; } break; case 2: b(); }',
    // FALLS THROUGH (all caps)
    'switch(foo) { case 0: a(); /* FALLS THROUGH */ case 1: b(); }',
    // Try/finally where finally breaks — terminal
    'switch(foo) { case 0: try { a(); } finally { break; } case 1: b(); }',
    // Try/finally where finally returns — terminal
    'function f1() { switch(foo) { case 0: try { a(); } finally { return; } case 1: b(); } }',
    // Try/catch/finally where finally breaks
    'switch(foo) { case 0: try { a(); } catch(e) { b(); } finally { break; } case 1: c(); }',
    // Labeled break — terminal
    'switch(foo) { case 0: label1: break; case 1: b(); }',
    // If/else if/else all terminate
    'switch(foo) { case 0: if(a) { break; } else if(b) { break; } else { break; } case 1: c(); }',
    // Deeply nested if/else
    'switch(foo) { case 0: if(a) { if(b) { break; } else { break; } } else { break; } case 1: c(); }',
    // Switch with only default
    'switch(foo) { default: a(); }',
    // All cases have breaks
    'switch(foo) { case 0: a(); break; case 1: b(); break; default: c(); break; }',
    // Multi-line comment between cases
    'switch(foo) { case 0: a();\n/* This falls through intentionally */\ncase 1: b(); }',
    // Custom commentPattern option
    {
      code: 'switch(foo) { case 0: a();\n/* break omitted */\ncase 1: b(); }',
      options: { commentPattern: 'break[\\s\\w]*omitted' },
    },
    // allowEmptyCase with empty statement
    {
      code: 'switch(foo) { case 0: ; case 1: a(); break; }',
      options: { allowEmptyCase: true },
    },
    // Infinite loops — terminal
    'switch(foo) { case 0: while(true) {} case 1: b(); }',
    'switch(foo) { case 0: while("x") {} case 1: b(); }',
    // Infinite loop: for(;;) {} — terminal
    'switch(foo) { case 0: for(;;) {} case 1: b(); }',
    // Infinite loop: do {} while(true) — terminal
    'switch(foo) { case 0: do {} while(true); case 1: b(); }',
    // Nested try/finally — inner finally break swallows exceptions, outer catch unreachable
    'switch(foo) { case 0: try { try { a(); } finally { break; } } catch(e) { b(); } case 1: c(); }',
    // Try with only break in try block — catch is unreachable
    'switch(foo) { case 0: try { break; } catch(e) { a(); } case 1: b(); }',
    // Try with only continue in try block — catch is unreachable
    'while(a) { switch(foo) { case 0: try { continue; } catch(e) { a(); } case 1: b(); } }',
    // Try with bare return — catch is unreachable
    'function f2() { switch(foo) { case 0: try { return; } catch(e) { a(); } case 1: b(); } }',
    // while(true) with break in nested switch — still infinite (break captured)
    'switch(foo) { case 0: while(true) { switch(bar) { case 1: break; } } case 1: b(); }',
    // while(true) with continue — still infinite
    'switch(foo) { case 0: while(true) { continue; } case 1: b(); }',
    // while(true) with break in nested for — still infinite (break captured by for)
    'switch(foo) { case 0: while(true) { for(var j=0;j<1;j++) { break; } } case 1: b(); }',
    // Try/catch where catch has throw — both terminate
    'switch(foo) { case 0: try { break; } catch(e) { throw e; } case 1: b(); }',
    // Triple nested blocks with break
    'switch(foo) { case 0: { { { break; } } } case 1: b(); }',
    // Labeled break targeting label wrapping the switch — terminal
    'outer3: switch(foo) { case 0: a(); break outer3; case 1: b(); }',
    // Inner switch with all branches terminal (return/throw + default)
    'function f3() { switch(a) { case 0: switch(b) { case 1: return; default: throw e; }\ndefault: throw e; } }',
    // return followed by unreachable code — still terminal
    'function f6() { switch(a) { case 0: return; a(); case 1: x(); } }',
    // break followed by unreachable code — still terminal
    'switch(a) { case 0: break; a(); case 1: x(); }',
    // unreachable code inside block after return
    'function f7() { switch(a) { case 0: if(x) { return; a(); } else { return; } case 1: x(); } }',
    // while(true) with labeled break to inner block — loop is still infinite
    'switch(a) { case 0: while(true) { inner: { break inner; } } case 1: b(); }',
    // Inner switch with breaks + outer break — terminal due to outer break
    'switch(a) { case 0: switch(b) { case 1: break; default: break; } break; case 1: x(); }',
  ],
  invalid: [
    // Fallthrough from case to case
    {
      code: 'switch(foo) { case 0: a(); case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Fallthrough from case to default
    {
      code: 'switch(foo) { case 0: a(); default: b(); }',
      errors: [{ messageId: 'default' }],
    },
    // Multiline fallthrough
    {
      code: 'switch(foo) { case 0:\n  a();\ncase 1:\n  b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Nested switch: inner break does NOT prevent outer fallthrough
    {
      code: 'switch(foo) { case 0: switch(bar) { case 1: break; } case 2: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // If without else: not terminal
    {
      code: 'switch(foo) { case 0: if (a) { break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // If/else where only one terminates
    {
      code: 'switch(foo) { case 0: if (a) { break; } else { c(); } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Empty block is not terminal
    {
      code: 'switch(foo) { case 0: { } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Default in middle falls through
    {
      code: 'switch(foo) { case 0: break; default: a(); case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Comment not matching pattern
    {
      code: 'switch(foo) { case 0: a();\n/* intentional */\ncase 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // If/else if without else — not terminal
    {
      code: 'switch(foo) { case 0: if(a) { break; } else if(b) { break; } case 1: c(); }',
      errors: [{ messageId: 'case' }],
    },
    // Labeled statement wrapping non-terminal
    {
      code: 'switch(foo) { case 0: label1: a(); case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // For-in loop (not infinite) — not terminal
    {
      code: 'switch(foo) { case 0: for(var x in obj) { break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // For-of loop (not infinite) — not terminal
    {
      code: 'switch(foo) { case 0: for(var x of arr) { break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // while loop with break — NOT infinite, falls through
    {
      code: 'switch(foo) { case 0: while(a) { break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Try with expression before break — catch is reachable
    {
      code: 'switch(foo) { case 0: try { foo(); break; } catch(e) { a(); } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Try with return expr — expression might throw
    {
      code: 'function f3() { switch(foo) { case 0: try { return foo(); } catch(e) { a(); } case 1: b(); } }',
      errors: [{ messageId: 'case' }],
    },
    // try/finally where finally does NOT terminate — falls through
    {
      code: 'switch(foo) { case 0: try { a(); } finally { b(); } case 1: c(); }',
      errors: [{ messageId: 'case' }],
    },
    // while(true) with conditional break — NOT infinite
    {
      code: 'switch(foo) { case 0: while(true) { if(x) break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // labeled break targeting label INSIDE switch — NOT terminal for switch
    {
      code: 'switch(foo) { case 0: inner1: { break inner1; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Nested if where inner doesn't terminate
    {
      code: 'switch(foo) { case 0: if(a) { if(b) { break; } } else { break; } case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // variable declaration — not terminal
    {
      code: 'switch(foo) { case 0: var y = 1; case 1: x(); }',
      errors: [{ messageId: 'case' }],
    },
    // Multiple consecutive fallthroughs
    {
      code: 'switch(foo) { case 0: a(); case 1: b(); case 2: c(); break; }',
      errors: [{ messageId: 'case' }, { messageId: 'case' }],
    },
    // nested labels — break outer exits both labels AND the while loop
    {
      code: 'switch(a) { case 0: outer: inner: while(true) { break outer; } case 1: x(); }',
      errors: [{ messageId: 'case' }],
    },
    // labeled break exits the while loop → not infinite, falls through
    {
      code: 'function t1() { switch(a) { case 0: label: while(true) { break label; } case 1: x(); } }',
      errors: [{ messageId: 'case' }],
    },
    // while(0x00) — zero in hex, falsy, falls through
    {
      code: 'switch(foo) { case 0: while(0x00) {} case 1: b(); }',
      errors: [{ messageId: 'case' }],
    },
    // Inner switch with trailing empty default — switch exits normally, falls through
    {
      code: 'function f5() { switch(a) { case 0: switch(b) { case 1: return; default: } case 1: x(); } }',
      errors: [{ messageId: 'case' }],
    },
    // Inner switch with all breaks, no outer break — break exits inner, falls through outer
    {
      code: 'switch(a) { case 0: switch(b) { case 1: break; default: break; } case 1: x(); }',
      errors: [{ messageId: 'case' }],
    },
    // Inner switch without default — NOT terminal
    {
      code: 'function f4() { switch(a) { case 0: switch(b) { case 1: return; case 2: return; } case 1: x(); } }',
      errors: [{ messageId: 'case' }],
    },
    // Inner switch with default but one case doesn't terminate — 2 errors
    {
      code: 'switch(a) { case 0: switch(b) { case 1: x(); default: break; } case 1: x(); }',
      errors: [{ messageId: 'default' }, { messageId: 'case' }],
    },
    // Default first, falls through to case
    {
      code: 'switch(foo) { default: a(); case 0: b(); break; }',
      errors: [{ messageId: 'case' }],
    },
    // Custom commentPattern — default pattern no longer matches
    {
      code: 'switch(foo) { case 0: a(); /* falls through */ case 1: b(); }',
      options: { commentPattern: 'break[\\s\\w]*omitted' },
      errors: [{ messageId: 'case' }],
    },
    // reportUnusedFallthroughComment — terminal case with comment
    {
      code: 'switch(foo) { case 0: a(); break; /* falls through */ case 1: b(); }',
      options: { reportUnusedFallthroughComment: true },
      errors: [{ messageId: 'unusedFallthroughComment' }],
    },
  ],
});
