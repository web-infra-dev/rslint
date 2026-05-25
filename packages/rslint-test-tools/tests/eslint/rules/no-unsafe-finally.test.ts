import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-finally', {
  valid: [
    // ====================================================================
    // No control flow in finally
    // ====================================================================
    'var foo = function() { try { return 1; } catch(err) { return 2; } finally { console.log("hola!") } }',

    // ====================================================================
    // Function-like boundaries: all 7 kinds stop return/throw propagation
    // ====================================================================
    // FunctionDeclaration
    'var foo = function() { try {} finally { function a(x) { return x } } }',
    // FunctionExpression
    'var foo = function() { try {} finally { var a = function(x) { return x } } }',
    // ArrowFunction
    'var foo = function() { try {} finally { var a = (x) => { return x } } }',
    // MethodDeclaration (object literal)
    'var foo = function() { try {} finally { var obj = { method() { return 1 } } } }',
    // GetAccessor
    'var foo = function() { try {} finally { var obj = { get x() { return 1 } } } }',
    // SetAccessor
    'var foo = function() { try {} finally { var obj = { set x(v) { return } } } }',
    // Constructor
    'var foo = function() { try {} finally { class C { constructor() { return } } } }',

    // Special function variants
    'var foo = function() { try {} finally { async function a() { return 1 } } }',
    'var foo = function() { try {} finally { function* gen() { return 1 } } }',
    'var foo = function() { try {} finally { async function* gen() { return 1 } } }',
    'var foo = function() { try {} finally { var a = async () => { return 1 } } }',

    // Throw inside nested function
    'var foo = function() { try {} finally { function a() { throw new Error() } } }',
    // Complex control flow inside nested function (from ESLint original tests)
    'var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { if(!x) { throw new Error() } } } }',
    'var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { while(true) { if(x) { break } else { continue } } } } }',
    'var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { label: while(true) { if(x) { break label; } else { continue } } } } }',
    // Arrow expression body (no control flow statement)
    'var foo = function() { try { return 1; } catch(err) { return 2 } finally { (x) => x } }',

    // ====================================================================
    // Class-like boundaries: ClassDeclaration and ClassExpression
    // ====================================================================
    'var foo = function() { try {} finally { class C { method() { return 1 } } } }',
    'var foo = function() { try {} finally { var C = class { method() { return 1 } } } }',
    'var foo = function() { try {} finally { class C { static fail() { throw new Error() } } } }',
    // Multi-level: arrow function inside class method inside finally
    'var foo = function() { try {} finally { class C { method() { var fn = () => { return 1 } } } } }',

    // ====================================================================
    // Loop boundaries for unlabeled break: all 5 kinds
    // ====================================================================
    'var foo = function() { try {} finally { while (true) break; } }',
    'var foo = function() { try {} finally { for (var i = 0; i < 10; i++) break; } }',
    'var foo = function() { try {} finally { for (var x in obj) break; } }',
    'var foo = function() { try {} finally { for (var x of arr) break; } }',
    'var foo = function() { try {} finally { do { break; } while (true); } }',

    // ====================================================================
    // Loop boundaries for unlabeled continue: all 5 kinds
    // ====================================================================
    'var foo = function() { try {} finally { while (true) continue; } }',
    'var foo = function() { try {} finally { for (var i = 0; i < 10; i++) continue; } }',
    'var foo = function() { try {} finally { for (var x in obj) continue; } }',
    'var foo = function() { try {} finally { for (var x of arr) continue; } }',
    'var foo = function() { try {} finally { do { continue; } while (true); } }',

    // ====================================================================
    // Switch boundary for unlabeled break
    // ====================================================================
    'var foo = function() { try {} finally { switch (true) { case true: break; } } }',

    // ====================================================================
    // Labeled break/continue with label INSIDE finally
    // ====================================================================
    'var foo = function() { try {} finally { label: while (true) { break label; } } }',
    'var foo = function() { try {} finally { label: while (true) { continue label; } } }',
    'var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { break label; } } }',
    'var foo = function() { try {} finally { label: for (var i = 0; i < 10; i++) { continue label; } } }',
    'var foo = function() { try {} finally { label: for (var x in obj) { break label; } } }',
    'var foo = function() { try {} finally { label: for (var x of arr) { continue label; } } }',
    'var foo = function() { try {} finally { label: do { break label; } while (true); } }',
    // Label on plain block
    'var foo = function() { try {} finally { label: { break label; } } }',

    // ====================================================================
    // Labeled continue with intermediate loop (all 5 loop types)
    // ====================================================================
    'label: while (true) { try {} finally { while (true) { continue label; } } }',
    'label: while (true) { try {} finally { for (var i = 0; i < 10; i++) { continue label; } } }',
    'label: while (true) { try {} finally { for (var x in obj) { continue label; } } }',
    'label: while (true) { try {} finally { for (var x of arr) { continue label; } } }',
    'label: while (true) { try {} finally { do { continue label; } while (true); } }',

    // ====================================================================
    // No finally block at all / control flow in try-catch only
    // ====================================================================
    'var foo = function() { try { return 1 } catch(err) { return 2 } }',
    'var foo = function() { try { throw new Error() } catch(err) { return 2 } finally { console.log("done") } }',
    'var foo = function() { try { return 1 } catch(err) { throw new Error() } finally { console.log("done") } }',

    // ====================================================================
    // Deep nesting — safe due to boundaries
    // ====================================================================
    'var foo = function() { try {} finally { if (true) { var fn = () => { return 1 }; } } }',
    'var foo = function() { try {} finally { while (true) { (function() { return 1 })(); break; } } }',
    'var foo = function() { try {} finally { try { while (true) { break; } } catch(e) {} } }',
    'var foo = function() { try {} finally { (function() { class C { method() { return (() => { return 1 })() } } })() } }',

    // ====================================================================
    // Nested try-finally — safe
    // ====================================================================
    'var foo = function() { try {} finally { try { console.log(1) } finally { console.log(2) } } }',
    'var foo = function() { try {} finally { var fn = function() { try { return 1 } finally { console.log(2) } } } }',
  ],
  invalid: [
    // ====================================================================
    // Basic: return in finally
    // ====================================================================
    {
      code: 'var foo = function() { try { return 1; } catch(err) { return 2; } finally { return 3; } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try { return 1; } finally { return 3; } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Basic: throw in finally
    // ====================================================================
    {
      code: 'var foo = function() { try { return 1 } catch(err) { return 2 } finally { throw new Error() } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { throw new Error() } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Unlabeled break/continue — targeting loop OUTSIDE finally
    // ====================================================================
    {
      code: 'while (true) try {} finally { break; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'while (true) try {} finally { continue; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'for (;;) try {} finally { break; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'for (var x in obj) try {} finally { continue; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'for (var x of arr) try {} finally { break; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'do { try {} finally { break; } } while (true)',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Labeled break — label OUTSIDE finally
    // ====================================================================
    {
      code: 'label: try { return 0; } finally { break label; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'label: try {} finally { break label; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    // Inner label exists but targets outer
    {
      code: 'outer: { try {} finally { inner: { break outer; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Labeled continue — label OUTSIDE finally (no intermediate loop)
    // ====================================================================
    {
      code: 'label: while (true) try {} finally { continue label; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'label: for (;;) try {} finally { continue label; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Continue passes through switch (switch is NOT sentinel for continue)
    // ====================================================================
    {
      code: 'while (true) try {} finally { switch (true) { case true: continue; } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Control flow inside non-boundary constructs in finally
    // ====================================================================
    {
      code: 'var foo = function() { try {} finally { if (true) { return 1; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { if (true) { return 1; } else { return 2; } } }',
      errors: [{ messageId: 'unsafeUsage' }, { messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { if (cond) throw new Error() } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { { { return 1; } } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Control flow in try/catch body INSIDE outer finally
    // ====================================================================
    {
      code: 'var foo = function() { try {} finally { try { return 1; } catch(e) {} } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { try { x } catch(e) { return 1; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { try { x } catch(e) { throw e; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Nested try-finally
    // ====================================================================
    {
      code: 'var foo = function() { try {} finally { try {} finally { return 1; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { try {} finally { throw new Error(); } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    // Triple nesting
    {
      code: 'var foo = function() { try {} finally { try {} finally { try {} finally { return 1; } } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Multiple unsafe statements
    // ====================================================================
    {
      code: 'var foo = function() { try {} finally { return 1; return 2; } }',
      errors: [{ messageId: 'unsafeUsage' }, { messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { return 1; throw new Error(); } }',
      errors: [{ messageId: 'unsafeUsage' }, { messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // Complex nesting — loops/switch/if don't stop return/throw
    // ====================================================================
    {
      code: 'var foo = function() { try {} finally { for (var i = 0; i < 1; i++) { return 1; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { switch (x) { case 1: throw new Error(); } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try {} finally { if (true) { for (var i = 0; i < 1; i++) { switch (x) { default: return 1; } } } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // From ESLint original tests — return value contains function/object
    // ====================================================================
    {
      code: 'var foo = function() { try { return 1 } catch(err) { return 2 } finally { return function(x) { return y } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'var foo = function() { try { return 1 } catch(err) { return 2 } finally { return { x: function(c) { return c } } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },

    // ====================================================================
    // From ESLint original tests — break/continue across switch boundaries
    // ====================================================================
    // Unlabeled break in finally inside switch case (switch is OUTSIDE finally)
    {
      code: 'var foo = function() { switch (true) { case true: try {} finally { break; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    // Labeled break from switch case inside finally to outer label
    {
      code: 'var foo = function() { a: while (true) try {} finally { switch (true) { case true: break a; } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    // Labeled break across nested switches (outer switch label)
    {
      code: 'var foo = function() { a: switch (true) { case true: try {} finally { switch (true) { case true: break a; } } } }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
  ],
});
